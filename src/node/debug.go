// +build debug

// Package node these functions are used only in debugging
package node

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"sort"
	"strings"
	"sync"

	"github.com/Fantom-foundation/go-lachesis/src/poset"
	"github.com/sirupsen/logrus"
	"github.com/tebeka/atexit"
)

// InfosLite small subset of debug info for node
type InfosLite struct {
	ParticipantEvents map[string]map[string]EventLite
	Rounds            []poset.RoundCreated
	Blocks            []poset.Block
}

// EventBodyLite small subset of event body for debugging
type EventBodyLite struct {
	Parents      [][]byte // hashes of the event's parents, self-parent first
	Creator      string   // creator's public key
	Index        int64    // index in the sequence of events created by Creator
	Transactions [][]byte
}

// EventMessageLite small subset of event body for debugging
type EventMessageLite struct {
	Body             EventBodyLite
	Signature        string // creator's digital signature of body
	TopologicalIndex int64
	Hash             string
	Round            int64
	RoundReceived    int64

	ClothoProof [][]byte
	FlagTable   []byte // FlagTable stores connection information
}

// EventLite small subset of event for debugging
type EventLite struct {
	CreatorID            uint64
	OtherParentCreatorID uint64
	Message              EventMessageLite
	LamportTimestamp     int64
}

// GetParticipantEventsLite returns all participants
func (g *Graph) GetParticipantEventsLite() map[string]map[string]EventLite {
	res := make(map[string]map[string]EventLite)

	store := g.Node.core.poset.Store
	peers := g.Node.core.poset.Participants

	// evs, err := store.ParticipantEvents(p.PubKeyHex, root.SelfParent.Index)
	evs, err := store.TopologicalEvents()

	if err != nil {
		panic(err)
	}

	res[g.Node.localAddr /*p.PubKeyHex*/] = make(map[string]EventLite)

	for _, event := range evs {

		if err != nil {
			panic(err)
		}

		peer, ok := peers.ReadByPubKey(event.GetCreator())
		if !ok {
			panic(fmt.Sprintf("Creator %v not found", event.GetCreator()))
		}

		hash := event.Hash()

		liteEvent := EventLite{
			CreatorID:            event.CreatorID(),
			OtherParentCreatorID: event.OtherParentCreatorID(),
			LamportTimestamp:     event.GetLamportTimestamp(),
			Message: EventMessageLite{
				Body: EventBodyLite{
					Parents:      event.Message.Body.Parents,
					Creator:      peer.NetAddr,
					Index:        event.Message.Body.Index,
					Transactions: event.Message.Body.Transactions,
				},
				Hash:        hash.String(),
				Signature:   event.Message.Signature,
				ClothoProof: event.Message.ClothoProof,
				FlagTable:   event.Message.FlagTable,
				//Round:            event.Message.Round,
				//RoundReceived:    event.Message.RoundReceived,
				TopologicalIndex: event.Message.TopologicalIndex,
				// 				FlagTable: event.FlagTable,
			},
		}

		res[g.Node.localAddr /*p.PubKeyHex*/][hash.String()] = liteEvent
	}

	return res
}

// GetInfosLite returns debug stats
func (g *Graph) GetInfosLite() InfosLite {
	return InfosLite{
		ParticipantEvents: g.GetParticipantEventsLite(),
		Rounds:            g.GetRounds(),
		Blocks:            g.GetBlocks(),
	}
}

// PrintStat prints debug stats
func (c *Core) PrintStat(logger *logrus.Entry) {
	logger.Warn("**core.HexID=", c.HexID())
	logger.Warn("****Known events:")
	for pidID, index := range c.KnownEvents() {
		peer, ok := c.participants.ReadByID(uint64(pidID))
		if ok {
			logger.Warn("    index=", index, " peer=", peer.NetAddr,
				" pubKeyHex=", peer.PubKeyHex)
		}
	}
	c.poset.PrintStat(logger)
}

