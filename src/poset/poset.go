package poset

import (
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"sort"
	"strconv"

	"github.com/sirupsen/logrus"

	"github.com/andrecronje/lachesis/src/common"
	"github.com/andrecronje/lachesis/src/peers"
)

//Poset is a DAG of Events. It also contains methods to extract a consensus
//order of Events and map them onto a blockchain.
type Poset struct {
	Participants            *peers.Peers     //[public key] => id
	Store                   Store            //store of Events, Rounds, and Blocks
	UndeterminedEvents      []string         //[index] => hash . FIFO queue of Events whose consensus order is not yet determined
	PendingRounds           []*pendingRound  //FIFO queue of Rounds which have not attained consensus yet
	LastConsensusRound      *int             //index of last consensus round
	FirstConsensusRound     *int             //index of first consensus round (only used in tests)
	AnchorBlock             *int             //index of last block with enough signatures
	LastCommitedRoundEvents int              //number of events in round before LastConsensusRound
	SigPool                 []BlockSignature //Pool of Block signatures that need to be processed
	ConsensusTransactions   int              //number of consensus transactions
	PendingLoadedEvents     int              //number of loaded events that are not yet committed
	commitCh                chan Block       //channel for committing Blocks
	topologicalIndex        int              //counter used to order events in topological order (only local)
	superMajority           int
	trustCount              int

	ancestorCache     *common.LRU
	selfAncestorCache *common.LRU
	stronglySeeCache  *common.LRU
	roundCache        *common.LRU
	timestampCache    *common.LRU

	logger *logrus.Entry
}

//NewPoset instantiates a Poset from a list of participants, underlying
//data store and commit channel
func NewPoset(participants *peers.Peers, store Store, commitCh chan Block, logger *logrus.Entry) *Poset {
	if logger == nil {
		log := logrus.New()
		log.Level = logrus.DebugLevel
		logger = logrus.NewEntry(log)
	}

	superMajority := 2*participants.Len()/3 + 1
	trustCount := int(math.Ceil(float64(participants.Len()) / float64(3)))

	cacheSize := store.CacheSize()
	poset := Poset{
		Participants:      participants,
		Store:             store,
		commitCh:          commitCh,
		ancestorCache:     common.NewLRU(cacheSize, nil),
		selfAncestorCache: common.NewLRU(cacheSize, nil),
		stronglySeeCache:  common.NewLRU(cacheSize, nil),
		roundCache:        common.NewLRU(cacheSize, nil),
		timestampCache:    common.NewLRU(cacheSize, nil),
		logger:            logger,
		superMajority:     superMajority,
		trustCount:        trustCount,
	}

	return &poset
}

/*******************************************************************************
Private Methods
*******************************************************************************/

//true if y is an ancestor of x
func (p *Poset) ancestor(x, y string) (bool, error) {
	if c, ok := p.ancestorCache.Get(Key{x, y}); ok {
		return c.(bool), nil
	}
	a, err := p.ancestor2(x, y)
	if err != nil {
		return false, err
	}
	p.ancestorCache.Add(Key{x, y}, a)
	return a, nil
}

func (p *Poset) ancestor2(x, y string) (bool, error) {
	if x == y {
		return true, nil
	}

	ex, err := p.Store.GetEvent(x)
	if err != nil {
		return false, err
	}

	ey, err := p.Store.GetEvent(y)
	if err != nil {
		return false, err
	}

	eyCreator := p.Participants.ByPubKey[ey.Creator()].ID
	entry, ok := ex.lastAncestors.GetByID(eyCreator)

	if !ok {
		return false, errors.New("Unknown event id " + strconv.Itoa(eyCreator))
	}

	lastAncestorKnownFromYCreator := entry.event.index

	return lastAncestorKnownFromYCreator >= ey.Index(), nil
}

//true if y is a self-ancestor of x
func (p *Poset) selfAncestor(x, y string) (bool, error) {
	if c, ok := p.selfAncestorCache.Get(Key{x, y}); ok {
		return c.(bool), nil
	}
	a, err := p.selfAncestor2(x, y)
	if err != nil {
		return false, err
	}
	p.selfAncestorCache.Add(Key{x, y}, a)
	return a, nil
}

func (p *Poset) selfAncestor2(x, y string) (bool, error) {
	if x == y {
		return true, nil
	}
	ex, err := p.Store.GetEvent(x)
	if err != nil {
		return false, err
	}
	exCreator := p.Participants.ByPubKey[ex.Creator()].ID

	ey, err := p.Store.GetEvent(y)
	if err != nil {
		return false, err
	}
	eyCreator := p.Participants.ByPubKey[ey.Creator()].ID

	return exCreator == eyCreator && ex.Index() >= ey.Index(), nil
}

//true if x sees y
func (p *Poset) see(x, y string) (bool, error) {
	return p.ancestor(x, y)
	//it is not necessary to detect forks because we assume that the InsertEvent
	//function makes it impossible to insert two Events at the same height for
	//the same participant.
}

//true if x strongly sees y
func (p *Poset) stronglySee(x, y string) (bool, error) {
	if c, ok := p.stronglySeeCache.Get(Key{x, y}); ok {
		return c.(bool), nil
	}
	ss, err := p.stronglySee2(x, y)
	if err != nil {
		return false, err
	}
	p.stronglySeeCache.Add(Key{x, y}, ss)
	return ss, nil
}

