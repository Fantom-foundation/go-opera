package node

import (
	"fmt"

	"github.com/andrecronje/lachesis/src/poset"
)

type Infos struct {
	ParticipantEvents map[string]map[string]poset.Event
	Rounds            []poset.RoundInfo
  Blocks            []poset.Block
}

type Graph struct {
	Node *Node
}

func (g *Graph) GetBlocks() []poset.Block {
	res := []poset.Block{}
 	blockIdx := 0
 	for blockIdx <= g.Node.core.poset.Store.LastBlockIndex() {
		r, err := g.Node.core.poset.Store.GetBlock(blockIdx)
 		if err != nil {
			break
		}
 		res = append(res, r)
 		blockIdx++
	}
 	return res
}

func (g *Graph) GetParticipantEvents() map[string]map[string]poset.Event {
	res := make(map[string]map[string]poset.Event)

	for _, p := range g.Node.core.participants.ByPubKey {
		root, _ := g.Node.core.poset.Store.GetRoot(p.PubKeyHex)

		evs, _ := g.Node.core.poset.Store.ParticipantEvents(p.PubKeyHex, -1)

		res[p.PubKeyHex] = make(map[string]poset.Event)

		selfParent := fmt.Sprintf("Root%d", p.ID)

		flagTable := make(map[string]int)
		flagTable[selfParent] = 1

		// Create and save the first Event
		initialEvent := poset.NewEvent([][]byte{}, []poset.BlockSignature{},
			[]string{}, []byte{}, 0, flagTable)

		res[p.PubKeyHex][root.SelfParent.Hash] = initialEvent

		for _, e := range evs {
			event, _ := g.Node.core.GetEvent(e)

			hash := event.Hex()

			res[p.PubKeyHex][hash] = event
		}
	}

	return res
}

func (g *Graph) GetRounds() []poset.RoundInfo {
	res := []poset.RoundInfo{}

	round := 0

	for round <= g.Node.core.poset.Store.LastRound() {
		r, err := g.Node.core.poset.Store.GetRound(round)

		if err != nil || !r.IsQueued() {
			break
		}

		res = append(res, r)

		round++
	}

	return res
}

func (g *Graph) GetInfos() Infos {
	return Infos{
		ParticipantEvents: g.GetParticipantEvents(),
		Rounds:            g.GetRounds(),
    Blocks:            g.GetBlocks(),
	}
}

func NewGraph(n *Node) *Graph {
	return &Graph{
		Node: n,
	}
}
