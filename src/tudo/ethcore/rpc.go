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
 * StoreKey
 * --------
 * curl -X POST -H "Content-Type: application/json" --data '{"jsonrpc": "2.0",
 *    "method": "tudo_storeKey", "params": [ "abc", "def" ], "id": "foo"}'
 *    localhost:8545
 */
func (api *TudoNodeAPI) StoreKey(privKey, userUuid string) map[string]interface{} {
	accman := api.node.AccountManager()
	kstore := api.node.kstore

	fmt.Printf("Account manager %v, key %v\n", accman, kstore)
	out := make(map[string]interface{})
	out["privKey"] = privKey
	out["userUuid"] = userUuid
	out["greeting"] = "Hello string"
	return out
}

func (api *TudoNodeAPI) NewAccount(userUuid,
	name, password string) map[string]interface{} {
	out := make(map[string]interface{})

	return out
}

func (api *TudoNodeAPI) ListUserTrans(userUuid string) map[string]interface{} {
	out := make(map[string]interface{})

	return out
}

func (api *TudoNodeAPI) ListAccountTrans(account string) map[string]interface{} {
	out := make(map[string]interface{})

	return out
}