func (p *Poset) stronglySee2(x, y string) (bool, error) {

	ex, err := p.Store.GetEvent(x)
	if err != nil {
		return false, err
	}

	ey, err := p.Store.GetEvent(y)
	if err != nil {
		return false, err
	}

	c := 0
	for i, entry := range ex.lastAncestors {
		if entry.event.index >= ey.firstDescendants[i].event.index {
			c++
		}
	}
	return c >= p.superMajority, nil
}

func (p *Poset) round(x string) (int, error) {
	if c, ok := p.roundCache.Get(x); ok {
		return c.(int), nil
	}
	r, err := p.round2(x)
	if err != nil {
		return -1, err
	}
	p.roundCache.Add(x, r)
	return r, nil
}

func (p *Poset) round2(x string) (int, error) {

	/*
		x is the Root
		Use Root.SelfParent.Round
	*/
	rootsBySelfParent, _ := p.Store.RootsBySelfParent()
	if r, ok := rootsBySelfParent[x]; ok {
		return r.SelfParent.Round, nil
	}

	ex, err := p.Store.GetEvent(x)
	if err != nil {
		return math.MinInt32, err
	}

	root, err := p.Store.GetRoot(ex.Creator())
	if err != nil {
		return math.MinInt32, err
	}

	/*
		The Event is directly attached to the Root.
	*/
	if ex.SelfParent() == root.SelfParent.Hash {
		//Root is authoritative EXCEPT if other-parent is not in the root
		if other, ok := root.Others[ex.Hex()]; (ex.OtherParent() == "") ||
			(ok && other.Hash == ex.OtherParent()) {

			return root.NextRound, nil
		}
	}

	/*
		The Event's parents are "normal" Events.
		Use the whitepaper formula: parentRound + roundInc
	*/
	parentRound, err := p.round(ex.SelfParent())
	if err != nil {
		return math.MinInt32, err
	}
	if ex.OtherParent() != "" {
		var opRound int
		//XXX
		if other, ok := root.Others[ex.Hex()]; ok && other.Hash == ex.OtherParent() {
			opRound = root.NextRound
		} else {
			opRound, err = p.round(ex.OtherParent())
			if err != nil {
				return math.MinInt32, err
			}
		}

		if opRound > parentRound {
			parentRound = opRound
		}
	}

	c := 0
	for _, w := range p.Store.RoundWitnesses(parentRound) {
		ss, err := p.stronglySee(x, w)
		if err != nil {
			return math.MinInt32, err
		}
		if ss {
			c++
		}
	}
	if c >= p.superMajority {
		parentRound++
	}

	return parentRound, nil
}

//true if x is a witness (first event of a round for the owner)
func (p *Poset) witness(x string) (bool, error) {
	ex, err := p.Store.GetEvent(x)
	if err != nil {
		return false, err
	}

	xRound, err := p.round(x)
	if err != nil {
		return false, err
	}
	spRound, err := p.round(ex.SelfParent())
	if err != nil {
		return false, err
	}
	return xRound > spRound, nil
}

func (p *Poset) roundReceived(x string) (int, error) {

	ex, err := p.Store.GetEvent(x)
	if err != nil {
		return -1, err
	}

	res := -1
	if ex.roundReceived != nil {
		res = *ex.roundReceived
	}

	return res, nil
}

func (p *Poset) lamportTimestamp(x string) (int, error) {
	if c, ok := p.timestampCache.Get(x); ok {
		return c.(int), nil
	}
	r, err := p.lamportTimestamp2(x)
	if err != nil {
		return -1, err
	}
	p.timestampCache.Add(x, r)
	return r, nil
}

func (p *Poset) lamportTimestamp2(x string) (int, error) {
	/*
		x is the Root
		User Root.SelfParent.LamportTimestamp
	*/
	rootsBySelfParent, _ := p.Store.RootsBySelfParent()
	if r, ok := rootsBySelfParent[x]; ok {
		return r.SelfParent.LamportTimestamp, nil
	}

	ex, err := p.Store.GetEvent(x)
	if err != nil {
		return math.MinInt32, err
	}

	//We are going to need the Root later
	root, err := p.Store.GetRoot(ex.Creator())
	if err != nil {
		return math.MinInt32, err
	}

	plt := math.MinInt32
	//If it is the creator's first Event, use the corresponding Root
	if ex.SelfParent() == root.SelfParent.Hash {
		plt = root.SelfParent.LamportTimestamp
	} else {
		t, err := p.lamportTimestamp(ex.SelfParent())
		if err != nil {
			return math.MinInt32, err
		}
		plt = t
	}

	if ex.OtherParent() != "" {
		opLT := math.MinInt32
		if _, err := p.Store.GetEvent(ex.OtherParent()); err == nil {
			//if we know the other-parent, fetch its Round directly
			t, err := p.lamportTimestamp(ex.OtherParent())
			if err != nil {
				return math.MinInt32, err
			}
			opLT = t
		} else if other, ok := root.Others[x]; ok && other.Hash == ex.OtherParent() {
			//we do not know the other-parent but it is referenced  in Root.Others
			//we use the Root's LamportTimestamp
			opLT = other.LamportTimestamp
		}

		if opLT > plt {
			plt = opLT
		}
	}

	return plt + 1, nil
}

//round(x) - round(y)
func (p *Poset) roundDiff(x, y string) (int, error) {

	xRound, err := p.round(x)
	if err != nil {
		return math.MinInt32, fmt.Errorf("event %s has negative round", x)
	}

	yRound, err := p.round(y)
	if err != nil {
		return math.MinInt32, fmt.Errorf("event %s has negative round", y)
	}

	return xRound - yRound, nil
}

