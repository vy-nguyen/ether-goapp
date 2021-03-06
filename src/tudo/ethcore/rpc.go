/*
 *--------1---------2---------3---------4---------5---------6---------7---------8--------
 * Copyright (c) 2018 by Vy Nguyen
 * BSD License
 *
 * @author vynguyen
 */
package ethcore

import (
	"context"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/eth"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/pborman/uuid"
	"tudo/models"
)

type TudoNodeAPI struct {
	node *TudoNode
}

func NewTudoNodeAPI(n *TudoNode) *TudoNodeAPI {
	return &TudoNodeAPI{node: n}
}

/**
 * UpdateAccount
 * -------------
 * curl -X POST -H "Content-Type: application/json" --data '{"jsonrpc": "2.0",
 *    "method": "tudo_updateAccount", "params": [ "abc", "def", ... ], "id": "foo"}'
 *    localhost:8545
 */
func (api *TudoNodeAPI) UpdateAccount(address, name,
	password, actType, ownerUuid, walletUuid string) map[string]interface{} {
	ks := api.node.kstore.GetStorageIf()

	out := make(map[string]interface{})
	if !common.IsHexAddress(address) {
		out["error"] = fmt.Sprintf("Invaid address %s", address)
		return out
	}
	addr := common.HexToAddress(address)
	wallet := uuid.Parse(walletUuid)
	owner := uuid.Parse(ownerUuid)

	if owner == nil {
		out["error"] = fmt.Sprintf("Invalid owner uuid %s", ownerUuid)
	}
	err := ks.UpdateAccount(addr, name, actType, owner, wallet)
	if err != nil {
		out["error"] = err.Error()
	}
	out["address"] = addr.Hex()
	out["ownerUuid"] = owner.String()
	return out
}

/**
 * NewAccount
 * ----------
 */
func (api *TudoNodeAPI) NewAccount(ownerUuid, walletUuid,
	name, password, actType string) map[string]interface{} {
	out := make(map[string]interface{})

	kstore := api.node.kstore
	acct, model, err := kstore.NewAccountOwner(ownerUuid,
		walletUuid, name, password, actType)
	if err != nil {
		out["error"] = err.Error()
		out["ownerUuid"] = ownerUuid
		out["walletUuid"] = walletUuid
	} else {
		out["address"] = acct.Address.Hex()
		out["ownerUuid"] = model.OwnerUuid
		out["walletUuid"] = model.WalletUuid
	}
	return out
}

/**
 * GetAccount
 * ----------
 */
func (api *TudoNodeAPI) GetAccount(address string) map[string]interface{} {
	out := make(map[string]interface{})

	if !common.IsHexAddress(address) {
		out["error"] = fmt.Sprintf("Invaid address %s", address)
		return out
	}
	ks := api.node.kstore.GetStorageIf()
	results, err := ks.GetAccount(common.HexToAddress(address))
	if err != nil {
		out["error"] = err.Error()
	} else {
		out["account"] = results
	}
	return out
}

/**
 * GetUserAccount
 * --------------
 */
func (api *TudoNodeAPI) GetUserAccount(ownerUuid string) map[string]interface{} {
	out := make(map[string]interface{})
	owner := uuid.Parse(ownerUuid)
	if owner == nil {
		out["error"] = fmt.Sprintf("Invalid owner uuid %s", ownerUuid)
		return out
	}
	ks := api.node.kstore.GetStorageIf()
	results, err := ks.GetUserAccount(owner)
	if err != nil {
		out["error"] = err.Error()
	} else {
		out["account"] = results
	}
	return out
}

/**
 * GetWallet
 * ---------
 */
func (api *TudoNodeAPI) GetWallet(walletUuid string) map[string]interface{} {
	out := make(map[string]interface{})
	wallet := uuid.Parse(walletUuid)
	if wallet == nil {
		out["error"] = fmt.Sprintf("Invalid wallet uuid %s", walletUuid)
		return out
	}
	ks := api.node.kstore.GetStorageIf()
	results, err := ks.GetWallet(wallet)
	if err != nil {
		out["error"] = err.Error()
	} else {
		out["account"] = results
	}
	return out
}

/**
 * ListUserTrans
 * -------------
 * @param userUuid - uuid recorded in mysql
 * @param fromArg - true to find transactions sent *from* userUuid, false is for tx
 *     received by userUuid.
 * @param startArg, limitArg - start + limit entries read from mysql.
 */
