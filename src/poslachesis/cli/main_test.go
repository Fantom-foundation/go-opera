package main

import (
	"bytes"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"

	"github.com/Fantom-foundation/go-lachesis/src/network"
	"github.com/Fantom-foundation/go-lachesis/src/proxy/wire"
)

//go:generate mockgen -package=main -destination=mock_test.go github.com/Fantom-foundation/go-lachesis/src/proxy/wire CtrlServer

func TestApp(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mock := NewMockCtrlServer(ctrl)

	server := mockCtrlApp(mock)
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
		defer out.Reset()

		err := app.Execute()
		if !assert.NoError(err) {
			return
		}

		assert.Contains(out.String(), "id mock")
	})

	t.Run("stake", func(t *testing.T) {
		assert := assert.New(t)
		mock.EXPECT().Stake(gomock.Any(), gomock.Any()).Return(&wire.StakeResponse{
			Value: 0.0023,
		}, nil)

		app.SetArgs([]string{"stake"})
		defer out.Reset()

		err := app.Execute()
		if !assert.NoError(err) {
			return
		}

		assert.Contains(out.String(), "0.0023")
	})

	t.Run("internal_txn missing flags", func(t *testing.T) {
		assert := assert.New(t)

		app.SetArgs([]string{"internal_txn"})
		defer out.Reset()

		err := app.Execute()
		if !assert.Error(err) {
			return
		}

		assert.Contains(out.String(), "required flag(s) \"amount\", \"receiver\" not set")
	})

	t.Run("internal_txn", func(t *testing.T) {
		assert := assert.New(t)
		mock.EXPECT().InternalTxn(gomock.Any(), &wire.InternalTxnRequest{
			Amount:   2,
			Receiver: "receiver",
		}).Return(&empty.Empty{}, nil)

		app.SetArgs([]string{"internal_txn", "--amount=2", "--receiver=receiver"})
		defer out.Reset()

		err := app.Execute()
		if !assert.NoError(err) {
			return
		}

		assert.Contains(out.String(), "transaction has been added")
	})
}

func mockCtrlApp(srv wire.CtrlServer) *grpc.Server {
	s := grpc.NewServer()
	wire.RegisterCtrlServer(s, srv)

	listener := network.TCPListener("localhost:55557")

	go func() {
		if err := s.Serve(listener); err != nil {
			panic(err)
		}
	}()

	return s
}
