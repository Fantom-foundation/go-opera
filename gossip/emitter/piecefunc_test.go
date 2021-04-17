package emitter

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Fantom-foundation/go-opera/gossip/emitter/piecefunc"
)

func TestPiecefunc(t *testing.T) {
	t.Run("confirmingEmitIntervalF", func(t *testing.T) {
		testPiecefunc(t, confirmingEmitIntervalF)
	})

	t.Run("scalarUpdMetricF", func(t *testing.T) {
		testPiecefunc(t, scalarUpdMetricF)
	})

	t.Run("eventMetricF", func(t *testing.T) {
		testPiecefunc(t, eventMetricF)
	})
}

func testPiecefunc(t *testing.T, pieces []piecefunc.Dot) {
	require := require.New(t)

	require.GreaterOrEqual(len(pieces), 2)
	var prev uint64
	for i, piece := range pieces {
		if i >= 1 {
			require.Greater(piece.X, prev)
		}
		prev = piece.X
	}
}