//Check the SelfParent is the Creator's last known Event
func (p *Poset) checkSelfParent(event Event) error {
	selfParent := event.SelfParent()
	creator := event.Creator()

	creatorLastKnown, _, err := p.Store.LastEventFrom(creator)
	if err != nil {
		return err
	}

	selfParentLegit := selfParent == creatorLastKnown

	if !selfParentLegit {
		return fmt.Errorf("Self-parent not last known event by creator")
	}

	return nil
}

//Check if we know the OtherParent
func (p *Poset) checkOtherParent(event Event) error {
	otherParent := event.OtherParent()
	if otherParent != "" {
		//Check if we have it
		_, err := p.Store.GetEvent(otherParent)
		if err != nil {
			//it might still be in the Root
			root, err := p.Store.GetRoot(event.Creator())
			if err != nil {
				return err
			}
			other, ok := root.Others[event.Hex()]
			if ok && other.Hash == event.OtherParent() {
				return nil
			}
			return fmt.Errorf("Other-parent not known")
		}
	}
	return nil
}

//initialize arrays of last ancestors and first descendants
func (p *Poset) initEventCoordinates(event *Event) error {
	members := p.Participants.Len()

	event.firstDescendants = make(OrderedEventCoordinates, members)
	for i, id := range p.Participants.ToIDSlice() {
		event.firstDescendants[i] = Index{
			participantId: id,
			event: EventCoordinates{
				index: math.MaxInt32,
			},
		}
	}

	event.lastAncestors = make(OrderedEventCoordinates, members)

	selfParent, selfParentError := p.Store.GetEvent(event.SelfParent())
	otherParent, otherParentError := p.Store.GetEvent(event.OtherParent())

	if selfParentError != nil && otherParentError != nil {
		for i, entry := range event.firstDescendants {
			event.lastAncestors[i] = Index{
				participantId: entry.participantId,
				event: EventCoordinates{
					index: -1,
				},
			}
		}
	} else if selfParentError != nil {
		copy(event.lastAncestors[:members], otherParent.lastAncestors)
	} else if otherParentError != nil {
		copy(event.lastAncestors[:members], selfParent.lastAncestors)
	} else {
		selfParentLastAncestors := selfParent.lastAncestors
		otherParentLastAncestors := otherParent.lastAncestors

		copy(event.lastAncestors[:members], selfParentLastAncestors)
		for i := range event.lastAncestors {
			if event.lastAncestors[i].event.index < otherParentLastAncestors[i].event.index {
				event.lastAncestors[i].event.index = otherParentLastAncestors[i].event.index
				event.lastAncestors[i].event.hash = otherParentLastAncestors[i].event.hash
			}
		}
	}

	index := event.Index()

	creator := event.Creator()
	creatorPeer, ok := p.Participants.ByPubKey[creator]
	if !ok {
		return fmt.Errorf("Could not find creator id (%s)", creator)
	}
	hash := event.Hex()

	i := event.firstDescendants.GetIDIndex(creatorPeer.ID)
	j := event.lastAncestors.GetIDIndex(creatorPeer.ID)

	if i == -1 {
		return fmt.Errorf("Could not find first descendant from creator id (%d)", creatorPeer.ID)
	}

	if j == -1 {
		return fmt.Errorf("Could not find last ancestor from creator id (%d)", creatorPeer.ID)
	}

	event.firstDescendants[i].event = EventCoordinates{index: index, hash: hash}
	event.lastAncestors[j].event = EventCoordinates{index: index, hash: hash}

	return nil
}

//update first decendant of each last ancestor to point to event
func (p *Poset) updateAncestorFirstDescendant(event Event) error {
	creatorPeer, ok := p.Participants.ByPubKey[event.Creator()]
	if !ok {
		return fmt.Errorf("Could not find creator id (%s)", event.Creator())
	}
	index := event.Index()
	hash := event.Hex()

	for i := range event.lastAncestors {
		ah := event.lastAncestors[i].event.hash
		for ah != "" {
			a, err := p.Store.GetEvent(ah)
			if err != nil {
				break
			}
			idx := a.firstDescendants.GetIDIndex(creatorPeer.ID)

			if idx == -1 {
				return fmt.Errorf("Could not find first descendant by creator id (%s)", event.Creator())
			}

			if a.firstDescendants[idx].event.index == math.MaxInt32 {
				a.firstDescendants[idx].event = EventCoordinates{index: index, hash: hash}
				if err := p.Store.SetEvent(a); err != nil {
					return err
				}
				ah = a.SelfParent()
			} else {
				break
			}
		}
	}

	return nil
}

func (p *Poset) createSelfParentRootEvent(ev Event) (RootEvent, error) {
	sp := ev.SelfParent()
	spLT, err := p.lamportTimestamp(sp)
	if err != nil {
		return RootEvent{}, err
	}
	spRound, err := p.round(sp)
	if err != nil {
		return RootEvent{}, err
	}
	selfParentRootEvent := RootEvent{
		Hash:             sp,
		CreatorID:        p.Participants.ByPubKey[ev.Creator()].ID,
		Index:            ev.Index() - 1,
		LamportTimestamp: spLT,
		Round:            spRound,
		//FlagTable:ev.FlagTable,
		//flags:ev.flags,
	}
	return selfParentRootEvent, nil
}

