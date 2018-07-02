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
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/pborman/uuid"
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
	password, ownerUuid, walletUuid string) map[string]interface{} {
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
	err := ks.UpdateAccount(addr, name, password, owner, wallet)
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
	name, password string) map[string]interface{} {
	out := make(map[string]interface{})

	kstore := api.node.kstore
	acct, model, err := kstore.NewAccountOwner(ownerUuid, walletUuid, name, password)
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
 */
func (api *TudoNodeAPI) ListUserTrans(userUuid,
	fromArg, startArg, limitArg string) map[string]interface{} {
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
	} else {
		out["transaction"] = results
	}
	return out
}

/**
 * ListAccountTrans
 * ----------------
 */
func (api *TudoNodeAPI) ListAccountTrans(address string,
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
		out["transaction"] = results
	}
	return out
}

func parseFromStartLimitArg(fromArg, startArg, limitArg string) (bool, int, int) {
	from, err := strconv.ParseBool(fromArg)
	if err != nil {
		from = true
	}
	start, err := strconv.ParseInt(startArg, 10, 32)
	if err != nil {
		start = 0
	}
	limit, err := strconv.ParseInt(limitArg, 10, 32)
	if err != nil {
		limit = 1000
	}
	return from, int(start), int(limit)
}

/**
 * ListUserAcctTrans
 * -----------------
 */
func (api *TudoNodeAPI) ListUserAcctTrans(address, userUuid,
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
		out["transaction"] = results
	}
	return out
}

/**
 * ListAccountInfo
 * ---------------
 */
func (api *TudoNodeAPI) ListAccountInfo(ctx context.Context,
	args []string) map[string]interface{} {
	return listAccoutInternal(api, ctx, false, args)
}

func (api *TudoNodeAPI) ListAccountInfoAndBlock(ctx context.Context,
	args []string) map[string]interface{} {
	return listAccoutInternal(api, ctx, true, args)
}

func listAccoutInternal(api *TudoNodeAPI, ctx context.Context, latest bool,
	args []string) map[string]interface{} {
	out := make(map[string]interface{})
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
	}
	if latest == true {
		bcApi := eth.BcPublicApi
		out["latest"], err = bcApi.GetBlockByNumber(ctx, -1, true)
	} else {
		out["latest"] = nil
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
