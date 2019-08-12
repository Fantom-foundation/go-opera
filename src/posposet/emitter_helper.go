package posposet

import (
	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/src/posposet/vector"
	"math/rand"
	"sort"
)

type SearchStrategy interface {
	Init(myLast *hash.Event)
	Find(heads hash.Events) hash.Event
}

// max is max num of parents to link with (including self-parent)
// returns set of parents to link, len(res) <= max
func (p *Poset) FindBestParents(me hash.Peer, max int, strategy SearchStrategy) []hash.Event {
	headsSet := p.store.GetHeads().Set()
	selfParent := p.store.GetLastEvent(me)

	res := make([]hash.Event, 0, max)
	if selfParent != nil {
		res = append(res, *selfParent)
		headsSet.Erase(*selfParent)
	}

	strategy.Init(selfParent)

	for ; len(res) < max && len(headsSet) > 0; {
		best := strategy.Find(headsSet.Slice())
		res = append(res, best)
		headsSet.Erase(best)
	}

	return res
}

/**
    SeeingStrategy
 */

type SeeingStrategy struct {
	vi       *vector.Index
	template []vector.HighestBefore
}

type eventScore struct {
	event hash.Event
	score idx.Event
	vec   []vector.HighestBefore
}

func (p *Poset) NewSeeingStrategy() *SeeingStrategy {
	return &SeeingStrategy{
		vi: p.events,
	}
}

func (st *SeeingStrategy) Init(selfParent *hash.Event) {
	if selfParent != nil {
		// we start searching by comparing with self-parent
		st.template = st.vi.GetEvent(*selfParent).HighestBefore
	}
}

func (st *SeeingStrategy) Find(heads hash.Events) hash.Event {
	scores := make([]eventScore, 0, 100)

	// estimate score of each head as number of members it sees higher than provided template
	for i, id := range heads {
		score := eventScore{}
		score.event = id
		score.vec = st.vi.GetEvent(id).HighestBefore
		if st.template == nil {
			st.template = make([]vector.HighestBefore, len(score.vec)) // nothing sees
		}
		for _, highest := range score.vec {
			// sees higher
			if highest.Seq > st.template[i].Seq {
				score.score += 1
			}
			// sees a fork
			if highest.IsForkSeen && !st.template[i].IsForkSeen {
				score.score += 1
			}
		}
		scores = append(scores, score)
	}

	// take the head with best score
	sort.Slice(scores, func(i, j int) bool {
		a, b := scores[i], scores[j]
		return a.score < b.score
	})
	// memorize its template for next calls
	st.template = scores[0].vec
	return scores[0].event
}

/**
    RandomStrategy
 */

type RandomStrategy struct {
	r *rand.Rand
}

func (p *Poset) NewRandomStrategy(r *rand.Rand) *RandomStrategy {
	return &RandomStrategy{
		r: r,
	}
}

func (st *RandomStrategy) Init(myLast *hash.Event) {}

func (st *RandomStrategy) Find(heads hash.Events) hash.Event {
	return heads[st.r.Intn(len(heads))]
}