func (api *TudoNodeAPI) ListUserTrans(ctx context.Context,
	userUuid, fromArg, startArg, limitArg string) map[string]interface{} {

	from, start, limit := parseFromStartLimitArg(fromArg, startArg, limitArg)
	out := make(map[string]interface{})
	user := uuid.Parse(userUuid)

	if user == nil {
		out["error"] = fmt.Sprintf("Invalid user uuid %s", userUuid)
		return out
	}
	ks := api.node.kstore.GetStorageIf()
	results, err := ks.GetTransaction(nil, &user, from, start, limit)
	if err != nil {
		out["error"] = err.Error()
		return out
	}
	api.getDetailTx(ctx, out, results)
	return out
}

func (api *TudoNodeAPI) getDetailTx(ctx context.Context, out map[string]interface{},
	results []models.Transaction) ([]*RPCTransaction, []map[string]interface{}) {

	txDetail := make([]*RPCTransaction, len(results))
	txBlocks := make([]map[string]interface{}, len(results))

	eth := api.node.GetEthereum()
	bcDb := eth.ChainDb()
	bcApi := eth.BcPublicApi

	for i, t := range results {
		hash := common.HexToHash(t.TxHash)
		tx, blockHash, blockNo, index := core.GetTransaction(bcDb, hash)
		if tx != nil {
			txDetail[i] = newRPCTransaction(eth, ctx, tx, blockHash, blockNo, index)
			block, _ := bcApi.GetBlockByHash(ctx, blockHash, false)
			txBlocks[i] = block
		}
	}
	if out != nil {
		out["transaction"] = results
		out["transBChain"] = txDetail
		out["transBlocks"] = txBlocks
	}
	return txDetail, txBlocks
}

/**
 * ListAccountTrans
 * ----------------
 */
func (api *TudoNodeAPI) ListAccountTrans(ctx context.Context, address string,
	fromArg, startArg, limitArg string) map[string]interface{} {
	from, start, limit := parseFromStartLimitArg(fromArg, startArg, limitArg)
	out := make(map[string]interface{})
	if !common.IsHexAddress(address) {
		out["error"] = fmt.Sprintf("Invaid address %s", address)
		return out
	}
	ks := api.node.kstore.GetStorageIf()
	acct := common.HexToAddress(address)
	results, err := ks.GetTransaction(&acct, nil, from, start, limit)

	if err != nil {
		out["error"] = err.Error()
	} else {
		api.getDetailTx(ctx, out, results)
	}
	return out
}

func parseFromStartLimitArg(fromArg, startArg, limitArg string) (*bool, int, int) {
	var fptr *bool
	from, err := strconv.ParseBool(fromArg)

	if err != nil {
		fptr = nil
	} else {
		fptr = &from
	}
	start, err := strconv.ParseInt(startArg, 10, 32)
	if err != nil {
		start = 0
	}
	limit, err := strconv.ParseInt(limitArg, 10, 32)
	if err != nil {
		limit = 1000
	}
	return fptr, int(start), int(limit)
}

/**
 * ListUserAcctTrans
 * -----------------
 */
func (api *TudoNodeAPI) ListUserAcctTrans(ctx context.Context, address, userUuid,
	fromArg, startArg, limitArg string) map[string]interface{} {

	from, start, limit := parseFromStartLimitArg(fromArg, startArg, limitArg)
	out := make(map[string]interface{})

	if !common.IsHexAddress(address) {
		out["error"] = fmt.Sprintf("Invaid address %s", address)
		return out
	}
	ks := api.node.kstore.GetStorageIf()
	owner := uuid.Parse(userUuid)
	acct := common.HexToAddress(address)
	results, err := ks.GetTransaction(&acct, &owner, from, start, limit)

	if err != nil {
		out["error"] = err.Error()
	} else {
		api.getDetailTx(ctx, out, results)
	}
	return out
}

func (api *TudoNodeAPI) listAcctTrans(ctx context.Context, address *common.Address,
	start, limit int, txOut *[]*RPCTransaction, blkOut *[]map[string]interface{}) {

	ks := api.node.kstore.GetStorageIf()
	results, err := ks.GetTransaction(address, nil, nil, start, limit)
	if err == nil {
		txD, blk := api.getDetailTx(ctx, nil, results)
		*txOut = append(*txOut, txD...)
		*blkOut = append(*blkOut, blk...)
	}
}

/**
 * ListAccountInfo
 * ---------------
 */
func (api *TudoNodeAPI) ListAccountInfo(ctx context.Context,
	args []string) map[string]interface{} {
	return listAccoutInternal(api, ctx, false, false, args)
}

