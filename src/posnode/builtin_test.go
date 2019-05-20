package posnode

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNodeNextBuiltInPeer(t *testing.T) {
	tt := []struct {
		name   string
		hosts  []string
		expect []string
	}{
		{
			name:  "empty",
			hosts: make([]string, 0),
			expect: []string{
				"",
				"",
			},
		},
		{
			name: "once",
			hosts: []string{
				"194.54.152.2",
				"194.54.152.3",
				"194.54.152.4",
			},
			expect: []string{
				"194.54.152.2",
				"194.54.152.3",
				"194.54.152.4",
			},
		},
		{
			name: "twice",
			hosts: []string{
				"194.54.152.2",
				"194.54.152.3",
				"194.54.152.4",
			},
			expect: []string{
				"194.54.152.2",
				"194.54.152.3",
				"194.54.152.4",
				"194.54.152.2",
				"194.54.152.3",
				"194.54.152.4",
			},
		},
	}

	for _, tc := range tt {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert := assert.New(t)

			n := NewForTests("", nil, nil)
			n.AddBuiltInPeers(tc.hosts...)

			for i, expect := range tc.expect {
				if !assert.Equal(expect, n.NextBuiltInPeer(), "item %d", i) {
					break
				}
			}
		})
	}
}