func (p *Poset) createOtherParentRootEvent(ev Event) (RootEvent, error) {

	op := ev.OtherParent()

	//it might still be in the Root
	root, err := p.Store.GetRoot(ev.Creator())
	if err != nil {
		return RootEvent{}, err
	}
	if other, ok := root.Others[ev.Hex()]; ok && other.Hash == op {
		return other, nil
	}

	otherParent, err := p.Store.GetEvent(op)
	if err != nil {
		return RootEvent{}, err
	}
	opLT, err := p.lamportTimestamp(op)
	if err != nil {
		return RootEvent{}, err
	}
	opRound, err := p.round(op)
	if err != nil {
		return RootEvent{}, err
	}
	otherParentRootEvent := RootEvent{
		Hash:             op,
		CreatorID:        p.Participants.ByPubKey[otherParent.Creator()].ID,
		Index:            otherParent.Index(),
		LamportTimestamp: opLT,
		Round:            opRound,
	}
	return otherParentRootEvent, nil

}

func (p *Poset) createRoot(ev Event) (Root, error) {

	evRound, err := p.round(ev.Hex())
	if err != nil {
		return Root{}, err
	}

	/*
		SelfParent
	*/
	selfParentRootEvent, err := p.createSelfParentRootEvent(ev)
	if err != nil {
		return Root{}, err
	}

	/*
		OtherParent
	*/
	var otherParentRootEvent *RootEvent
	if ev.OtherParent() != "" {
		opre, err := p.createOtherParentRootEvent(ev)
		if err != nil {
			return Root{}, err
		}
		otherParentRootEvent = &opre
	}

	root := Root{
		NextRound:  evRound,
		SelfParent: selfParentRootEvent,
		Others:     map[string]RootEvent{},
	}

	if otherParentRootEvent != nil {
		root.Others[ev.Hex()] = *otherParentRootEvent
	}

	return root, nil
}

func (p *Poset) setWireInfo(event *Event) error {
	selfParentIndex := -1
	otherParentCreatorID := -1
	otherParentIndex := -1

	//could be the first Event inserted for this creator. In this case, use Root
	if lf, isRoot, _ := p.Store.LastEventFrom(event.Creator()); isRoot && lf == event.SelfParent() {
		root, err := p.Store.GetRoot(event.Creator())
		if err != nil {
			return err
		}
		selfParentIndex = root.SelfParent.Index
	} else {
		selfParent, err := p.Store.GetEvent(event.SelfParent())
		if err != nil {
			return err
		}
		selfParentIndex = selfParent.Index()
	}

	if event.OtherParent() != "" {
		//Check Root then regular Events
		root, err := p.Store.GetRoot(event.Creator())
		if err != nil {
			return err
		}
		if other, ok := root.Others[event.Hex()]; ok && other.Hash == event.OtherParent() {
			otherParentCreatorID = other.CreatorID
			otherParentIndex = other.Index
		} else {
			otherParent, err := p.Store.GetEvent(event.OtherParent())
			if err != nil {
				return err
			}
			otherParentCreatorID = p.Participants.ByPubKey[otherParent.Creator()].ID
			otherParentIndex = otherParent.Index()
		}
	}

	event.SetWireInfo(selfParentIndex,
		otherParentCreatorID,
		otherParentIndex,
		p.Participants.ByPubKey[event.Creator()].ID)

	return nil
}

func (p *Poset) updatePendingRounds(decidedRounds map[int]int) {
	for _, ur := range p.PendingRounds {
		if _, ok := decidedRounds[ur.Index]; ok {
			ur.Decided = true
		}
	}
}

//Remove processed Signatures from SigPool
func (p *Poset) removeProcessedSignatures(processedSignatures map[int]bool) {
	var newSigPool []BlockSignature
	for _, bs := range p.SigPool {
		if _, ok := processedSignatures[bs.Index]; !ok {
			newSigPool = append(newSigPool, bs)
		}
	}
	p.SigPool = newSigPool
}

/*******************************************************************************
Public Methods
*******************************************************************************/

//InsertEvent attempts to insert an Event in the DAG. It verifies the signature,
//checks the ancestors are known, and prevents the introduction of forks.
func (p *Poset) InsertEvent(event Event, setWireInfo bool) error {
	//verify signature
	if ok, err := event.Verify(); !ok {
		if err != nil {
			return err
		}
		return fmt.Errorf("Invalid Event signature")
	}

	if err := p.checkSelfParent(event); err != nil {
		return fmt.Errorf("CheckSelfParent: %s", err)
	}

	if err := p.checkOtherParent(event); err != nil {
		return fmt.Errorf("CheckOtherParent: %s", err)
	}

	event.topologicalIndex = p.topologicalIndex
	p.topologicalIndex++

	if setWireInfo {
		if err := p.setWireInfo(&event); err != nil {
			return fmt.Errorf("SetWireInfo: %s", err)
		}
	}

	if err := p.initEventCoordinates(&event); err != nil {
		return fmt.Errorf("InitEventCoordinates: %s", err)
	}

	if err := p.Store.SetEvent(event); err != nil {
		return fmt.Errorf("SetEvent: %s", err)
	}

	if err := p.updateAncestorFirstDescendant(event); err != nil {
		return fmt.Errorf("UpdateAncestorFirstDescendant: %s", err)
	}

	p.UndeterminedEvents = append(p.UndeterminedEvents, event.Hex())

	if event.IsLoaded() {
		p.PendingLoadedEvents++
	}

	p.SigPool = append(p.SigPool, event.BlockSignatures()...)

	return nil
}