func (api *TudoNodeAPI) ListAccountInfoAndBlock(ctx context.Context,
	args []string) map[string]interface{} {
	return listAccoutInternal(api, ctx, true, false, args)
}

func (api *TudoNodeAPI) ListAccountInfoAndTx(ctx context.Context,
	args []string) map[string]interface{} {
	return listAccoutInternal(api, ctx, false, true, args)
}

func listAccoutInternal(api *TudoNodeAPI, ctx context.Context, latest, txs bool,
	args []string) map[string]interface{} {

	out := make(map[string]interface{})
	txOut := make([]*RPCTransaction, 0)
	blkOut := make([]map[string]interface{}, 0)

	eth := api.node.GetEthereum()
	ethApi := eth.ApiBackend
	state, _, err := ethApi.StateAndHeaderByNumber(ctx, -1)

	results := make([]*AccountInfo, len(args))
	for idx, addr := range args {
		acct := common.HexToAddress(addr)
		balance := state.GetBalance(acct)
		results[idx] = &AccountInfo{
			Account: addr,
			Balance: *balance,
		}
		if txs == true {
			api.listAcctTrans(ctx, &acct, 0, 100, &txOut, &blkOut)
		}
	}
	if latest == true {
		bcApi := eth.BcPublicApi
		out["latest"], err = bcApi.GetBlockByNumber(ctx, -1, true)
	} else {
		out["latest"] = nil
	}
	if txs == true {
		out["transBChain"] = txOut
		out["transBlocks"] = blkOut
	}
	out["accounts"] = results
	out["error"] = err
	return out
}

/**
 * ListBlocks
 * ----------
 */
func (api *TudoNodeAPI) ListBlocks(ctx context.Context,
	start, cnt, txDetail string) map[string]interface{} {

	startBlk, err := strconv.ParseInt(start, 10, 64)
	if err != nil {
		startBlk = -1
	}
	count, err := strconv.ParseInt(cnt, 10, 32)
	if err != nil {
		count = 100
	}
	detail, err := strconv.ParseBool(txDetail)
	if err != nil {
		detail = false
	}
	if count <= 0 || count > 100 {
		count = 100
	}
	out := make(map[string]interface{})
	result := make([]map[string]interface{}, count)

	eth := api.node.GetEthereum()
	bcApi := eth.BcPublicApi

	block, err := bcApi.GetBlockByNumber(ctx, rpc.BlockNumber(startBlk), detail)
	if block != nil {
		count--
		result[count] = block
		for count--; count >= 0; count-- {
			startBlk--
			block, err = bcApi.GetBlockByNumber(ctx, rpc.BlockNumber(startBlk), detail)
			result[count] = block
		}
	}
	out["blocks"] = result
	out["error"] = err
	return out
}

/**
 * ListBlockHash
 * -------------
 */
func (api *TudoNodeAPI) ListBlockHash(ctx context.Context,
	hashes []string) map[string]interface{} {
	var res []interface{}

	eth := api.node.GetEthereum()
	bcApi := eth.BcPublicApi

	for _, s := range hashes {
		hash := common.HexToHash(s)
		block, _ := bcApi.GetBlockByHash(ctx, hash, true)
		if block != nil {
			res = append(res, block)
		}
	}
	out := make(map[string]interface{})
	out["blocks"] = res
	return out
}

/**
 * ListTrans
 * ---------
 */
func (api *TudoNodeAPI) ListTrans(ctx context.Context,
	trans []string) map[string]interface{} {
	var res []interface{}

	eth := api.node.GetEthereum()
	bcDb := eth.ChainDb()

	for _, s := range trans {
		hash := common.HexToHash(s)
		tx, blockHash, blockNo, index := core.GetTransaction(bcDb, hash)
		if tx == nil {
			tx = eth.ApiBackend.GetPoolTransaction(hash)
		}
		if tx != nil {
			res = append(res, newRPCTransaction(eth, ctx, tx, blockHash, blockNo, index))
		}
	}
	out := make(map[string]interface{})
	out["trans"] = res
	return out
}

