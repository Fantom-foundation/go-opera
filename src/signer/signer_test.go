package signer

import (
	"context"
	"io/ioutil"
	"math/big"
	"path/filepath"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/signer/core"
)

func TestSignerAPI_New(t *testing.T) {
	// Init new signer api & ui handler
	signer, ui := NewSignerAPI(tmpDirName())

	ui.inputCh <- "password_with_more_than_10_chars"
	address, err := signer.New(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	t.Log(address.Hex())
}

func TestSignerAPI_List(t *testing.T) {
	// Init new signer api & ui handler
	signer, ui := NewSignerAPI(tmpDirName())

	// create new account
	ui.inputCh <- "password_with_more_than_10_chars"
	_, err := signer.New(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	// Some time to allow changes to propagate
	time.Sleep(250 * time.Millisecond)

	// get list account
	addresses, err := signer.List(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	t.Log(addresses[0].Hex())
}

func TestSignerAPI_SignTransaction(t *testing.T) {
	// Init new signer api & ui handler
	signer, ui := NewSignerAPI(tmpDirName())

	// create new account
	ui.inputCh <- "password_with_more_than_10_chars"
	_, err := signer.New(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	// Some time to allow changes to propagate
	time.Sleep(250 * time.Millisecond)

	// get list account
	addresses, err := signer.List(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	// make transaction
	tx := testTx(common.NewMixedcaseAddress(addresses[0]))

	// sign transaction
	ui.inputCh <- "password_with_more_than_10_chars"
	result, err := signer.SignTransaction(context.Background(), tx, nil)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(result.Tx.GasPrice())
}

func TestSignerAPI_SignData(t *testing.T) {
	// Init new signer api & ui handler
	signer, ui := NewSignerAPI(tmpDirName())

	// create new account
	ui.inputCh <- "password_with_more_than_10_chars"
	_, err := signer.New(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	// Some time to allow changes to propagate
	time.Sleep(250 * time.Millisecond)

	// get list account
	addresses, err := signer.List(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	// make mixed address creator
	mixedAddress := common.NewMixedcaseAddress(addresses[0])

	// sign hash
	ui.inputCh <- "password_with_more_than_10_chars"
	signature, err := signer.SignData(context.Background(), core.TextPlain.Mime, mixedAddress, hexutil.Encode([]byte("test hash")))
	if err != nil {
		t.Fatal(err)
	}
	if signature == nil || len(signature) != 65 {
		t.Errorf("Expected 65 byte signature (got %d bytes)", len(signature))
	}

	t.Log(signature.String())
}

func TestSignerAPI_SignTypedData(t *testing.T) {
	// Init new signer api & ui handler
	signer, ui := NewSignerAPI(tmpDirName())

	// create new account
	ui.inputCh <- "password_with_more_than_10_chars"
	_, err := signer.New(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	// Some time to allow changes to propagate
	time.Sleep(250 * time.Millisecond)

	// get list account
	addresses, err := signer.List(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	// make mixed address creator
	mixedAddress := common.NewMixedcaseAddress(addresses[0])

	// sign typed data
	ui.inputCh <- "password_with_more_than_10_chars"
	signature, err := signer.SignTypedData(context.Background(), mixedAddress, typedData)
	if err != nil {
		t.Fatal(err)
	}
	if signature == nil || len(signature) != 65 {
		t.Errorf("Expected 65 byte signature (got %d bytes)", len(signature))
	}

	t.Log(signature.String())
}

func testTx(from common.MixedcaseAddress) core.SendTxArgs {
	to := common.NewMixedcaseAddress(common.HexToAddress("0x1337"))
	gas := hexutil.Uint64(21000)
	gasPrice := (hexutil.Big)(*big.NewInt(2000000000))
	value := (hexutil.Big)(*big.NewInt(1e18))
	nonce := (hexutil.Uint64)(0)
	data := hexutil.Bytes(common.Hex2Bytes("01020304050607080a"))
	tx := core.SendTxArgs{
		From:     from,
		To:       &to,
		Gas:      gas,
		GasPrice: gasPrice,
		Value:    value,
		Data:     &data,
		Nonce:    nonce}
	return tx
}

func tmpDirName() string {
	d, err := ioutil.TempDir("", "lachesis-config")
	if err != nil {
		panic(err)
	}
	d, err = filepath.EvalSymlinks(d)
	if err != nil {
		panic(err)
	}
	return d
}
