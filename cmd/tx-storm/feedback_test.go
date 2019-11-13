package main

import (
	"context"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/ethclient"
)

func GetBlockExample(t *testing.T) {
	const url = "http://127.0.0.1:4002"

	client, err := ethclient.Dial(url)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	header, err := client.HeaderByNumber(context.TODO(), nil)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(header)

	block, err := client.BlockByNumber(context.TODO(), big.NewInt(29))
	if err != nil {
		t.Fatal(err)
	}

	t.Log(block)
}