func newRPCTransaction(ethApi *eth.Ethereum, ctx context.Context,
	tx *types.Transaction, blockHash common.Hash,
	blockNumber uint64, index uint64) *RPCTransaction {
	var signer types.Signer = types.FrontierSigner{}
	if tx.Protected() {
		signer = types.NewEIP155Signer(tx.ChainId())
	}
	toBalance := big.NewInt(0)
	fromBalance := big.NewInt(0)

	from, _ := types.Sender(signer, tx)
	v, r, s := tx.RawSignatureValues()

	if blockNumber != 0 {
		blkNo := (rpc.BlockNumber)(blockNumber)
		state, _, err := ethApi.ApiBackend.StateAndHeaderByNumber(ctx, blkNo)
		if state != nil && err == nil {
			toBalance = state.GetBalance(*tx.To())
			fromBalance = state.GetBalance(from)
		}
	}
	result := &RPCTransaction{
		From:        from,
		Gas:         hexutil.Uint64(tx.Gas()),
		GasPrice:    (*hexutil.Big)(tx.GasPrice()),
		Hash:        tx.Hash(),
		Input:       hexutil.Bytes(tx.Data()),
		Nonce:       hexutil.Uint64(tx.Nonce()),
		To:          tx.To(),
		Value:       (*hexutil.Big)(tx.Value()),
		V:           (*hexutil.Big)(v),
		R:           (*hexutil.Big)(r),
		S:           (*hexutil.Big)(s),
		ToBalance:   (*hexutil.Big)(toBalance),
		FromBalance: (*hexutil.Big)(fromBalance),
	}
	if blockHash != (common.Hash{}) {
		result.BlockHash = blockHash
		result.BlockNumber = (*hexutil.Big)(new(big.Int).SetUint64(blockNumber))
		result.TransactionIndex = hexutil.Uint(index)
	}
	return result
}

func (api *TudoNodeAPI) PayUserAccount(ctx context.Context, from, fromUuid, to, toUuid,
	weiAmount, text string) map[string]interface{} {

	out := make(map[string]interface{})
	ks := api.node.kstore.GetStorageIf()

	fromAddr := common.HexToAddress(from)
	fromAcct, err := ks.GetAccountOwner(fromAddr.Hex(), fromUuid)
	if err != nil || fromAcct == nil ||
		strings.Compare(fromAddr.Hex(), fromAcct.Account) != 0 {
		out["error"] = fmt.Sprintf("Invalid from account %s", from)
		return out
	}
	toAddr := common.HexToAddress(to)
	toAcct, err := ks.GetAccountOwner(to, toUuid)
	if err != nil || toAcct == nil ||
		strings.Compare(toAddr.Hex(), toAcct.Account) != 0 {
		out["error"] = fmt.Sprintf("Invalid to account %s", to)
		return out
	}
	value := new(big.Int)
	value, ok := value.SetString(weiAmount, 10)
	if !ok {
		out["error"] = fmt.Sprintf("Invalid amount %s", weiAmount)
		return out
	}
	weiVal := hexutil.Big(*value)

	eth := api.node.GetEthereum()
	txPool := eth.TxPublicPoolApi
	sendTx := txPool.NewSendTxArgs(fromAddr, &toAddr, &weiVal, nil, nil)

	txHash, err := txPool.SendTransaction(ctx, sendTx)
	if err != nil {
		out["error"] = err.Error()
	} else {
		out["txHash"] = txHash.Hex()
	}
	return out
}

/**
 * DumpAccounts
 * ------------
 */
func (api *TudoNodeAPI) DumpAccounts(ctx context.Context) map[string]interface{} {
	out := make(map[string]interface{})

	eth := api.node.GetEthereum()
	bc := eth.BlockChain()
	latest := bc.CurrentBlock()

	if latest == nil {
		fmt.Printf("No block!")
		return out
	}
	stateDb, err := bc.StateAt(latest.Root())
	accounts := stateDb.RawDump()

	out["accounts"] = accounts.Accounts
	out["error"] = err
	return out
}

/**
 * DumpTrans
 * ---------
 */
func (api *TudoNodeAPI) DumpTrans(ctx context.Context) map[string]interface{} {
	out := make(map[string]interface{})

	eth := api.node.GetEthereum()
	bc := eth.BlockChain()
	latest := bc.CurrentBlock()

	if latest == nil {
		fmt.Printf("No block!")
		return out
	}
	orm := api.node.GetOrm()
	currNo := latest.Number().Uint64()
	for i := uint64(0); i <= currNo; i++ {
		block := bc.GetBlockByNumber(i)
		if block == nil {
			fmt.Printf("Failed to fetch block %d\n", i)
			continue
		}
		txs := block.Transactions()
		if len(txs) == 0 {
			continue
		}
		for _, tx := range txs {
			LogTransaction(tx, orm)
		}
	}
	return out
}
