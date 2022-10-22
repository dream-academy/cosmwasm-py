package main

import "C"

import (
	"context"
	"encoding/json"
	"errors"

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

var wallets map[int]Wallet
var walletCounter int

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
func initWallet(chainId string, nodeUri string) (int, error) {
	ctx := NewContext(chainId, nodeUri)
	client, err := client.NewClientFromNode(nodeUri)
	if err != nil {
		return -1, err
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
	return walletCounter - 1, nil
}

//export addKeyRandom
func addKeyRandom(walletId int, uid string) (string, error) {
	// default to english
	c, exists := wallets[walletId]
	if !exists {
		return "", errors.New("invalid wallet id")
	}
	_, m, err := c.clientCtx.Keyring.NewMnemonic(uid, keyring.Language(1), "m/44'/118'/0'/0/0", "", hd.Secp256k1)
	return m, err
}

//export addKeyMnemonic
func addKeyMnemonic(walletId int, uid string, mnemonic string) error {
	c, exists := wallets[walletId]
	if !exists {
		return errors.New("invalid wallet id")
	}
	_, err := c.clientCtx.Keyring.NewAccount(uid, mnemonic, "", "m/44'/118'/0'/0/0", hd.Secp256k1)
	return err
}

//export getKey
func getKey(walletId int, uid string) (string, error) {
	c, exists := wallets[walletId]
	if !exists {
		return "", errors.New("invalid wallet id")
	}
	info, err := c.clientCtx.Keyring.Key(uid)
	if err != nil {
		return "", err
	} else {
		return info.GetAddress().String(), nil
	}
}

// export txWasmStore
func txWasmStore(walletId int, sender_uid string, wasmBytecode []byte) (string, error) {
	c, exists := wallets[walletId]
	if !exists {
		return "", errors.New("invalid wallet id")
	}
	sender_info, err := c.clientCtx.Keyring.Key(sender_uid)
	if err != nil {
		return "", err
	}
	c.clientCtx.FromAddress = sender_info.GetAddress()
	c.clientCtx.FromName = sender_uid
	msg := types.MsgStoreCode{
		Sender:       sender_info.GetAddress().String(),
		WASMByteCode: wasmBytecode,
		// for now, we only support AllowEverybody
		InstantiatePermission: &types.AllowEverybody,
	}
	return c.BroadcastTx(&msg)
}

// export txWasmInstatitate
func txWasmInstatitate(walletId int, sender_uid string, codeId uint64, label string, iMsg []byte, fundsUmlg int64) (string, error) {
	c, exists := wallets[walletId]
	if !exists {
		return "", errors.New("invalid wallet id")
	}
	funds := []sdk.Coin{{Denom: "umlg", Amount: sdk.NewInt(fundsUmlg)}}
	sender_info, err := c.clientCtx.Keyring.Key(sender_uid)
	if err != nil {
		return "", err
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
	return c.BroadcastTx(&msg)
}

//export txWasmExecute
func txWasmExecute(walletId int, sender_uid string, contract string, execMsg string, fundsUmlg int64) (string, error) {
	c, exists := wallets[walletId]
	if !exists {
		return "", errors.New("invalid wallet id")
	}
	funds := []sdk.Coin{{Denom: "umlg", Amount: sdk.NewInt(fundsUmlg)}}
	sender_info, err := c.clientCtx.Keyring.Key(sender_uid)
	if err != nil {
		return "", err
	}
	c.clientCtx.FromAddress = sender_info.GetAddress()
	c.clientCtx.FromName = sender_uid
	msg := types.MsgExecuteContract{
		Sender:   sender_info.GetAddress().String(),
		Contract: contract,
		Funds:    funds,
		Msg:      []byte(execMsg),
	}
	return c.BroadcastTx(&msg)
}

//export queryContractStateSmart
func queryContractStateSmart(walletId int, contract string, queryMsg string) (string, error) {
	c, exists := wallets[walletId]
	if !exists {
		return "", errors.New("invalid wallet id")
	}
	_, err := sdk.AccAddressFromBech32(contract)
	if err != nil {
		return "", err
	}
	if queryMsg == "" {
		return "", errors.New("query data must not be empty")
	}
	queryData := []byte(queryMsg)
	if !json.Valid(queryData) {
		return "", errors.New("query data must be json")
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
		return "", err
	}
	return res.String(), nil
}

//export cfgInit
func cfgInit() {
	cfg := sdk.GetConfig()
	cfg.SetBech32PrefixForAccount(app.Bech32PrefixAccAddr, app.Bech32PrefixAccPub)
	cfg.SetBech32PrefixForValidator(app.Bech32PrefixValAddr, app.Bech32PrefixValPub)
	cfg.SetBech32PrefixForConsensusNode(app.Bech32PrefixConsAddr, app.Bech32PrefixConsPub)
	cfg.SetAddressVerifier(wasmtypes.VerifyAddressLen())
	cfg.Seal()
}

func main() {

}