/*
DivideRounds assigns a Round and LamportTimestamp to Events, and flags them as
witnesses if necessary. Pushes Rounds in the PendingRounds queue if necessary.
*/
func (p *Poset) DivideRounds() error {

	for _, hash := range p.UndeterminedEvents {

		ev, err := p.Store.GetEvent(hash)
		if err != nil {
			return err
		}

		updateEvent := false

		/*
		   Compute Event's round, update the corresponding Round object, and
		   add it to the PendingRounds queue if necessary.
		*/
		if ev.round == nil {

			roundNumber, err := p.round(hash)
			if err != nil {
				return err
			}

			ev.SetRound(roundNumber)
			updateEvent = true

			roundInfo, err := p.Store.GetRound(roundNumber)
			if err != nil && !common.Is(err, common.KeyNotFound) {
				return err
			}

			/*
				Why the lower bound?
				Normally, once a Round has attained consensus, it is impossible for
				new Events from a previous Round to be inserted; the lower bound
				appears redundant. This is the case when the poset grows
				linearly, without jumps, which is what we intend by 'Normally'.
				But the Reset function introduces a dicontinuity  by jumping
				straight to a specific place in the poset. This technique relies
				on a base layer of Events (the corresponding Frame's Events) for
				other Events to be added on top, but the base layer must not be
				reprocessed.
			*/
			if !roundInfo.queued &&
				(p.LastConsensusRound == nil ||
					roundNumber >= *p.LastConsensusRound) {

				p.PendingRounds = append(p.PendingRounds, &pendingRound{roundNumber, false})
				roundInfo.queued = true
			}

			witness, err := p.witness(hash)
			if err != nil {
				return err
			}
			roundInfo.AddEvent(hash, witness)

			err = p.Store.SetRound(roundNumber, roundInfo)
			if err != nil {
				return err
			}
		}

		/*
			Compute the Event's LamportTimestamp
		*/
		if ev.lamportTimestamp == nil {

			lamportTimestamp, err := p.lamportTimestamp(hash)
			if err != nil {
				return err
			}

			ev.SetLamportTimestamp(lamportTimestamp)
			updateEvent = true
		}

		if updateEvent {
			p.Store.SetEvent(ev)
		}
	}

	return nil
}

//DecideFame decides if witnesses are famous
func (p *Poset) DecideFame() error {

	//Initialize the vote map
	votes := make(map[string]map[string]bool) //[x][y]=>vote(x,y)
	setVote := func(votes map[string]map[string]bool, x, y string, vote bool) {
		if votes[x] == nil {
			votes[x] = make(map[string]bool)
		}
		votes[x][y] = vote
	}

	decidedRounds := map[int]int{} // [round number] => index in p.PendingRounds

	for pos, r := range p.PendingRounds {
		roundIndex := r.Index
		roundInfo, err := p.Store.GetRound(roundIndex)
		if err != nil {
			return err
		}
		for _, x := range roundInfo.Witnesses() {
			if roundInfo.IsDecided(x) {
				continue
			}
		VOTE_LOOP:
			for j := roundIndex + 1; j <= p.Store.LastRound(); j++ {
				for _, y := range p.Store.RoundWitnesses(j) {
					diff := j - roundIndex
					if diff == 1 {
						ycx, err := p.see(y, x)
						if err != nil {
							return err
						}
						setVote(votes, y, x, ycx)
					} else {
						//count votes
						var ssWitnesses []string
						for _, w := range p.Store.RoundWitnesses(j - 1) {
							ss, err := p.stronglySee(y, w)
							if err != nil {
								return err
							}
							if ss {
								ssWitnesses = append(ssWitnesses, w)
							}
						}
						yays := 0
						nays := 0
						for _, w := range ssWitnesses {
							if votes[w][x] {
								yays++
							} else {
								nays++
							}
						}
						v := false
						t := nays
						if yays >= nays {
							v = true
							t = yays
						}

						//normal round
						if math.Mod(float64(diff), float64(p.Participants.Len())) > 0 {
							if t >= p.superMajority {
								roundInfo.SetFame(x, v)
								setVote(votes, y, x, v)
								break VOTE_LOOP //break out of j loop
							} else {
								setVote(votes, y, x, v)
							}
						} else { //coin round
							if t >= p.superMajority {
								setVote(votes, y, x, v)
							} else {
								setVote(votes, y, x, middleBit(y)) //middle bit of y's hash
							}
						}
					}
				}
			}
		}

		err = p.Store.SetRound(roundIndex, roundInfo)
		if err != nil {
			return err
		}

		if roundInfo.WitnessesDecided() {
			decidedRounds[roundIndex] = pos
		}

	}

	p.updatePendingRounds(decidedRounds)
	return nil
}

