package main

import (
	"bytes"
	"testing"

	"github.com/Fantom-foundation/go-lachesis/src/network"
	wire "github.com/Fantom-foundation/go-lachesis/src/proxy/wire"
	gomock "github.com/golang/mock/gomock"
	empty "github.com/golang/protobuf/ptypes/empty"
	"github.com/stretchr/testify/assert"
	grpc "google.golang.org/grpc"
)

//go:generate mockgen -package=main -source=../../proxy/wire/grpc.pb.go -destination=mock_test.go

func TestApp(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mock := NewMockManagementServer(ctrl)

	server := mockManagementServer(mock)
	defer server.Stop()

	app := prepareApp()
	var out bytes.Buffer
	app.SetOutput(&out)

	t.Run("id", func(t *testing.T) {
		assert := assert.New(t)
		mock.EXPECT().ID(gomock.Any(), gomock.Any()).Return(&wire.IDResponse{
			Id: "id mock",
		}, nil)

		app.SetArgs([]string{"id"})
		app.Execute()

		assert.Contains(out.String(), "id mock")
	})

	t.Run("stake", func(t *testing.T) {
		assert := assert.New(t)
		mock.EXPECT().Stake(gomock.Any(), gomock.Any()).Return(&wire.StakeResponse{
			Value: 0.0023,
		}, nil)

		app.SetArgs([]string{"stake"})
		app.Execute()

		assert.Contains(out.String(), "0.0023")
	})

	t.Run("internal_txn", func(t *testing.T) {
		assert := assert.New(t)
		mock.EXPECT().InternalTxn(gomock.Any(), &wire.InternalTxnRequest{
			Amount:   2,
			Receiver: "receiver",
		}).Return(&empty.Empty{}, nil)

		app.SetArgs([]string{"internal_txn", "--amount=2", "--to=receiver"})
		app.Execute()

		assert.Contains(out.String(), "transaction has been added")
	})

}

func mockManagementServer(srv wire.ManagementServer) *grpc.Server {
	s := grpc.NewServer()
	wire.RegisterManagementServer(s, srv)

	listen := network.TCPListener
	listener := listen("localhost:55557")

	go func() {
		if err := s.Serve(listener); err != nil {
			panic(err)
		}
	}()

	return s
}
