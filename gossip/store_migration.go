package gossip

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/Fantom-foundation/lachesis-base/kvdb"
	"github.com/Fantom-foundation/lachesis-base/kvdb/table"
	"github.com/Fantom-foundation/lachesis-base/lachesis"
	"github.com/ethereum/go-ethereum/common"

	"github.com/Fantom-foundation/go-opera/inter"
	"github.com/Fantom-foundation/go-opera/inter/iblockproc"
	"github.com/Fantom-foundation/go-opera/opera"
	"github.com/Fantom-foundation/go-opera/utils/concurrent"
	"github.com/Fantom-foundation/go-opera/utils/migration"
)

func isEmptyDB(db kvdb.Iteratee) bool {
	it := db.NewIterator(nil, nil)
	defer it.Release()
	return !it.Next()
}

func (s *Store) migrateData() error {
	versions := migration.NewKvdbIDStore(s.table.Version)
	if isEmptyDB(s.mainDB) {
		// short circuit if empty DB
		versions.SetID(s.migrations().ID())
		return nil
	}

	err := s.migrations().Exec(versions, s.flushDBs)
	return err
}

func (s *Store) migrations() *migration.Migration {
	return migration.
		Begin("opera-gossip-store").
		Next("used gas recovery", unsupportedMigration).
		Next("tx hashes recovery", s.recoverTxHashes).
		Next("DAG heads recovery", s.recoverHeadsStorage).
		Next("DAG last events recovery", s.recoverLastEventsStorage).
		Next("BlockState recovery", s.recoverBlockState).
		Next("LlrState recovery", s.recoverLlrState).
		Next("erase gossip-async db", s.eraseGossipAsyncDB).
		Next("erase SFC API table", s.eraseSfcApiTable).
		Next("erase legacy genesis DB", s.eraseGenesisDB).
		Next("calculate upgrade heights", s.calculateUpgradeHeights)
}

func unsupportedMigration() error {
	return fmt.Errorf("DB version isn't supported, please restart from scratch")
}

var (
	fixTxHash1  = common.HexToHash("0xb6840d4c0eb562b0b1731760223d91b36edc6c160958e23e773e6058eea30458")
	fixTxEvent1 = hash.HexToEventHash("0x00001718000003d4d3955bf592e12fb80a60574fa4b18bd5805b4c010d75e86d")
	fixTxHash2  = common.HexToHash("0x3aeede91740093cb8feb1296a34cf70d86d2f802cff860edd798978e94a40534")
	fixTxEvent2 = hash.HexToEventHash("0x0000179e00000c464d756a7614d0ca067fcb37ee4452004bf308c9df561e85e8")
)

const (
	fixTxEventPos1 = 2
	fixTxBlock1    = 4738821
	fixTxEventPos2 = 0
	fixTxBlock2    = 4801307
)

func fixEventTxHashes(e *inter.EventPayload) {
	if e.ID() == fixTxEvent1 {
		e.Txs()[fixTxEventPos1].SetHash(fixTxHash1)
	}
	if e.ID() == fixTxEvent2 {
		e.Txs()[fixTxEventPos2].SetHash(fixTxHash2)
	}
}

type kvEntry struct {
	key   []byte
	value []byte
}

