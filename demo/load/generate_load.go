/*
generate some load
*/
package main

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"fmt"
	"math/big"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Fantom-foundation/go-opera/ftmclient"
	"github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

func main() {
	if len(os.Args) != 5 {
		fmt.Println(`this script requires 3 params: 
            * RPC url of node
						* number of accounts to create
						* number of transactions`)
		os.Exit(1)
	}

	url := os.Args[1]
	numAccts := os.Args[2]
	numTxs := os.Args[3]
	numNodes := os.Args[4]
	endpoints, _ := strconv.Atoi(numNodes)

	clients := make([]*ftmclient.Client, endpoints)

	for n := range clients {
		u := url
		u = fmt.Sprintf(u[:len(u)-1]+"%d", n)
		c, err := ftmclient.Dial(u)
		if err != nil {
			utils.Fatalf("failed to dial client: %v", err)
		}
		clients[n] = c
	}
	client := clients[0]
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	baseArgs := []string{"attach", "--exec"}
	opera := filepath.Join("build", "opera")
	getAccount := "ftm.accounts[0]"

	getAccountArgs := append(baseArgs, getAccount, url)
	cmd := exec.Command(opera, getAccountArgs...)
	stdout, err := cmd.Output()
	if err != nil {
		utils.Fatalf("failed to get account address: %v", err)
	}
	sanitizedAddr := string(stdout)[1 : len(stdout)-2]
	fmt.Println(" funded account: " + sanitizedAddr + "\n")
	fundedAddr := common.HexToAddress(sanitizedAddr)
	//fmt.Println(fundedAddr.Hex())

	balance, err := client.BalanceAt(ctx, fundedAddr, nil)
	if err != nil {
		utils.Fatalf("failed to get balance: %v", err)
	}
	fmt.Printf("balance of funded address: " + balance.Text(10) + "\n")
	privateKeyECDSA, err := ecdsa.GenerateKey(crypto.S256(), rand.Reader)
	if err != nil {
		utils.Fatalf("Failed to create account: %v", err)
	}

	baseAccount := crypto.PubkeyToAddress(privateKeyECDSA.PublicKey)
	baseAccountHex := baseAccount.Hex()
	fmt.Println(baseAccountHex)

	sendTx := fmt.Sprintf("ftm.sendTransaction({from: ftm.accounts[0], to: \"%s\", value: \"%d\"})", baseAccountHex, balance.Sub(balance, big.NewInt(100000000000000)))
	fmt.Println(sendTx)
	sendTxArgs := append(baseArgs, sendTx, url)
	//fmt.Println(sendTxArgs)
	cmd = exec.Command(opera, sendTxArgs...)
	stdout, err = cmd.Output()
	if err != nil {
		utils.Fatalf("Failed to execute send tx: %v", err)
	}
	if strings.Contains(string(stdout), "Error") {
		utils.Fatalf("failed to send txs: %s", string(stdout))
	}
	fmt.Print("send tx hash: " + string(stdout) + "\n")
	accountsNum, _ := strconv.Atoi(numAccts)

	time.Sleep(2 * time.Second)

	balance, err = client.BalanceAt(ctx, baseAccount, nil)
	if err != nil {
		utils.Fatalf("failed to get balance: %v", err)
	}
	fmt.Printf("balance of base account now: " + balance.Text(10) + "\n")

	prvKeys := make([]*ecdsa.PrivateKey, accountsNum)
	addrs := make([]common.Address, accountsNum)
	for i := 0; i < accountsNum; i++ {
		prvECDSA, err := ecdsa.GenerateKey(crypto.S256(), rand.Reader)
		if err != nil {
			utils.Fatalf("Failed to create account: %v", err)
		}
		prvKeys[i] = prvECDSA
		addrs[i] = crypto.PubkeyToAddress(prvECDSA.PublicKey)
	}

	txsNum, _ := strconv.Atoi(numTxs)
	txs := make([]*types.Transaction, txsNum)
	nonce := uint64(0)
	val := big.NewInt(int64(111111))
	gasLimit := uint64(610000)
	gasPrice := big.NewInt(int64(10000000000))
	chainID := big.NewInt(int64(0xfa3))

	for i := 0; i < txsNum; i++ {
		tx := types.NewTransaction(nonce+1, addrs[i%len(addrs)], val, gasLimit, gasPrice, nil)
		signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKeyECDSA)
		if err != nil {
			utils.Fatalf("Failed to sign tx: %v", err)
		}
		txs[i] = signedTx
		nonce++
	}

	count := 0
	for _, tx := range txs {
		err = client.SendTransaction(ctx, tx)
		if err != nil {
			fmt.Println(fmt.Sprintf("Failed to send tx via cli: %v", err))
			utils.Fatalf("transactions sent so far: %d", count)
		}
		count++
	}
	/*
		for i, tx := range txs {
			err = clients[i%len(clients)].SendTransaction(ctx, tx)
			if err != nil {
				fmt.Println(fmt.Sprintf("Failed to send tx via cli: %v", err))
				utils.Fatalf("transactions sent so far: %d", count)
			}
			count++
			time.Sleep(100 * time.Millisecond)
		}
	*/
	fmt.Println(fmt.Sprintf("sent %d transactions", count))

}