//DecideRoundReceived assigns a RoundReceived to undetermined events when they
//reach consensus
func (p *Poset) DecideRoundReceived() error {

	var newUndeterminedEvents []string

	/* From whitepaper - 18/03/18
	   "[...] An event is said to be “received” in the first round where all the
	   unique famous witnesses have received it, if all earlier rounds have the
	   fame of all witnesses decided"
	*/
	for _, x := range p.UndeterminedEvents {

		received := false
		r, err := p.round(x)
		if err != nil {
			return err
		}

		for i := r + 1; i <= p.Store.LastRound(); i++ {

			tr, err := p.Store.GetRound(i)
			if err != nil {
				//Can happen after a Reset/FastSync
				if p.LastConsensusRound != nil &&
					r < *p.LastConsensusRound {
					received = true
					break
				}
				return err
			}

			//We are looping from earlier to later rounds; so if we encounter
			//one round with undecided witnesses, we are sure that this event
			//is not "received". Break out of i loop
			if !(tr.WitnessesDecided()) {
				break
			}

			fws := tr.FamousWitnesses()
			//set of famous witnesses that see x
			var s []string
			for _, w := range fws {
				see, err := p.see(w, x)
				if err != nil {
					return err
				}
				if see {
					s = append(s, w)
				}
			}

			if len(s) == len(fws) && len(s) > 0 {

				received = true

				ex, err := p.Store.GetEvent(x)
				if err != nil {
					return err
				}
				ex.SetRoundReceived(i)

				err = p.Store.SetEvent(ex)
				if err != nil {
					return err
				}

				tr.SetConsensusEvent(x)
				err = p.Store.SetRound(i, tr)
				if err != nil {
					return err
				}

				//break out of i loop
				break
			}

		}

		if !received {
			newUndeterminedEvents = append(newUndeterminedEvents, x)
		}
	}

	p.UndeterminedEvents = newUndeterminedEvents

	return nil
}

//ProcessDecidedRounds takes Rounds whose witnesses are decided, computes the
//corresponding Frames, maps them into Blocks, and commits the Blocks via the
//commit channel
func (p *Poset) ProcessDecidedRounds() error {

	//Defer removing processed Rounds from the PendingRounds Queue
	processedIndex := 0
	defer func() {
		p.PendingRounds = p.PendingRounds[processedIndex:]
	}()

	for _, r := range p.PendingRounds {

		//Although it is possible for a Round to be 'decided' before a previous
		//round, we should NEVER process a decided round before all the previous
		//rounds are processed.
		if !r.Decided {
			break
		}

		//This is similar to the lower bound introduced in DivideRounds; it is
		//redundant in normal operations, but becomes necessary after a Reset.
		//Indeed, after a Reset, LastConsensusRound is added to PendingRounds,
		//but its ConsensusEvents (which are necessarily 'under' this Round) are
		//already deemed committed. Hence, skip this Round after a Reset.
		if p.LastConsensusRound != nil && r.Index == *p.LastConsensusRound {
			continue
		}

		frame, err := p.GetFrame(r.Index)
		if err != nil {
			return fmt.Errorf("Getting Frame %d: %v", r.Index, err)
		}

		round, err := p.Store.GetRound(r.Index)
		if err != nil {
			return err
		}
		p.logger.WithFields(logrus.Fields{
			"round_received": r.Index,
			"witnesses":      round.FamousWitnesses(),
			"events":         len(frame.Events),
			"roots":          frame.Roots,
		}).Debugf("Processing Decided Round")

		if len(frame.Events) > 0 {

			for _, e := range frame.Events {
				err := p.Store.AddConsensusEvent(e)
				if err != nil {
					return err
				}
				p.ConsensusTransactions += len(e.Transactions())
				if e.IsLoaded() {
					p.PendingLoadedEvents--
				}
			}

			lastBlockIndex := p.Store.LastBlockIndex()
			block, err := NewBlockFromFrame(lastBlockIndex+1, frame)
			if err != nil {
				return err
			}
			if err := p.Store.SetBlock(block); err != nil {
				return err
			}

			if p.commitCh != nil {
				p.commitCh <- block
			}

		} else {
			p.logger.Debugf("No Events to commit for ConsensusRound %d", r.Index)
		}

		processedIndex++

		if p.LastConsensusRound == nil || r.Index > *p.LastConsensusRound {
			p.setLastConsensusRound(r.Index)
		}

	}

	return nil
}

//GetFrame computes the Frame corresponding to a RoundReceived.
func (p *Poset) GetFrame(roundReceived int) (Frame, error) {

	//Try to get it from the Store first
	frame, err := p.Store.GetFrame(roundReceived)
	if err == nil || !common.Is(err, common.KeyNotFound) {
		return frame, err
	}

	//Get the Round and corresponding consensus Events
	round, err := p.Store.GetRound(roundReceived)
	if err != nil {
		return Frame{}, err
	}

	var events []Event
	for _, eh := range round.ConsensusEvents() {
		e, err := p.Store.GetEvent(eh)
		if err != nil {
			return Frame{}, err
		}
		events = append(events, e)
	}

	sort.Sort(ByLamportTimestamp(events))

	// Get/Create Roots
	roots := make(map[string]Root)
	//The events are in topological order. Each time we run into the first Event
	//of a participant, we create a Root for it.
	for _, ev := range events {
		c := ev.Creator()
		if _, ok := roots[c]; !ok {
			root, err := p.createRoot(ev)
			if err != nil {
				return Frame{}, err
			}
			roots[ev.Creator()] = root
		}
	}

	//Every participant needs a Root in the Frame. For the participants that
	//have no Events in this Frame, we create a Root from their last consensus
	//Event, or their last known Root
	for _, peer := range p.Participants.ToPubKeySlice() {
		if _, ok := roots[peer]; !ok {
			var root Root
			lastConsensusEventHash, isRoot, err := p.Store.LastConsensusEventFrom(peer)
			if err != nil {
				return Frame{}, err
			}
			if isRoot {
				root, _ = p.Store.GetRoot(peer)
			} else {
				lastConsensusEvent, err := p.Store.GetEvent(lastConsensusEventHash)
				if err != nil {
					return Frame{}, err
				}
				root, err = p.createRoot(lastConsensusEvent)
				if err != nil {
					return Frame{}, err
				}
			}
			roots[peer] = root
		}
	}

	//Some Events in the Frame might have other-parents that are outside of the
	//Frame (cf root.go ex 2)
	//When inserting these Events in a newly reset poset, the CheckOtherParent
	//method would return an error because the other-parent would not be found.
	//So we make it possible to also look for other-parents in the creator's Root.
	treated := map[string]bool{}
	for _, ev := range events {
		treated[ev.Hex()] = true
		otherParent := ev.OtherParent()
		if otherParent != "" {
			opt, ok := treated[otherParent]
			if !opt || !ok {
				if ev.SelfParent() != roots[ev.Creator()].SelfParent.Hash {
					other, err := p.createOtherParentRootEvent(ev)
					if err != nil {
						return Frame{}, err
					}
					roots[ev.Creator()].Others[ev.Hex()] = other
				}
			}
		}
	}

	//order roots
	orderedRoots := make([]Root, p.Participants.Len())
	for i, peer := range p.Participants.ToPeerSlice() {
		orderedRoots[i] = roots[peer.PubKeyHex]
	}

	res := Frame{
		Round:  roundReceived,
		Roots:  orderedRoots,
		Events: events,
	}

	if err := p.Store.SetFrame(res); err != nil {
		return Frame{}, err
	}

	return res, nil
}

