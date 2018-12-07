// +build debug

// These functions are used only in debugging
package node

import (
	"encoding/json"
	"os"
	"os/signal"

	"github.com/Fantom-foundation/go-lachesis/src/poset"
	"github.com/sirupsen/logrus"
        "github.com/tebeka/atexit"
)

type InfosLite struct {
	ParticipantEvents map[string]map[string]EventLite
	Rounds            []poset.RoundCreated
	Blocks            []poset.Block
}


type EventBodyLite struct {
	Parents         []string         //hashes of the event's parents, self-parent first
	Creator         string           //creator's public key
	Index           int64            //index in the sequence of events created by Creator
	Transactions    [][]byte
}

type EventMessageLite struct {
	Body      EventBodyLite
	Signature string //creator's digital signature of body
	TopologicalIndex int64
	Hex              string
	Round            int64
	RoundReceived    int64

	WitnessProof []string
//	FlagTable []byte // FlagTable stores connection information
}
type EventLite struct {
	CreatorID            int64
	OtherParentCreatorID int64
	Message EventMessageLite
}


func (g *Graph) GetParticipantEventsLite() map[string]map[string]EventLite {
	res := make(map[string]map[string]EventLite)

	store := g.Node.core.poset.Store
	peers := g.Node.core.poset.Participants


	//		evs, err := store.ParticipantEvents(p.PubKeyHex, root.SelfParent.Index)
	evs, err := store.TopologicalEvents()

	if err != nil {
		panic(err)
	}

	res[g.Node.localAddr/*p.PubKeyHex*/] = make(map[string]EventLite)


	for _, event := range evs {

		if err != nil {
			panic(err)
		}

		hash := event.Hex()

		lite_event := EventLite{
			CreatorID: event.CreatorID(),
			OtherParentCreatorID: event.OtherParentCreatorID(),
			Message: EventMessageLite {
				Body: EventBodyLite{
					Parents: event.Message.Body.Parents,
					Creator: peers.ByPubKey[event.Creator()].NetAddr,
					Index: event.Message.Body.Index,
					Transactions: event.Message.Body.Transactions,
				},
				Hex: event.Message.Hex,
				Signature: event.Message.Signature,
				WitnessProof: event.Message.WitnessProof,
				Round: event.Message.Round,
				RoundReceived: event.Message.RoundReceived,
				TopologicalIndex: event.Message.TopologicalIndex,
				//				FlagTable: event.FlagTable,
			},
		}

		res[g.Node.localAddr/*p.PubKeyHex*/][hash] = lite_event
	}

	return res
}

func (g *Graph) GetInfosLite() InfosLite {
	return InfosLite{
		ParticipantEvents: g.GetParticipantEventsLite(),
		Rounds:            g.GetRounds(),
    Blocks:            g.GetBlocks(),
	}
}

func (c *Core) PrintStat(logger *logrus.Entry) {
	logger.Warn("**core.HexID=", c.HexID())
	c.poset.PrintStat(logger)
}

func (n *Node) PrintStat() {
	n.logger.Warn("*Node=", n.localAddr)
	g := NewGraph(n)
	encoder := json.NewEncoder(n.logger.Logger.Out)
	encoder.SetIndent("", "  ")
	res := g.GetInfosLite()
	encoder.Encode(res)
	n.core.PrintStat(n.logger)
}

func (n *Node) Register() {
	atexit.Register(func() {
		// You must build with debug tag to have PrintStat() defined
		n.PrintStat()
	})
	// use the following way of exit to execute registered atexit handlers:
	// import "github.com/tebeka/atexit"
	// atexit.Exit(0)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	go func () {
		<-signalChan
		atexit.Exit(13)
	}()
}
