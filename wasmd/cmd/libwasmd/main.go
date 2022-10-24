package main

import "C"

import (
	"context"
	"encoding/json"
	"unsafe"

	"github.com/CosmWasm/wasmd/app"
	"github.com/CosmWasm/wasmd/x/wasm/types"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/pflag"
)

type Wallet struct {
	clientCtx     client.Context
	accountNumber uint64
	sequence      uint64
	gasAdjustment float64
	note          string
	timeoutHeight uint64
	gasStr        string
}

var wallets map[C.int]Wallet
var walletCounter C.int

func (c *Wallet) initCmdFlags() *pflag.FlagSet {
	flagSet := pflag.NewFlagSet("mock", pflag.PanicOnError)
	flagSet.Uint64(flags.FlagAccountNumber, c.accountNumber, "")
	flagSet.Uint64(flags.FlagSequence, c.sequence, "")
	flagSet.Float64(flags.FlagGasAdjustment, c.gasAdjustment, "")
	flagSet.String(flags.FlagNote, c.note, "")
	flagSet.Uint64(flags.FlagTimeoutHeight, c.timeoutHeight, "")
	flagSet.String(flags.FlagGas, c.gasStr, "")
	flagSet.String(flags.FlagGasPrices, "0.25umlg", "")
	return flagSet
}

func (c *Wallet) BroadcastTx(msg sdk.Msg) (string, error) {
	flagSet := c.initCmdFlags()
	txf := tx.NewFactoryCLI(c.clientCtx, flagSet)
	// prepare sets up accNum and seqNum properly
	txf, err := txf.Prepare(c.clientCtx)
	if err != nil {
		return "", err
	}
	_, adjusted, err := tx.CalculateGas(c.clientCtx, txf, msg)
	txf = txf.WithGas(adjusted)
	if err != nil {
		return "", err
	}
	utx, err := tx.BuildUnsignedTx(txf, msg)
	if err != nil {
		return "", err
	}
	utx.SetFeeGranter(c.clientCtx.GetFeeGranterAddress())
	err = tx.Sign(txf, c.clientCtx.GetFromName(), utx, true)
	if err != nil {
		return "", err
	}

	txBytes, err := c.clientCtx.TxConfig.TxEncoder()(utx.GetTx())
	if err != nil {
		return "", err
	}

	res, err := c.clientCtx.BroadcastTx(txBytes)
	if err != nil {
		return "", err
	}
	return res.String(), nil
}

//export initWallet
func initWallet(_chainId *C.char, _nodeUri *C.char) (C.int, *C.char) {
	chainId := C.GoString(_chainId)
	nodeUri := C.GoString(_nodeUri)
	ctx := NewContext(chainId, nodeUri)
	client, err := client.NewClientFromNode(nodeUri)
	if err != nil {
		return -1, C.CString(err.Error())
	}
	ctx = ctx.WithClient(client)
	kb := keyring.NewInMemory()
	ctx.Keyring = kb

	w := Wallet{
		clientCtx: ctx,
		// accountNumber and sequence must be set to 0 initially so that factory.Prepare can set it properly
		accountNumber: 0,
		sequence:      0,
		gasAdjustment: 1.2,
		note:          "",
		timeoutHeight: 100000000000000,
		gasStr:        "",
	}
	wallets[walletCounter] = w
	walletCounter += 1
	return C.int(walletCounter - 1), nil
}

//export addKeyRandom
func addKeyRandom(walletId C.int, _uid *C.char) (*C.char, *C.char) {
	uid := C.GoString(_uid)
	// default to english
	c, exists := wallets[walletId]
	if !exists {
		return C.CString(""), C.CString("invalid wallet id")
	}
	_, m, err := c.clientCtx.Keyring.NewMnemonic(uid, keyring.Language(1), "m/44'/118'/0'/0/0", "", hd.Secp256k1)
	if err != nil {
		return C.CString(""), C.CString(err.Error())
	} else {
		return C.CString(m), nil
	}
}

