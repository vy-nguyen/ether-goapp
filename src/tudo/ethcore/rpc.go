/*
 *--------1---------2---------3---------4---------5---------6---------7---------8--------
 * Copyright (c) 2018 by Vy Nguyen
 * BSD License
 *
 * @author vynguyen
 */
package ethcore

import (
	"fmt"
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
	password, walletUuid string) map[string]interface{} {
	accman := api.node.AccountManager()
	kstore := api.node.kstore

	fmt.Printf("Account manager %v, key %v\n", accman, kstore)
	out := make(map[string]interface{})
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
		out["address"] = "0"
		out["ownerUuid"] = ownerUuid
		out["walletUuid"] = walletUuid
	} else {
		out["error"] = ""
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
func (api *TudoNodeApi) GetAccount(address string) map[string]interface{} {
	out := make(map[string]interface{})

	return out
}

/**
 * GetUserAccount
 * --------------
 */
func (api *TudoNodeApi) GetUserAccount(ownerUuid string) map[string]interface{} {
	out := make(map[string]interface{})

	return out
}

/**
 * GetWallet
 * ---------
 */
func (api *TudoNodeApi) GetWallet(walletUuid string) map[string]interface{} {
	out := make(map[string]interface{})

	return out
}

/**
 * ListUserTrans
 * -------------
 */
func (api *TudoNodeAPI) ListUserTrans(userUuid string,
	start, end int) map[string]interface{} {
	out := make(map[string]interface{})

	return out
}

/**
 * ListAccountTrans
 * ----------------
 */
func (api *TudoNodeAPI) ListAccountTrans(account string,
	start, end int) map[string]interface{} {
	out := make(map[string]interface{})

	return out
}