// PrintStat prints debug stats
func (n *Node) PrintStat() {
	n.logger.Warn("*Node=", n.localAddr)

	g := NewGraph(n)

	encoder := json.NewEncoder(n.logger.Logger.Out)
	encoder.SetIndent("", "  ")

	res := g.GetInfosLite()
	encoder.Encode(res)

	n.core.PrintStat(n.logger)

	file, err := os.Create(fmt.Sprintf("Node_%v.gv", n.localAddr))
	if err != nil {
		panic(err)
	}
	fmt.Fprintf(file, "digraph ANode { /* %v */\nrankdir=RL; ranksep=1.5;\n", n.localAddr)

	fr := make(map[int64]map[string][]EventLite)
	cr := make(map[string][]EventLite)
	maxFrame := int64(0)
	for _, events := range res.ParticipantEvents {
		for _, le := range events {
			if le.Message.Round > maxFrame {
				maxFrame = le.Message.Round
			}
			fr[le.Message.Round] = make(map[string][]EventLite)
			cr[le.Message.Body.Creator] = append(cr[le.Message.Body.Creator], le)
		}
	}

	for _, events := range res.ParticipantEvents {
		for _, le := range events {
			fr[le.Message.Round][le.Message.Body.Creator] = append(fr[le.Message.Round][le.Message.Body.Creator], le)
		}
	}

	fmt.Fprint(file, "layers = \"f0")
	for i := int64(1); i <= maxFrame; i++ {
		fmt.Fprintf(file, ":f%v", i)
	}
	fmt.Fprintln(file, "\";")

	var keys []string
	for k, _ := range cr {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var replacer = strings.NewReplacer(".", "_", ":", "_")

	for _, creator := range keys {
		lightEvents := cr[creator]
		fmt.Fprintf(file, "subgraph cluster_%v { rank = same; ranksep = 2.5; ", replacer.Replace(creator))
		for _, le := range lightEvents {
			fmt.Fprintf(file, "v%v [shape=none,layer=\"f%v\" label=<<TABLE BORDER=\"0\" CELLBORDER=\"1\" CELLSPACING=\"0\" CELLPADDING=\"4\"><TR><TD>r</TD><td>l</td><td>rr</td><td>ind</td><td>cr</td></TR><tr><td>%v</td><td>%v</td><td>%v</td><td>%v</td><td>%v</td></tr><tr><td colspan=\"5\">ft:",
				le.Message.Hash, le.Message.Round, le.Message.Round, le.LamportTimestamp, le.Message.RoundReceived, le.Message.Body.Index, le.Message.Body.Creator)
			ft := poset.NewFlagTable()
			ft.Unmarshal(le.Message.FlagTable)
			for k, v := range ft {
				ev, err := n.core.poset.Store.GetEventBlock(k)
				if err == nil {
					peer, ok := n.core.poset.Participants.ReadByPubKey(ev.GetCreator())
					if !ok {
						panic(fmt.Sprintf("Peer %v not found", k))
					}
					fmt.Fprintf(file, " %v:%v", peer.NetAddr, v)
				} else {
					fmt.Fprintf(file, " ???:%v", v)
				}
			}

			fmt.Fprintf(file, "</td></tr></TABLE>>];\n")
			n.logger.Warnf("v%v r:%v rr:%v ind:%v cr:%v l:%v", le.Message.Hash, le.Message.Round, le.Message.RoundReceived, le.Message.Body.Index, le.Message.Body.Creator, le.LamportTimestamp)
		}
		fmt.Fprint(file, " }\n")
	}

	for _, events := range res.ParticipantEvents {
		for _, le := range events {
			var parent, otherParent poset.EventHash
			parent.Set(le.Message.Body.Parents[0])
			otherParent.Set(le.Message.Body.Parents[1])
			if !parent.Zero() {
				fmt.Fprintf(file, "v%v -> v%v;\n", le.Message.Hash, parent.String())
			}
			if !otherParent.Zero() {
				fmt.Fprintf(file, "v%v -> v%v;\n", le.Message.Hash, otherParent.String())
			}
		}
	}
	fmt.Fprintln(file, "}\n")
	file.Close()
}

// Register a print listener
func (n *Node) Register() {
	var once sync.Once
	onceBody := func() {
		// You must build with debug tag to have PrintStat() defined
		n.PrintStat()
	}
	atexit.Register(func() {
		once.Do(onceBody)
	})
	// use the following way of exit to execute registered atexit handlers:
	// import "github.com/tebeka/atexit"
	// atexit.Exit(0)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	go func() {
		<-signalChan
		atexit.Exit(13)
	}()
}