//export addKeyMnemonic
func addKeyMnemonic(walletId C.int, _uid *C.char, _mnemonic *C.char) *C.char {
	uid := C.GoString(_uid)
	mnemonic := C.GoString(_mnemonic)
	c, exists := wallets[walletId]
	if !exists {
		return C.CString("invalid wallet id")
	}
	_, err := c.clientCtx.Keyring.NewAccount(uid, mnemonic, "", "m/44'/118'/0'/0/0", hd.Secp256k1)
	if err != nil {
		return C.CString(err.Error())
	} else {
		return nil
	}
}

//export getKey
func getKey(walletId C.int, _uid *C.char) (*C.char, *C.char) {
	uid := C.GoString(_uid)
	c, exists := wallets[walletId]
	if !exists {
		return C.CString(""), C.CString("invalid wallet id")
	}
	info, err := c.clientCtx.Keyring.Key(uid)
	if err != nil {
		return C.CString(""), C.CString(err.Error())
	} else {
		return C.CString(info.GetAddress().String()), nil
	}
}

//export txWasmStore
func txWasmStore(walletId C.int, _sender_uid *C.char, wasmBytecodeData *C.char, wasmBytecodeLen C.int) (*C.char, *C.char) {
	sender_uid := C.GoString(_sender_uid)
	wasmBytecode := C.GoBytes(unsafe.Pointer(wasmBytecodeData), wasmBytecodeLen)
	c, exists := wallets[walletId]
	if !exists {
		return C.CString(""), C.CString("invalid wallet id")
	}
	sender_info, err := c.clientCtx.Keyring.Key(sender_uid)
	if err != nil {
		return C.CString(""), C.CString(err.Error())
	}
	c.clientCtx.FromAddress = sender_info.GetAddress()
	c.clientCtx.FromName = sender_uid
	msg := types.MsgStoreCode{
		Sender:       sender_info.GetAddress().String(),
		WASMByteCode: wasmBytecode,
		// for now, we only support AllowEverybody
		InstantiatePermission: &types.AllowEverybody,
	}
	res, err := c.BroadcastTx(&msg)
	if err != nil {
		return C.CString(""), C.CString(err.Error())
	} else {
		return C.CString(res), nil
	}
}

//export txWasmInstatitate
func txWasmInstatitate(walletId C.int, _sender_uid *C.char, codeId uint64, _label *C.char, iMsgData *C.char, iMsgLen C.int, fundsUmlg int64) (*C.char, *C.char) {
	sender_uid := C.GoString(_sender_uid)
	label := C.GoString(_label)
	iMsg := C.GoBytes(unsafe.Pointer(iMsgData), iMsgLen)
	c, exists := wallets[walletId]
	if !exists {
		return C.CString(""), C.CString("invalid wallet id")
	}
	funds := []sdk.Coin{{Denom: "umlg", Amount: sdk.NewInt(fundsUmlg)}}
	sender_info, err := c.clientCtx.Keyring.Key(sender_uid)
	if err != nil {
		return C.CString(""), C.CString(err.Error())
	}
	c.clientCtx.FromAddress = sender_info.GetAddress()
	c.clientCtx.FromName = sender_uid
	msg := types.MsgInstantiateContract{
		Sender: sender_info.GetAddress().String(),
		// for now, no-admin
		Admin:  "",
		CodeID: codeId,
		Label:  label,
		Msg:    iMsg,
		Funds:  funds,
	}
	res, err := c.BroadcastTx(&msg)
	if err != nil {
		return C.CString(""), C.CString(err.Error())
	} else {
		return C.CString(res), nil
	}
}

//export txWasmExecute
func txWasmExecute(walletId C.int, _sender_uid *C.char, _contract *C.char, _execMsg *C.char, fundsUmlg int64) (*C.char, *C.char) {
	sender_uid := C.GoString(_sender_uid)
	contract := C.GoString(_contract)
	execMsg := C.GoString(_execMsg)

	c, exists := wallets[walletId]
	if !exists {
		return C.CString(""), C.CString("invalid wallet id")
	}
	funds := []sdk.Coin{{Denom: "umlg", Amount: sdk.NewInt(fundsUmlg)}}
	sender_info, err := c.clientCtx.Keyring.Key(sender_uid)
	if err != nil {
		return C.CString(""), C.CString(err.Error())
	}
	c.clientCtx.FromAddress = sender_info.GetAddress()
	c.clientCtx.FromName = sender_uid
	msg := types.MsgExecuteContract{
		Sender:   sender_info.GetAddress().String(),
		Contract: contract,
		Funds:    funds,
		Msg:      []byte(execMsg),
	}
	res, err := c.BroadcastTx(&msg)
	if err != nil {
		return C.CString(""), C.CString(err.Error())
	} else {
		return C.CString(res), nil
	}
}

