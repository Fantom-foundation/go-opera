export MOCKGEN?=mockgen
export PROTOC?=protoc
export RM?=rm

.PHONY: build clean proto buildtests

clean:
	$(RM) -rf *.pb.go

proto: $(PBGOS)

build:

buildtests: mock_test.go

mock_test.go: main.go
	$(MOCKGEN) -package=main -destination=mock_test.go github.com/Fantom-foundation/go-lachesis/src/proxy Node,Consensus
