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

	"github.com/ethereum/go-ethereum/common"
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
func (api *TudoNodeAPI) ListUserTrans(userUuid string,
	from bool, start, limit int) map[string]interface{} {
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
	from bool, start, limit int) map[string]interface{} {
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

/**
 * ListUserAcctTrans
 * -----------------
 */
func (api *TudoNodeAPI) ListUserAcctTrans(address, userUuid string,
	from bool, start, limit int) map[string]interface{} {
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
	out := make(map[string]interface{})
	eth := api.node.GetEthereum().ApiBackend
	state, _, err := eth.StateAndHeaderByNumber(ctx, -1)

	results := make([]*AccountInfo, len(args))
	for idx, addr := range args {
		acct := common.HexToAddress(addr)
		balance := state.GetBalance(acct)
		results[idx] = &AccountInfo{
			Account: addr,
			Balance: *balance,
		}
	}
	out["error"] = err
	out["accounts"] = results
	return out
}
