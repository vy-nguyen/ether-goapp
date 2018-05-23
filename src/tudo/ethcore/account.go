/*
 *--------1---------2---------3---------4---------5---------6---------7---------8--------
 * Copyright (c) 2018 by Vy Nguyen
 * BSD License
 *
 * @author vynguyen
 */
package ethcore

import (
	"tudo/kstore"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/node"
)

func makeAccountManager(conf *node.Config) (accounts.Manager, error) {
	scryptN, scryptP, _, err := conf.AccountConfig()

	if err != nil {
		return nil, err
	}
	backends := []accounts.Backend{
		kstore.NewKeyStore(scryptN, scryptP),
	}
	return accounts.NewManager(backends...), nil
}
