package launcher

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/Fantom-foundation/go-opera/txtrace"
)

func TestTxTracing(t *testing.T) {

	// Start test node on random ports and keep it running for another requests
	port := strconv.Itoa(trulyRandInt(10000, 65536))
	wsport := strconv.Itoa(trulyRandInt(10000, 65536))
	cliNode := exec(t,
		"--fakenet", "1/1", "--tracenode", "--port", "0", "--maxpeers", "0", "--nodiscover", "--nat", "none",
		"--ws", "--ws.port", wsport, "--http", "--http.api", "eth,web3,net,txpool,ftm,sfc,trace", "--http.port", port)

	// Wait for node to start
	endpoint := "ws://127.0.0.1:" + wsport
	waitForEndpoint(t, endpoint, 60*time.Second)

	// Deploy a smart contract from the testdata javascript file
	cliConsoleDeploy := exec(t, "attach", "--datadir", cliNode.Datadir, "--exec", "loadScript('testdata/txTracingTest.js')")
	contractAddress := string(*cliConsoleDeploy.GetOutPipeData())
	contractAddress = contractAddress[strings.Index(contractAddress, "0x") : len(contractAddress)-1]
	cliConsoleDeploy.WaitExit()

	abi := `[{"inputs":[],"name":"deploy","outputs":[{"internalType":"address","name":"","type":"address"}],"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[],"name":"getA","outputs":[{"name":"","type":"uint8"}],"payable":false,"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"uint256","name":"_a","type":"uint256"}],"name":"setA","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"uint256","name":"_a","type":"uint256"}],"name":"setInA","outputs":[],"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[],"name":"tst","outputs":[{"internalType":"address","name":"","type":"address"}],"stateMutability":"view","type":"function"}]`

	cliConsole := exec(t, "attach", "--datadir", cliNode.Datadir)
	cliConsoleOutput := *cliConsole.GetOutDataTillCursor()

	// Initialize test contract for interaction
	cliConsole.InputLine("var abi=" + abi)
	cliConsoleOutput = *cliConsole.GetOutDataTillCursor()
	cliConsole.InputLine(`var testContract = ftm.contract(abi)`)
	cliConsoleOutput = *cliConsole.GetOutDataTillCursor()
	cliConsole.InputLine("testContract = testContract.at('" + contractAddress + "')")
	cliConsoleOutput = *cliConsole.GetOutDataTillCursor()

	// Call simple contract call to check created trace
	cliConsole.InputLine("testContract.setA.sendTransaction(24, {from:ftm.accounts[1]})")
	cliConsoleOutput = *cliConsole.GetOutDataTillCursor()
	txHashCall := cliConsoleOutput[strings.Index(cliConsoleOutput, "0x") : len(cliConsoleOutput)-3]

	cliConsole.InputLine("testContract.deploy.sendTransaction({from:ftm.accounts[1]})")
	cliConsoleOutput = *cliConsole.GetOutDataTillCursor()
	txHashDeploy := cliConsoleOutput[strings.Index(cliConsoleOutput, "0x") : len(cliConsoleOutput)-3]
	time.Sleep(5000 * time.Millisecond)

	// Close node console
	cliConsole.InputLine("exit")
	cliConsole.WaitExit()

	traceResult1, err := getTrace(txHashCall, port)
	if err != nil {
		log.Fatalln(err)
	}

	traceResult2, err := getTrace(txHashDeploy, port)
	if err != nil {
		log.Fatalln(err)
	}

	// Stop test node
	cliNode.Kill()
	cliNode.WaitExit()

	// Compare results
	// Test first transaction result trace, which should be
	// just a simple call to a contract function
	require.Equal(t, traceResult1.Result[0].TraceType, "call")

	// Test second transaction result trace, which should be
	// call to a contract, which will create a new contract and
	// call a two other functions on new contract
	require.Equal(t, traceResult2.Result[0].TraceType, "call")
	require.Equal(t, traceResult2.Result[1].TraceType, "create")
	require.Equal(t, traceResult2.Result[2].TraceType, "call")
	require.Equal(t, traceResult2.Result[2].TraceType, "call")

	// Check the addresses of inner traces
	require.Equal(t, len(traceResult2.Result[0].TraceAddress), 0)
	require.Equal(t, int(traceResult2.Result[1].TraceAddress[0]), 0)
	require.Equal(t, int(traceResult2.Result[2].TraceAddress[0]), 1)
	require.Equal(t, int(traceResult2.Result[3].TraceAddress[0]), 2)
}

func getTrace(txHash string, nodePort string) (response, error) {

	jsonStr := []byte(`{"method":"trace_transaction","params":["` + txHash + `"],"id":1,"jsonrpc":"2.0"}`)
	resp, err := http.Post("http://127.0.0.1:"+nodePort, "application/json", bytes.NewBuffer(jsonStr))
	if err != nil {
		log.Fatalln(err)
	}

	//We Read the response body on the line below.
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	//Convert the body to type string
	var res response
	err = json.Unmarshal(body, &res)
	return res, err
}

type response struct {
	Jsonrpc string
	Id      int64
	Result  []txtrace.ActionTrace
}
