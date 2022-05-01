module github.com/Fantom-foundation/go-opera

go 1.14

require (
	github.com/Fantom-foundation/lachesis-base v0.0.0-20220103160934-6b4931c60582
	github.com/VictoriaMetrics/fastcache v1.9.0 // indirect
	github.com/allegro/bigcache v1.2.1 // indirect
	github.com/btcsuite/btcd v0.22.0-beta // indirect
	github.com/certifi/gocertifi v0.0.0-20191021191039-0944d244cd40 // indirect
	github.com/cespare/cp v1.1.1
	github.com/davecgh/go-spew v1.1.1
	github.com/deckarep/golang-set v1.7.1
	github.com/docker/docker v1.13.1
	github.com/dvyukov/go-fuzz v0.0.0-20201127111758-49e582c6c23d
	github.com/ethereum/go-ethereum v1.10.8
	github.com/evalphobia/logrus_sentry v0.8.2
	github.com/fjl/memsize v0.0.0-20190710130421-bcb5799ab5e5
	github.com/fsnotify/fsnotify v1.5.1 // indirect
	github.com/gballet/go-libpcsclite v0.0.0-20191108122812-4678299bea08 // indirect
	github.com/getsentry/raven-go v0.2.0 // indirect
	github.com/go-kit/kit v0.10.0 // indirect
	github.com/go-ole/go-ole v1.2.6 // indirect
	github.com/golang/mock v1.4.4
	github.com/google/go-cmp v0.5.7 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/gorilla/websocket v1.5.0 // indirect
	github.com/hashicorp/golang-lru v0.5.5-0.20210104140557-80c98217689d
	github.com/holiman/bloomfilter/v2 v2.0.3
	github.com/jackpal/go-nat-pmp v1.0.2 // indirect
	github.com/julienschmidt/httprouter v1.3.0 // indirect
	github.com/karalabe/usb v0.0.0-20191104083709-911d15fe12a9 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/ledgerwatch/erigon-lib v0.0.0-20220428083508-625c9f5385d2
	github.com/ledgerwatch/log/v3 v3.4.1
	github.com/mattn/go-colorable v0.1.11
	github.com/mattn/go-isatty v0.0.14
	github.com/naoina/toml v0.1.2-0.20170918210437-9fafd6967416
	github.com/niemeyer/pretty v0.0.0-20200227124842-a10e7caefd8e // indirect
	github.com/onsi/ginkgo v1.16.5 // indirect
	github.com/onsi/gomega v1.17.0 // indirect
	github.com/opentracing/opentracing-go v1.1.0
	github.com/pkg/errors v0.9.1
	github.com/prometheus/tsdb v0.10.0 // indirect
	github.com/rjeczalik/notify v0.9.2 // indirect
	github.com/rs/cors v1.8.2 // indirect
	github.com/sirupsen/logrus v1.6.0
	github.com/status-im/keycard-go v0.0.0-20190424133014-d95853db0f48
	github.com/stretchr/testify v1.7.1
	github.com/syndtr/goleveldb v1.0.1-0.20210305035536-64b5b1c73954
	github.com/tklauser/go-sysconf v0.3.9 // indirect
	github.com/tyler-smith/go-bip39 v1.0.2
	github.com/uber/jaeger-client-go v2.20.1+incompatible
	github.com/uber/jaeger-lib v2.2.0+incompatible
	golang.org/x/crypto v0.0.0-20220411220226-7b82a4e95df4
	golang.org/x/time v0.0.0-20220224211638-0e9765cccd65 // indirect
	google.golang.org/genproto v0.0.0-20200825200019-8632dd797987 // indirect
	gopkg.in/check.v1 v1.0.0-20200902074654-038fdea0a05b // indirect
	gopkg.in/urfave/cli.v1 v1.20.0
)

replace github.com/ledgerwatch/erigon => github.com/ledgerwatch/erigon v1.9.7-0.20220421151921-057740ac2019

replace github.com/ethereum/go-ethereum => github.com/Fantom-foundation/go-ethereum v1.10.8-ftm-rc4

replace github.com/dvyukov/go-fuzz => github.com/guzenok/go-fuzz v0.0.0-20210103140116-f9104dfb626f