func (s *Store) recoverTxHashes() error {
	if s.GetRules().NetworkID != 0xfa {
		return nil
	}

	diff1 := []kvEntry{
		{key: common.Hex2Bytes("4c720000000000484f05497bf2b3f14a9d4f4e8ce89fddc26e4c912393e9f6a71dc188c19fa7115cfba70000000000000007"), value: nil},
		{key: common.Hex2Bytes("4c720000000000484f05497bf2b3f14a9d4f4e8ce89fddc26e4c912393e9f6a71dc188c19fa7115cfba70000000000000008"), value: nil},
		{key: common.Hex2Bytes("4c720000000000484f05497bf2b3f14a9d4f4e8ce89fddc26e4c912393e9f6a71dc188c19fa7115cfba70000000000000009"), value: nil},
		{key: common.Hex2Bytes("4c720000000000484f05497bf2b3f14a9d4f4e8ce89fddc26e4c912393e9f6a71dc188c19fa7115cfba7000000000000000a"), value: nil},
		{key: common.Hex2Bytes("4c740000000000000000000000000000000000000000000000000000000000000000020000000000484f05497bf2b3f14a9d4f4e8ce89fddc26e4c912393e9f6a71dc188c19fa7115cfba70000000000000007"), value: nil},
		{key: common.Hex2Bytes("4c740000000000000000000000000000000000000000000000000000000000000000020000000000484f05497bf2b3f14a9d4f4e8ce89fddc26e4c912393e9f6a71dc188c19fa7115cfba70000000000000008"), value: nil},
		{key: common.Hex2Bytes("4c740000000000000000000000000000000000000000000000000000000000000001030000000000484f05497bf2b3f14a9d4f4e8ce89fddc26e4c912393e9f6a71dc188c19fa7115cfba7000000000000000a"), value: nil},
		{key: common.Hex2Bytes("4c7400000000000000000000000004d02149058cc8c8d0cf5f6fd1dc5394a80d7371020000000000484f05497bf2b3f14a9d4f4e8ce89fddc26e4c912393e9f6a71dc188c19fa7115cfba7000000000000000a"), value: nil},
		{key: common.Hex2Bytes("4c7400000000000000000000000004d02149058cc8c8d0cf5f6fd1dc5394a80d7371030000000000484f05497bf2b3f14a9d4f4e8ce89fddc26e4c912393e9f6a71dc188c19fa7115cfba70000000000000009"), value: nil},
		{key: common.Hex2Bytes("4c740000000000000000000000004d5362dd18ea4ba880c829b0152b7ba371741e59030000000000484f05497bf2b3f14a9d4f4e8ce89fddc26e4c912393e9f6a71dc188c19fa7115cfba70000000000000007"), value: nil},
		{key: common.Hex2Bytes("4c740000000000000000000000005cc61a78f164885776aa610fb0fe1257df78e59b000000000000484f05497bf2b3f14a9d4f4e8ce89fddc26e4c912393e9f6a71dc188c19fa7115cfba70000000000000007"), value: nil},
		{key: common.Hex2Bytes("4c740000000000000000000000005cc61a78f164885776aa610fb0fe1257df78e59b000000000000484f05497bf2b3f14a9d4f4e8ce89fddc26e4c912393e9f6a71dc188c19fa7115cfba70000000000000008"), value: nil},
		{key: common.Hex2Bytes("4c740000000000000000000000005cc61a78f164885776aa610fb0fe1257df78e59b000000000000484f05497bf2b3f14a9d4f4e8ce89fddc26e4c912393e9f6a71dc188c19fa7115cfba70000000000000009"), value: nil},
		{key: common.Hex2Bytes("4c740000000000000000000000009083ea3756bde6ee6f27a6e996806fbd37f6f093000000000000484f05497bf2b3f14a9d4f4e8ce89fddc26e4c912393e9f6a71dc188c19fa7115cfba7000000000000000a"), value: nil},
		{key: common.Hex2Bytes("4c740000000000000000000000009083ea3756bde6ee6f27a6e996806fbd37f6f093020000000000484f05497bf2b3f14a9d4f4e8ce89fddc26e4c912393e9f6a71dc188c19fa7115cfba70000000000000009"), value: nil},
		{key: common.Hex2Bytes("4c740000000000000000000000009083ea3756bde6ee6f27a6e996806fbd37f6f093030000000000484f05497bf2b3f14a9d4f4e8ce89fddc26e4c912393e9f6a71dc188c19fa7115cfba70000000000000008"), value: nil},
		{key: common.Hex2Bytes("4c7490890809c654f11d6e72a28fa60149770a0d11ec6c92319d6ceb2bb0a4ea1a15010000000000484f05497bf2b3f14a9d4f4e8ce89fddc26e4c912393e9f6a71dc188c19fa7115cfba7000000000000000a"), value: nil},
		{key: common.Hex2Bytes("4c74ddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef010000000000484f05497bf2b3f14a9d4f4e8ce89fddc26e4c912393e9f6a71dc188c19fa7115cfba70000000000000007"), value: nil},
		{key: common.Hex2Bytes("4c74ddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef010000000000484f05497bf2b3f14a9d4f4e8ce89fddc26e4c912393e9f6a71dc188c19fa7115cfba70000000000000008"), value: nil},
		{key: common.Hex2Bytes("4c74ddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef010000000000484f05497bf2b3f14a9d4f4e8ce89fddc26e4c912393e9f6a71dc188c19fa7115cfba70000000000000009"), value: nil},
		{key: common.Hex2Bytes("4c720000000000484f05b6840d4c0eb562b0b1731760223d91b36edc6c160958e23e773e6058eea304580000000000000007"), value: common.Hex2Bytes("ddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef00000000000000000000000000000000000000000000000000000000000000000000000000000000000000004d5362dd18ea4ba880c829b0152b7ba371741e5900001718000003d9d4038f394e6b06123d0dc96da0e15ef6e16a799dcf0277df5cc61a78f164885776aa610fb0fe1257df78e59b000000000000000000000000000000000000000000000000502eca46e3971c71")},
		{key: common.Hex2Bytes("4c720000000000484f05b6840d4c0eb562b0b1731760223d91b36edc6c160958e23e773e6058eea304580000000000000008"), value: common.Hex2Bytes("ddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef00000000000000000000000000000000000000000000000000000000000000000000000000000000000000009083ea3756bde6ee6f27a6e996806fbd37f6f09300001718000003d9d4038f394e6b06123d0dc96da0e15ef6e16a799dcf0277df5cc61a78f164885776aa610fb0fe1257df78e59b00000000000000000000000000000000000000000000000321d3e6c4e3e71c71")},
		{key: common.Hex2Bytes("4c720000000000484f05b6840d4c0eb562b0b1731760223d91b36edc6c160958e23e773e6058eea304580000000000000009"), value: common.Hex2Bytes("ddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef0000000000000000000000009083ea3756bde6ee6f27a6e996806fbd37f6f09300000000000000000000000004d02149058cc8c8d0cf5f6fd1dc5394a80d737100001718000003d9d4038f394e6b06123d0dc96da0e15ef6e16a799dcf0277df5cc61a78f164885776aa610fb0fe1257df78e59b0000000000000000000000000000000000000000000000000439248ee41549c0")},
		{key: common.Hex2Bytes("4c720000000000484f05b6840d4c0eb562b0b1731760223d91b36edc6c160958e23e773e6058eea30458000000000000000a"), value: common.Hex2Bytes("90890809c654f11d6e72a28fa60149770a0d11ec6c92319d6ceb2bb0a4ea1a1500000000000000000000000004d02149058cc8c8d0cf5f6fd1dc5394a80d7371000000000000000000000000000000000000000000000000000000000000000100001718000003d9d4038f394e6b06123d0dc96da0e15ef6e16a799dcf0277df9083ea3756bde6ee6f27a6e996806fbd37f6f0930000000000000000000000000000000000000000000000000000000000000000")},
		{key: common.Hex2Bytes("4c740000000000000000000000000000000000000000000000000000000000000000020000000000484f05b6840d4c0eb562b0b1731760223d91b36edc6c160958e23e773e6058eea304580000000000000007"), value: common.Hex2Bytes("03")},
		{key: common.Hex2Bytes("4c740000000000000000000000000000000000000000000000000000000000000000020000000000484f05b6840d4c0eb562b0b1731760223d91b36edc6c160958e23e773e6058eea304580000000000000008"), value: common.Hex2Bytes("03")},
		{key: common.Hex2Bytes("4c740000000000000000000000000000000000000000000000000000000000000001030000000000484f05b6840d4c0eb562b0b1731760223d91b36edc6c160958e23e773e6058eea30458000000000000000a"), value: common.Hex2Bytes("03")},
		{key: common.Hex2Bytes("4c7400000000000000000000000004d02149058cc8c8d0cf5f6fd1dc5394a80d7371020000000000484f05b6840d4c0eb562b0b1731760223d91b36edc6c160958e23e773e6058eea30458000000000000000a"), value: common.Hex2Bytes("03")},
		{key: common.Hex2Bytes("4c7400000000000000000000000004d02149058cc8c8d0cf5f6fd1dc5394a80d7371030000000000484f05b6840d4c0eb562b0b1731760223d91b36edc6c160958e23e773e6058eea304580000000000000009"), value: common.Hex2Bytes("03")},
		{key: common.Hex2Bytes("4c740000000000000000000000004d5362dd18ea4ba880c829b0152b7ba371741e59030000000000484f05b6840d4c0eb562b0b1731760223d91b36edc6c160958e23e773e6058eea304580000000000000007"), value: common.Hex2Bytes("03")},
		{key: common.Hex2Bytes("4c740000000000000000000000005cc61a78f164885776aa610fb0fe1257df78e59b000000000000484f05b6840d4c0eb562b0b1731760223d91b36edc6c160958e23e773e6058eea304580000000000000007"), value: common.Hex2Bytes("03")},
		{key: common.Hex2Bytes("4c740000000000000000000000005cc61a78f164885776aa610fb0fe1257df78e59b000000000000484f05b6840d4c0eb562b0b1731760223d91b36edc6c160958e23e773e6058eea304580000000000000008"), value: common.Hex2Bytes("03")},
		{key: common.Hex2Bytes("4c740000000000000000000000005cc61a78f164885776aa610fb0fe1257df78e59b000000000000484f05b6840d4c0eb562b0b1731760223d91b36edc6c160958e23e773e6058eea304580000000000000009"), value: common.Hex2Bytes("03")},
		{key: common.Hex2Bytes("4c740000000000000000000000009083ea3756bde6ee6f27a6e996806fbd37f6f093000000000000484f05b6840d4c0eb562b0b1731760223d91b36edc6c160958e23e773e6058eea30458000000000000000a"), value: common.Hex2Bytes("03")},
		{key: common.Hex2Bytes("4c740000000000000000000000009083ea3756bde6ee6f27a6e996806fbd37f6f093020000000000484f05b6840d4c0eb562b0b1731760223d91b36edc6c160958e23e773e6058eea304580000000000000009"), value: common.Hex2Bytes("03")},
		{key: common.Hex2Bytes("4c740000000000000000000000009083ea3756bde6ee6f27a6e996806fbd37f6f093030000000000484f05b6840d4c0eb562b0b1731760223d91b36edc6c160958e23e773e6058eea304580000000000000008"), value: common.Hex2Bytes("03")},
		{key: common.Hex2Bytes("4c7490890809c654f11d6e72a28fa60149770a0d11ec6c92319d6ceb2bb0a4ea1a15010000000000484f05b6840d4c0eb562b0b1731760223d91b36edc6c160958e23e773e6058eea30458000000000000000a"), value: common.Hex2Bytes("03")},
		{key: common.Hex2Bytes("4c74ddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef010000000000484f05b6840d4c0eb562b0b1731760223d91b36edc6c160958e23e773e6058eea304580000000000000007"), value: common.Hex2Bytes("03")},
		{key: common.Hex2Bytes("4c74ddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef010000000000484f05b6840d4c0eb562b0b1731760223d91b36edc6c160958e23e773e6058eea304580000000000000008"), value: common.Hex2Bytes("03")},
		{key: common.Hex2Bytes("4c74ddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef010000000000484f05b6840d4c0eb562b0b1731760223d91b36edc6c160958e23e773e6058eea304580000000000000009"), value: common.Hex2Bytes("03")},
		{key: common.Hex2Bytes("78497bf2b3f14a9d4f4e8ce89fddc26e4c912393e9f6a71dc188c19fa7115cfba7"), value: common.Hex2Bytes("e783453462a000000000000000000000000000000000000000000000000000000000000000008001")},
		{key: common.Hex2Bytes("78b6840d4c0eb562b0b1731760223d91b36edc6c160958e23e773e6058eea30458"), value: common.Hex2Bytes("e783484f05a000001718000003d4d3955bf592e12fb80a60574fa4b18bd5805b4c010d75e86d0202")},
	}

	diff2 := []kvEntry{
		{key: common.Hex2Bytes("4c72000000000049431bc511216e8bff6c347014ed695cb308f2792d07fd30079f12286768716b5bfacb0000000000000004"), value: nil},
		{key: common.Hex2Bytes("4c72000000000049431bc511216e8bff6c347014ed695cb308f2792d07fd30079f12286768716b5bfacb0000000000000005"), value: nil},
		{key: common.Hex2Bytes("4c74000000000000000000000000000000000000000000000000000000000000000103000000000049431bc511216e8bff6c347014ed695cb308f2792d07fd30079f12286768716b5bfacb0000000000000005"), value: nil},
		{key: common.Hex2Bytes("4c740000000000000000000000001325625ae81846e80ac9d0b8113f31e1f8b479a802000000000049431bc511216e8bff6c347014ed695cb308f2792d07fd30079f12286768716b5bfacb0000000000000005"), value: nil},
		{key: common.Hex2Bytes("4c740000000000000000000000001325625ae81846e80ac9d0b8113f31e1f8b479a803000000000049431bc511216e8bff6c347014ed695cb308f2792d07fd30079f12286768716b5bfacb0000000000000004"), value: nil},
		{key: common.Hex2Bytes("4c740000000000000000000000005cc61a78f164885776aa610fb0fe1257df78e59b00000000000049431bc511216e8bff6c347014ed695cb308f2792d07fd30079f12286768716b5bfacb0000000000000004"), value: nil},
		{key: common.Hex2Bytes("4c740000000000000000000000009083ea3756bde6ee6f27a6e996806fbd37f6f09300000000000049431bc511216e8bff6c347014ed695cb308f2792d07fd30079f12286768716b5bfacb0000000000000005"), value: nil},
		{key: common.Hex2Bytes("4c740000000000000000000000009083ea3756bde6ee6f27a6e996806fbd37f6f09302000000000049431bc511216e8bff6c347014ed695cb308f2792d07fd30079f12286768716b5bfacb0000000000000004"), value: nil},
		{key: common.Hex2Bytes("4c7490890809c654f11d6e72a28fa60149770a0d11ec6c92319d6ceb2bb0a4ea1a1501000000000049431bc511216e8bff6c347014ed695cb308f2792d07fd30079f12286768716b5bfacb0000000000000005"), value: nil},
		{key: common.Hex2Bytes("4c74ddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef01000000000049431bc511216e8bff6c347014ed695cb308f2792d07fd30079f12286768716b5bfacb0000000000000004"), value: nil},
		{key: common.Hex2Bytes("4c72000000000049431b3aeede91740093cb8feb1296a34cf70d86d2f802cff860edd798978e94a405340000000000000004"), value: common.Hex2Bytes("ddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef0000000000000000000000009083ea3756bde6ee6f27a6e996806fbd37f6f0930000000000000000000000001325625ae81846e80ac9d0b8113f31e1f8b479a80000179e00000c49ff37137ee31a4020047bbff39016c3df896319816f2cfe565cc61a78f164885776aa610fb0fe1257df78e59b00000000000000000000000000000000000000000000000f5e7017a20b72fcb3")},
		{key: common.Hex2Bytes("4c72000000000049431b3aeede91740093cb8feb1296a34cf70d86d2f802cff860edd798978e94a405340000000000000005"), value: common.Hex2Bytes("90890809c654f11d6e72a28fa60149770a0d11ec6c92319d6ceb2bb0a4ea1a150000000000000000000000001325625ae81846e80ac9d0b8113f31e1f8b479a800000000000000000000000000000000000000000000000000000000000000010000179e00000c49ff37137ee31a4020047bbff39016c3df896319816f2cfe569083ea3756bde6ee6f27a6e996806fbd37f6f0930000000000000000000000000000000000000000000000000000000000000000")},
		{key: common.Hex2Bytes("4c74000000000000000000000000000000000000000000000000000000000000000103000000000049431b3aeede91740093cb8feb1296a34cf70d86d2f802cff860edd798978e94a405340000000000000005"), value: common.Hex2Bytes("03")},
		{key: common.Hex2Bytes("4c740000000000000000000000001325625ae81846e80ac9d0b8113f31e1f8b479a802000000000049431b3aeede91740093cb8feb1296a34cf70d86d2f802cff860edd798978e94a405340000000000000005"), value: common.Hex2Bytes("03")},
		{key: common.Hex2Bytes("4c740000000000000000000000001325625ae81846e80ac9d0b8113f31e1f8b479a803000000000049431b3aeede91740093cb8feb1296a34cf70d86d2f802cff860edd798978e94a405340000000000000004"), value: common.Hex2Bytes("03")},
		{key: common.Hex2Bytes("4c740000000000000000000000005cc61a78f164885776aa610fb0fe1257df78e59b00000000000049431b3aeede91740093cb8feb1296a34cf70d86d2f802cff860edd798978e94a405340000000000000004"), value: common.Hex2Bytes("03")},
		{key: common.Hex2Bytes("4c740000000000000000000000009083ea3756bde6ee6f27a6e996806fbd37f6f09300000000000049431b3aeede91740093cb8feb1296a34cf70d86d2f802cff860edd798978e94a405340000000000000005"), value: common.Hex2Bytes("03")},
		{key: common.Hex2Bytes("4c740000000000000000000000009083ea3756bde6ee6f27a6e996806fbd37f6f09302000000000049431b3aeede91740093cb8feb1296a34cf70d86d2f802cff860edd798978e94a405340000000000000004"), value: common.Hex2Bytes("03")},
		{key: common.Hex2Bytes("4c7490890809c654f11d6e72a28fa60149770a0d11ec6c92319d6ceb2bb0a4ea1a1501000000000049431b3aeede91740093cb8feb1296a34cf70d86d2f802cff860edd798978e94a405340000000000000005"), value: common.Hex2Bytes("03")},
		{key: common.Hex2Bytes("4c74ddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef01000000000049431b3aeede91740093cb8feb1296a34cf70d86d2f802cff860edd798978e94a405340000000000000004"), value: common.Hex2Bytes("03")},
		{key: common.Hex2Bytes("783aeede91740093cb8feb1296a34cf70d86d2f802cff860edd798978e94a40534"), value: common.Hex2Bytes("e78349431ba00000179e00000c464d756a7614d0ca067fcb37ee4452004bf308c9df561e85e88001")},
		{key: common.Hex2Bytes("78c511216e8bff6c347014ed695cb308f2792d07fd30079f12286768716b5bfacb"), value: common.Hex2Bytes("e78344dd6ca000000000000000000000000000000000000000000000000000000000000000008080")},
	}

	diff := []kvEntry{}
	blockIdx := s.GetLatestBlockIndex()
	if blockIdx >= fixTxBlock1 {
		diff = append(diff, diff1...)
	}
	if blockIdx >= fixTxBlock2 {
		diff = append(diff, diff2...)
	}

	for _, kv := range diff {
		var err error
		if kv.value != nil {
			err = s.mainDB.Put(kv.key, kv.value)
		} else {
			err = s.mainDB.Delete(kv.key)
		}
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Store) recoverHeadsStorage() error {
	s.loadEpochStore(s.GetEpoch())
	es := s.getEpochStore(s.GetEpoch())
	it := es.table.Heads.NewIterator(nil, nil)
	defer it.Release()
	heads := make(hash.EventsSet)
	for it.Next() {
		// note: key may be empty if DB was committed before being migrated, which is possible between migration steps
		if len(it.Key()) > 0 {
			heads.Add(hash.BytesToEvent(it.Key()))
		}
		_ = es.table.Heads.Delete(it.Key())
	}
	es.SetHeads(concurrent.WrapEventsSet(heads))
	es.FlushHeads()
	return nil
}

func (s *Store) recoverLastEventsStorage() error {
	s.loadEpochStore(s.GetEpoch())
	es := s.getEpochStore(s.GetEpoch())
	it := es.table.LastEvents.NewIterator(nil, nil)
	defer it.Release()
	lasts := make(map[idx.ValidatorID]hash.Event)
	for it.Next() {
		// note: key may be empty if DB was committed before being migrated, which is possible between migration steps
		if len(it.Key()) > 0 {
			lasts[idx.BytesToValidatorID(it.Key())] = hash.BytesToEvent(it.Value())
		}
		_ = es.table.LastEvents.Delete(it.Key())
	}
	es.SetLastEvents(concurrent.WrapValidatorEventsSet(lasts))
	es.FlushLastEvents()
	return nil
}

type ValidatorBlockStateV0 struct {
	Cheater          bool
	LastEvent        hash.Event
	Uptime           inter.Timestamp
	LastOnlineTime   inter.Timestamp
	LastGasPowerLeft inter.GasPowerLeft
	LastBlock        idx.Block
	DirtyGasRefund   uint64
	Originated       *big.Int
}

type BlockStateV0 struct {
	LastBlock          iblockproc.BlockCtx
	FinalizedStateRoot hash.Hash

	EpochGas      uint64
	EpochCheaters lachesis.Cheaters

	ValidatorStates       []ValidatorBlockStateV0
	NextValidatorProfiles iblockproc.ValidatorProfiles

	DirtyRules opera.Rules

	AdvanceEpochs idx.Epoch
}

type BlockEpochStateV0 struct {
	BlockState *BlockStateV0
	EpochState *iblockproc.EpochStateV0
}

func (s *Store) convertBlockEpochStateV0(oldEBS *BlockEpochStateV0) BlockEpochState {
	oldES := oldEBS.EpochState
	oldBS := oldEBS.BlockState

	newValidatorState := make([]iblockproc.ValidatorBlockState, len(oldBS.ValidatorStates))
	cheatersWritten := 0
	for i, vs := range oldBS.ValidatorStates {
		lastEvent := &inter.Event{}
		if vs.LastEvent != hash.ZeroEvent {
			lastEvent = s.GetEvent(vs.LastEvent)
		}
		newValidatorState[i] = iblockproc.ValidatorBlockState{
			LastEvent: iblockproc.EventInfo{
				ID:           vs.LastEvent,
				GasPowerLeft: lastEvent.GasPowerLeft(),
				Time:         lastEvent.MedianTime(),
			},
			Uptime:           vs.Uptime,
			LastOnlineTime:   vs.LastOnlineTime,
			LastGasPowerLeft: vs.LastGasPowerLeft,
			LastBlock:        vs.LastBlock,
			DirtyGasRefund:   vs.DirtyGasRefund,
			Originated:       vs.Originated,
		}
		if vs.Cheater {
			cheatersWritten++
		}
	}

	newBS := &iblockproc.BlockState{
		LastBlock:             oldBS.LastBlock,
		FinalizedStateRoot:    oldBS.FinalizedStateRoot,
		EpochGas:              oldBS.EpochGas,
		EpochCheaters:         oldBS.EpochCheaters,
		CheatersWritten:       uint32(cheatersWritten),
		ValidatorStates:       newValidatorState,
		NextValidatorProfiles: oldBS.NextValidatorProfiles,
		DirtyRules:            &oldBS.DirtyRules,
		AdvanceEpochs:         oldBS.AdvanceEpochs,
	}
	if oldES.Rules.String() == oldBS.DirtyRules.String() {
		newBS.DirtyRules = nil
	}

	newEs := &iblockproc.EpochState{
		Epoch:             oldES.Epoch,
		EpochStart:        oldES.EpochStart,
		PrevEpochStart:    oldES.PrevEpochStart,
		EpochStateRoot:    oldES.EpochStateRoot,
		Validators:        oldES.Validators,
		ValidatorStates:   make([]iblockproc.ValidatorEpochState, len(oldES.ValidatorStates)),
		ValidatorProfiles: oldES.ValidatorProfiles,
		Rules:             oldES.Rules,
	}
	for i, v := range oldES.ValidatorStates {
		newEs.ValidatorStates[i].GasRefund = v.GasRefund
		newEs.ValidatorStates[i].PrevEpochEvent.ID = v.PrevEpochEvent
		lastEvent := &inter.Event{}
		if v.PrevEpochEvent != hash.ZeroEvent {
			lastEvent = s.GetEvent(v.PrevEpochEvent)
		}
		newEs.ValidatorStates[i].PrevEpochEvent.Time = lastEvent.MedianTime()
		newEs.ValidatorStates[i].PrevEpochEvent.GasPowerLeft = lastEvent.GasPowerLeft()
	}

	return BlockEpochState{
		BlockState: newBS,
		EpochState: newEs,
	}
}

func (s *Store) recoverBlockState() error {
	// current block state
	v0, ok := s.rlp.Get(s.table.BlockEpochState, []byte(sKey), &BlockEpochStateV0{}).(*BlockEpochStateV0)
	if !ok {
		return errors.New("epoch state reading failed: genesis not applied")
	}
	v1 := s.convertBlockEpochStateV0(v0)
	s.SetBlockEpochState(*v1.BlockState, *v1.EpochState)
	s.FlushBlockEpochState()

	// history block state
	for epoch := idx.Epoch(1); epoch <= v0.EpochState.Epoch; epoch++ {
		v, ok := s.rlp.Get(s.table.BlockEpochStateHistory, epoch.Bytes(), &BlockEpochStateV0{}).(*BlockEpochStateV0)
		if !ok {
			continue
		}
		v1 = s.convertBlockEpochStateV0(v)
		s.SetHistoryBlockEpochState(epoch, *v1.BlockState, *v1.EpochState)
	}

	return nil
}

func (s *Store) recoverLlrState() error {
	v1, ok := s.rlp.Get(s.table.BlockEpochState, []byte(sKey), &BlockEpochState{}).(*BlockEpochState)
	if !ok {
		return errors.New("epoch state reading failed: genesis not applied")
	}

	epoch := v1.EpochState.Epoch + 1
	block := v1.BlockState.LastBlock.Idx + 1

	s.setLlrState(LlrState{
		LowestEpochToDecide: epoch,
		LowestEpochToFill:   epoch,
		LowestBlockToDecide: block,
		LowestBlockToFill:   block,
	})
	s.FlushLlrState()
	return nil
}

func (s *Store) eraseSfcApiTable() error {
	sfcapiTable := table.New(s.mainDB, []byte("S"))
	it := sfcapiTable.NewIterator(nil, nil)
	defer it.Release()
	i := 0
	for it.Next() {
		err := sfcapiTable.Delete(it.Key())
		if err != nil {
			return err
		}
		i++
		if i%1000 == 0 && s.IsCommitNeeded() {
			err := s.Commit()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Store) eraseGossipAsyncDB() error {
	asyncDB, err := s.dbs.OpenDB("gossip-async")
	if err != nil {
		return fmt.Errorf("failed to open gossip-async to drop: %v", err)
	}

	_ = asyncDB.Close()
	asyncDB.Drop()

	return nil
}

func (s *Store) eraseGenesisDB() error {
	genesisDB, err := s.dbs.OpenDB("genesis")
	if err != nil {
		return nil
	}

	_ = genesisDB.Close()
	genesisDB.Drop()
	return nil
}

func (s *Store) calculateUpgradeHeights() error {
	var prevEs *iblockproc.EpochState
	s.ForEachHistoryBlockEpochState(func(bs iblockproc.BlockState, es iblockproc.EpochState) bool {
		s.WriteUpgradeHeight(bs, es, prevEs)
		prevEs = &es
		return true
	})
	if prevEs == nil {
		// special case when no history is available
		s.WriteUpgradeHeight(s.GetBlockState(), s.GetEpochState(), nil)
	}
	return nil
}