//export queryContractStateSmart
func queryContractStateSmart(walletId C.int, _contract *C.char, _queryMsg *C.char) (*C.char, *C.char) {
	contract := C.GoString(_contract)
	queryMsg := C.GoString(_queryMsg)
	c, exists := wallets[walletId]
	if !exists {
		return C.CString(""), C.CString("invalid wallet id")
	}
	_, err := sdk.AccAddressFromBech32(contract)
	if err != nil {
		return C.CString(""), C.CString(err.Error())
	}
	if queryMsg == "" {
		return C.CString(""), C.CString("query data must not be empty")
	}
	queryData := []byte(queryMsg)
	if !json.Valid(queryData) {
		return C.CString(""), C.CString("query data must be json")
	}
	queryClient := types.NewQueryClient(c.clientCtx)
	res, err := queryClient.SmartContractState(
		context.Background(),
		&types.QuerySmartContractStateRequest{
			Address:   contract,
			QueryData: queryData,
		},
	)
	if err != nil {
		return C.CString(""), C.CString(err.Error())
	}
	return C.CString(res.String()), nil
}

//export queryContractStateRaw
func queryContractStateRaw(walletId C.int, _contract *C.char, _queryMsg *C.char) (*C.char, *C.char) {
	contract := C.GoString(_contract)
	queryMsg := C.GoString(_queryMsg)
	c, exists := wallets[walletId]
	if !exists {
		return C.CString(""), C.CString("invalid wallet id")
	}
	_, err := sdk.AccAddressFromBech32(contract)
	if err != nil {
		return C.CString(""), C.CString(err.Error())
	}
	if queryMsg == "" {
		return C.CString(""), C.CString("query data must not be empty")
	}
	queryData := []byte(queryMsg)
	queryClient := types.NewQueryClient(c.clientCtx)
	res, err := queryClient.RawContractState(
		context.Background(),
		&types.QueryRawContractStateRequest{
			Address:   contract,
			QueryData: queryData,
		},
	)
	if err != nil {
		return C.CString(""), C.CString(err.Error())
	}
	return C.CString(res.String()), nil
}

//export queryContractStateAll
func queryContractStateAll(walletId C.int, _contract *C.char) (*C.char, *C.char) {
	contract := C.GoString(_contract)
	c, exists := wallets[walletId]
	if !exists {
		return C.CString(""), C.CString("invalid wallet id")
	}
	_, err := sdk.AccAddressFromBech32(contract)
	if err != nil {
		return C.CString(""), C.CString(err.Error())
	}
	queryClient := types.NewQueryClient(c.clientCtx)
	res, err := queryClient.AllContractState(
		context.Background(),
		&types.QueryAllContractStateRequest{
			Address: contract,
		},
	)
	if err != nil {
		return C.CString(""), C.CString(err.Error())
	}
	return C.CString(res.String()), nil
}

//export libwasmdInit
func libwasmdInit() {
	wallets = map[C.int]Wallet{}
	walletCounter = 0
	cfg := sdk.GetConfig()
	cfg.SetBech32PrefixForAccount(app.Bech32PrefixAccAddr, app.Bech32PrefixAccPub)
	cfg.SetBech32PrefixForValidator(app.Bech32PrefixValAddr, app.Bech32PrefixValPub)
	cfg.SetBech32PrefixForConsensusNode(app.Bech32PrefixConsAddr, app.Bech32PrefixConsPub)
	cfg.SetAddressVerifier(wasmtypes.VerifyAddressLen())
	cfg.Seal()
}

func main() {

}
