package posnode

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Fantom-foundation/go-lachesis/src/logger"
)

func TestNodeBuiltInPeer(t *testing.T) {
	logger.SetTestMode(t)

	ttt := []struct {
		name string
		in   []string
		out  []string
	}{
		{
			name: "empty",
			in:   []string{},
			out: []string{
				"",
				"",
			},
		},
		{
			name: "non empty",
			in: []string{
				"194.54.152.2",
				"lachesis-node-1",
				"google.com",
			},
			out: []string{
				"194.54.152.2",
				"lachesis-node-1",
				"google.com",
				"194.54.152.2",
				"lachesis-node-1",
				"google.com",
			},
		},
	}

	for _, tt := range ttt {
		t.Run(tt.name, func(t *testing.T) {
			n := NewForTests("", nil, nil)
			n.AddBuiltInPeers(tt.in...)

			for i, expect := range tt.out {
				if !assert.Equalf(t, expect, n.NextBuiltInPeer(), "item %d", i) {
					break
				}
			}

		})
	}
}
