module github.com/Fantom-foundation/go-opera

go 1.14

require (
	github.com/DeBankDeFi/eth v1.9.3 // indirect
	github.com/Fantom-foundation/lachesis-base v0.0.0-20210420092627-c16f01e35562
	github.com/allegro/bigcache v1.2.1 // indirect
	github.com/aristanetworks/goarista v0.0.0-20191023202215-f096da5361bb // indirect
	github.com/btcsuite/btcd v0.20.1-beta // indirect
	github.com/certifi/gocertifi v0.0.0-20191021191039-0944d244cd40 // indirect
	github.com/cespare/cp v1.1.1
	github.com/davecgh/go-spew v1.1.1
	github.com/deckarep/golang-set v1.7.1
	github.com/docker/docker v1.13.1
	github.com/dvyukov/go-fuzz v0.0.0-20201127111758-49e582c6c23d
	github.com/edsrzf/mmap-go v1.0.0 // indirect
	github.com/ethereum/go-ethereum v1.9.25
	github.com/evalphobia/logrus_sentry v0.8.2
	github.com/fjl/memsize v0.0.0-20190710130421-bcb5799ab5e5
	github.com/gballet/go-libpcsclite v0.0.0-20191108122812-4678299bea08 // indirect
	github.com/getsentry/raven-go v0.2.0 // indirect
	github.com/golang/mock v1.4.4
	github.com/hashicorp/golang-lru v0.5.4
	github.com/holiman/uint256 v1.2.0
	github.com/influxdata/influxdb v1.7.9 // indirect
	github.com/karalabe/usb v0.0.0-20191104083709-911d15fe12a9 // indirect
	github.com/mattn/go-colorable v0.1.8
	github.com/mattn/go-isatty v0.0.12
	github.com/minio/minio v0.0.0-20210513042507-5c0a7189c763
	github.com/naoina/toml v0.1.2-0.20170918210437-9fafd6967416
	github.com/olekukonko/tablewriter v0.0.2 // indirect
	github.com/opentracing/opentracing-go v1.2.0
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.8.0
	github.com/prometheus/client_model v0.2.0
	github.com/prometheus/tsdb v0.10.0 // indirect
	github.com/sirupsen/logrus v1.8.0
	github.com/status-im/keycard-go v0.0.0-20190424133014-d95853db0f48
	github.com/stretchr/testify v1.6.1
	github.com/syndtr/goleveldb v1.0.1-0.20200815110645-5c35d600f0ca
	github.com/tyler-smith/go-bip39 v1.0.2
	github.com/uber/jaeger-client-go v2.20.1+incompatible
	github.com/uber/jaeger-lib v2.2.0+incompatible
	go.uber.org/atomic v1.5.1 // indirect
	golang.org/x/crypto v0.0.0-20200622213623-75b288015ac9
	gopkg.in/urfave/cli.v1 v1.20.0
)

replace github.com/ethereum/go-ethereum => github.com/Fantom-foundation/go-ethereum v1.9.22-ftm-0.5

replace github.com/dvyukov/go-fuzz => github.com/guzenok/go-fuzz v0.0.0-20210103140116-f9104dfb626f
