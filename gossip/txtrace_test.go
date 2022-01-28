package gossip

// SynthereumDeployer contract
//go:generate bash -c "docker run --rm -v $(pwd)/contract/SynthereumDeployer:/src -v $(pwd)/contract:/dst ethereum/solc:0.8.4 -o /dst/solc/ --optimize --optimize-runs=200 --bin --abi --allow-paths /src --overwrite /src/Deployer.sol"
// you have to use abigen after github.com/ethereum/go-ethereum/pull/23940
//go:generate bash -c "cd ${GOPATH}/src/github.com/ethereum/go-ethereum && go run ./cmd/abigen --bin=${PWD}/contract/solc/SynthereumDeployer.bin --abi=${PWD}/contract/solc/SynthereumDeployer.abi --pkg=SynthereumDeployer --type=Contract --out=${PWD}/contract/SynthereumDeployer/contract.go"
