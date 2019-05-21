package posnode

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNodeAddBuiltInPeers(t *testing.T) {
	tt := struct {
		in  []string
		out []string
	}{
		in: []string{
			"192.168.0.2",
			"lachesis-node-1",
			"192.168.0.2:55555",
			"google.com",
			"",
		},
		out: []string{
			"192.168.0.2",
			"lachesis-node-1",
			"google.com",
		},
	}

	assert := assert.New(t)

	n := NewForTests("", nil, nil)
	n.AddBuiltInPeers(tt.in...)

	assert.Equal(tt.out, n.builtin.hosts)
}

func TestNodeNextBuiltInPeer(t *testing.T) {
	tt := struct {
		in  []string
		out []string
	}{
		in: []string{
			"194.54.152.2",
			"194.54.152.3",
			"194.54.152.4",
		},
		out: []string{
			"194.54.152.2",
			"194.54.152.3",
			"194.54.152.4",
			"194.54.152.2",
			"194.54.152.3",
			"194.54.152.4",
		},
	}

	assert := assert.New(t)

	n := NewForTests("", nil, nil)
	n.AddBuiltInPeers(tt.in...)

	for i, expect := range tt.out {
		if !assert.Equal(expect, n.NextBuiltInPeer(), "item %d", i) {
			break
		}
	}
}