//ProcessSigPool runs through the SignaturePool and tries to map a Signature to
//a known Block. If a Signature is found to be valid for a known Block, it is
//appended to the block and removed from the SignaturePool
func (p *Poset) ProcessSigPool() error {
	processedSignatures := map[int]bool{} //index in SigPool => Processed?
	defer p.removeProcessedSignatures(processedSignatures)

	for i, bs := range p.SigPool {
		//check if validator belongs to list of participants
		validatorHex := fmt.Sprintf("0x%X", bs.Validator)
		if _, ok := p.Participants.ByPubKey[validatorHex]; !ok {
			p.logger.WithFields(logrus.Fields{
				"index":     bs.Index,
				"validator": validatorHex,
			}).Warning("Verifying Block signature. Unknown validator")
			continue
		}
		//only check if bs is greater than AnchorBlock, otherwise simply remove
		if p.AnchorBlock == nil ||
			bs.Index > *p.AnchorBlock {
			block, err := p.Store.GetBlock(bs.Index)
			if err != nil {
				p.logger.WithFields(logrus.Fields{
					"index": bs.Index,
					"msg":   err,
				}).Warning("Verifying Block signature. Could not fetch Block")
				continue
			}
			valid, err := block.Verify(bs)
			if err != nil {
				p.logger.WithFields(logrus.Fields{
					"index": bs.Index,
					"msg":   err,
				}).Error("Verifying Block signature")
				return err
			}
			if !valid {
				p.logger.WithFields(logrus.Fields{
					"index":     bs.Index,
					"validator": p.Participants.ByPubKey[validatorHex],
					"block":     block,
				}).Warning("Verifying Block signature. Invalid signature")
				continue
			}

			block.SetSignature(bs)

			if err := p.Store.SetBlock(block); err != nil {
				p.logger.WithFields(logrus.Fields{
					"index": bs.Index,
					"msg":   err,
				}).Warning("Saving Block")
			}

			if len(block.Signatures) > p.trustCount &&
				(p.AnchorBlock == nil ||
					block.Index() > *p.AnchorBlock) {
				p.setAnchorBlock(block.Index())
				p.logger.WithFields(logrus.Fields{
					"block_index": block.Index(),
					"signatures":  len(block.Signatures),
					"trustCount":  p.trustCount,
				}).Debug("Setting AnchorBlock")
			}
		}

		processedSignatures[i] = true
	}

	return nil
}

//GetAnchorBlockWithFrame returns the AnchorBlock and the corresponding Frame.
//This can be used as a base to Reset a Poset
func (p *Poset) GetAnchorBlockWithFrame() (Block, Frame, error) {

	if p.AnchorBlock == nil {
		return Block{}, Frame{}, fmt.Errorf("No Anchor Block")
	}

	block, err := p.Store.GetBlock(*p.AnchorBlock)
	if err != nil {
		return Block{}, Frame{}, err
	}

	frame, err := p.GetFrame(block.RoundReceived())
	if err != nil {
		return Block{}, Frame{}, err
	}

	return block, frame, nil
}

//Reset clears the Poset and resets it from a new base.
func (p *Poset) Reset(block Block, frame Frame) error {

	//Clear all state
	p.LastConsensusRound = nil
	p.FirstConsensusRound = nil
	p.AnchorBlock = nil

	p.UndeterminedEvents = []string{}
	p.PendingRounds = []*pendingRound{}
	p.PendingLoadedEvents = 0
	p.topologicalIndex = 0

	cacheSize := p.Store.CacheSize()
	p.ancestorCache = common.NewLRU(cacheSize, nil)
	p.selfAncestorCache = common.NewLRU(cacheSize, nil)
	p.stronglySeeCache = common.NewLRU(cacheSize, nil)
	p.roundCache = common.NewLRU(cacheSize, nil)

	participants := p.Participants.ToPeerSlice()

	//Initialize new Roots
	rootMap := map[string]Root{}
	for id, root := range frame.Roots {
		p := participants[id]
		rootMap[p.PubKeyHex] = root
	}
	if err := p.Store.Reset(rootMap); err != nil {
		return err
	}

	//Insert Block
	if err := p.Store.SetBlock(block); err != nil {
		return err
	}

	p.setLastConsensusRound(block.RoundReceived())

	//Insert Frame Events
	for _, ev := range frame.Events {
		if err := p.InsertEvent(ev, false); err != nil {
			return err
		}
	}

	return nil
}

