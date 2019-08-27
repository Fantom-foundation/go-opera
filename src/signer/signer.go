package signer

import (
	"path/filepath"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core"
	"github.com/ethereum/go-ethereum/signer/fourbyte"
	"github.com/ethereum/go-ethereum/signer/storage"
)

// UIHandler wrapper for go-ethereum/signer/core.UIClientAPI
type UIHandler struct {
	core.CommandlineUI

	inputCh chan string
}

// OnInputRequired is invoked when clef requires user input, for example master password or
// pin-code for unlocking hardware wallets
func (ui *UIHandler) OnInputRequired(info core.UserInputRequest) (core.UserInputResponse, error) {
	input := <-ui.inputCh
	return core.UserInputResponse{Text: input}, nil
}

// ApproveNewAccount prompt the user for confirmation to create new Account, and reveal to caller
func (ui *UIHandler) ApproveNewAccount(request *core.NewAccountRequest) (core.NewAccountResponse, error) {
	return core.NewAccountResponse{true}, nil
}

// ApproveListing prompt the user for confirmation to list accounts
// the list of accounts to list can be modified by the UI
func (ui *UIHandler) ApproveListing(request *core.ListRequest) (core.ListResponse, error) {
	return core.ListResponse{request.Accounts}, nil
}

// ApproveTx prompt the user for confirmation to request to sign Transaction
func (ui *UIHandler) ApproveTx(request *core.SignTxRequest) (core.SignTxResponse, error) {
	return core.SignTxResponse{request.Transaction, true}, nil
}

// ApproveSignData prompt the user for confirmation to request to sign data
func (ui *UIHandler) ApproveSignData(request *core.SignDataRequest) (core.SignDataResponse, error) {
	return core.SignDataResponse{true}, nil
}

// ShowError displays error message to user
func (ui *UIHandler) ShowError(message string) {
	return
}

// ShowInfo displays info message to user
func (ui *UIHandler) ShowInfo(message string) {
	return
}

// NewSignerAPI return SignerAPI & UIHandler
func NewSignerAPI(configDir string) (*core.SignerAPI, *UIHandler) {
	db, err := fourbyte.NewWithFile("lachesis-signer")
	if err != nil {
		panic(err.Error())
	}

	ui := &UIHandler{
		inputCh: make(chan string, 20),
	}
	am := core.StartClefAccountManager(filepath.Join(configDir, "keystore"), true, true, "")

	vaultLocation := filepath.Join(configDir, common.Bytes2Hex(crypto.Keccak256([]byte("vault"), nil)[:10]))
	pwkey := crypto.Keccak256([]byte("credentials"), nil)

	pwStorage := storage.NewAESEncryptedStorage(filepath.Join(vaultLocation, "credentials.json"), pwkey)

	// TODO: change chainID to own
	api := core.NewSignerAPI(am, 1337, true, ui, db, true, pwStorage)

	return api, ui
}
