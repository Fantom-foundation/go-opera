// +build debug

// These functions are used only in debugging
package node

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/andrecronje/lachesis/src/poset"
)

type InfosLite struct {
	ParticipantEvents map[string]map[string]EventLite
	Rounds            []poset.RoundInfo
	Blocks            []poset.Block
}


type EventBodyLite struct {
	Parents         []string         //hashes of the event's parents, self-parent first
	Creator         string           //creator's public key
	Index           int64            //index in the sequence of events created by Creator
}

type EventLite struct {
	Body      EventBodyLite
	Signature string //creator's digital signature of body
	TopologicalIndex int

//	FlagTable []byte // FlagTable stores connection information
}


func (g *Graph) GetParticipantEventsLite() map[string]map[string]EventLite {
	res := make(map[string]map[string]EventLite)

	store := g.Node.core.poset.Store
	peers := g.Node.core.poset.Participants

//	for _, p := range peers.ByPubKey {
//		root, err := store.GetRoot(p.PubKeyHex)

//		if err != nil {
//			panic(err)
//		}

		//		evs, err := store.ParticipantEvents(p.PubKeyHex, root.SelfParent.Index)
		evs, err := store.TopologicalEvents()

		if err != nil {
			panic(err)
		}

		res[g.Node.localAddr/*p.PubKeyHex*/] = make(map[string]EventLite)

//		selfParent := fmt.Sprintf("Root%d", p.ID)
//
//		flagTable := make(map[string]int)
//		flagTable[selfParent] = 1

		// Create and save the first Event
//		ft, _ :=  json.Marshal(flagTable)
//		initialEvent := EventLite{
//			Body: EventBodyLite{
//				Parents: []string{},
//				Creator: "",
//				Index: 0,
//			},
//			FlagTable: ft,
//		}

//		res[p.PubKeyHex][root.SelfParent.Hash] = initialEvent

		for _, event := range evs {
//			event, err := store.GetEvent(e)

			if err != nil {
				panic(err)
			}

			hash := event.Hex()

			lite_event := EventLite{
				Body: EventBodyLite{
					Parents: event.Body.Parents,
					Creator: peers.ByPubKey[event.Creator()].NetAddr,
					Index: event.Body.Index,
				},
				Signature: event.Signature,
//				TopologicalIndex: event.TopologicalIndex,
//				FlagTable: event.FlagTable,
			}

			res[g.Node.localAddr/*p.PubKeyHex*/][hash] = lite_event
		}
//	}

	return res
}

func (g *Graph) GetInfosLite() InfosLite {
	return InfosLite{
		ParticipantEvents: g.GetParticipantEventsLite(),
		Rounds:            g.GetRounds(),
    Blocks:            g.GetBlocks(),
	}
}

func (c *Core) PrintStat() {
	fmt.Println("**core.HexID=", c.HexID())
	c.poset.PrintStat()
}

func (n *Node) PrintStat() {
	fmt.Println("*Node=", n.localAddr)
	g := NewGraph(n)
	//res, _ := json.Marshal(g.GetInfos())
	//fmt.Println("Node Graph=", res)
	encoder := json.NewEncoder(os.Stdout)
	res := g.GetInfosLite()
	encoder.Encode(res)
	n.core.PrintStat()
}