//Bootstrap loads all Events from the Store's DB (if there is one) and feeds
//them to the Poset (in topological order) for consensus ordering. After this
//method call, the Poset should be in a state coherent with the 'tip' of the
//Poset
func (p *Poset) Bootstrap() error {
	if badgerStore, ok := p.Store.(*BadgerStore); ok {
		//Retreive the Events from the underlying DB. They come out in topological
		//order
		topologicalEvents, err := badgerStore.dbTopologicalEvents()
		if err != nil {
			return err
		}

		//Insert the Events in the Poset
		for _, e := range topologicalEvents {
			if err := p.InsertEvent(e, true); err != nil {
				return err
			}
		}

		//Compute the consensus order of Events
		if err := p.DivideRounds(); err != nil {
			return err
		}
		if err := p.DecideFame(); err != nil {
			return err
		}
		if err := p.DecideRoundReceived(); err != nil {
			return err
		}
		if err := p.ProcessDecidedRounds(); err != nil {
			return err
		}
		if err := p.ProcessSigPool(); err != nil {
			return err
		}
	}

	return nil
}

//ReadWireInfo converts a WireEvent to an Event by replacing int IDs with the
//corresponding public keys.
func (p *Poset) ReadWireInfo(wevent WireEvent) (*Event, error) {
	selfParent := rootSelfParent(wevent.Body.CreatorID)
	otherParent := ""
	var err error

	creator := p.Participants.ById[wevent.Body.CreatorID]
	creatorBytes, err := hex.DecodeString(creator.PubKeyHex[2:])
	if err != nil {
		return nil, err
	}

	if wevent.Body.SelfParentIndex >= 0 {
		selfParent, err = p.Store.ParticipantEvent(creator.PubKeyHex, wevent.Body.SelfParentIndex)
		if err != nil {
			return nil, err
		}
	}
	if wevent.Body.OtherParentIndex >= 0 {
		otherParentCreator := p.Participants.ById[wevent.Body.OtherParentCreatorID]
		otherParent, err = p.Store.ParticipantEvent(otherParentCreator.PubKeyHex, wevent.Body.OtherParentIndex)
		if err != nil {
			//PROBLEM Check if other parent can be found in the root
			//problem, we do not known the WireEvent's EventHash, and
			//we do not know the creators of the roots RootEvents
			root, err := p.Store.GetRoot(creator.PubKeyHex)
			if err != nil {
				return nil, err
			}
			//loop through others
			found := false
			for _, re := range root.Others {
				if re.CreatorID == wevent.Body.OtherParentCreatorID &&
					re.Index == wevent.Body.OtherParentIndex {
					otherParent = re.Hash
					found = true
					break
				}
			}

			if !found {
				return nil, fmt.Errorf("OtherParent not found")
			}

		}
	}

	if len(wevent.FlagTable) == 0 {
		return nil, fmt.Errorf("flag table is null")
	}

	body := EventBody{
		Transactions:    wevent.Body.Transactions,
		BlockSignatures: wevent.BlockSignatures(creatorBytes),
		Parents:         []string{selfParent, otherParent},
		Creator:         creatorBytes,

		Index:                wevent.Body.Index,
		selfParentIndex:      wevent.Body.SelfParentIndex,
		otherParentCreatorID: wevent.Body.OtherParentCreatorID,
		otherParentIndex:     wevent.Body.OtherParentIndex,
		creatorID:            wevent.Body.CreatorID,
	}

	event := &Event{
		Body:      body,
		Signature: wevent.Signature,
		FlagTable: wevent.FlagTable,
	}

	return event, nil
}

//CheckBlock returns an error if the Block does not contain valid signatures
//from MORE than 1/3 of participants
func (p *Poset) CheckBlock(block Block) error {
	validSignatures := 0
	for _, s := range block.GetSignatures() {
		ok, _ := block.Verify(s)
		if ok {
			validSignatures++
		}
	}
	if validSignatures <= p.trustCount {
		return fmt.Errorf("Not enough valid signatures: got %d, need %d", validSignatures, p.trustCount+1)
	}

	p.logger.WithField("valid_signatures", validSignatures).Debug("CheckBlock")
	return nil
}

/*******************************************************************************
Setters
*******************************************************************************/

func (p *Poset) setLastConsensusRound(i int) {
	if p.LastConsensusRound == nil {
		p.LastConsensusRound = new(int)
	}
	*p.LastConsensusRound = i

	if p.FirstConsensusRound == nil {
		p.FirstConsensusRound = new(int)
		*p.FirstConsensusRound = i
	}
}

func (p *Poset) setAnchorBlock(i int) {
	if p.AnchorBlock == nil {
		p.AnchorBlock = new(int)
	}
	*p.AnchorBlock = i
}

/*******************************************************************************
   Helpers
*******************************************************************************/

func middleBit(ehex string) bool {
	hash, err := hex.DecodeString(ehex[2:])
	if err != nil {
		fmt.Printf("ERROR decoding hex string: %s\n", err)
	}
	if len(hash) > 0 && hash[len(hash)/2] == 0 {
		return false
	}
	return true
}
