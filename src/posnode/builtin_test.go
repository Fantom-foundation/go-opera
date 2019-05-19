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
			name: "with hosts",
			hosts: []string{
				"194.54.152.2:55555",
				"194.54.152.3:55555",
				"194.54.152.4:55555",
				"194.54.152.5:55555",
				"194.54.152.6:55555",
			},
			expect: []string{
				"194.54.152.2:55555",
				"194.54.152.3:55555",
				"194.54.152.4:55555",
				"194.54.152.5:55555",
				"194.54.152.6:55555",
			},
		},
		{
			name:  "empty hosts",
			hosts: make([]string, 0),
			expect: []string{
				"",
				"",
			},
		},
		{
			name: "combined",
			hosts: []string{
				"194.54.152.2:55555",
				"194.54.152.3:55555",
				"194.54.152.4:55555",
			},
			expect: []string{
				"194.54.152.2:55555",
				"194.54.152.3:55555",
				"194.54.152.4:55555",
				"",
				"",
			},
		},
	}

	for _, tc := range tt {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert := assert.New(t)

			n := NewForTests("", nil, nil)
			n.builtin.hosts = tc.hosts

			for _, expect := range tc.expect {
				assert.Equal(expect, n.NextBuiltInPeer())
			}
		})
	}
}
