package posnode

//NOTE: mockgen does not work properly out of GOPATH
//go:generate mockgen -package=posnode -source=consensus.go -destination=mock_test.go Consensus

import (
	"testing"

	"github.com/golang/mock/gomock"
)

func ExampleNode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	consensus := NewMockConsensus(ctrl)

	store := NewMemStore()

	n := NewForTests("server.fake", store, consensus)

	n.Start()
	defer n.Stop()

	select {}
}
