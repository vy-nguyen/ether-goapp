/*
 *--------1---------2---------3---------4---------5---------6---------7---------8--------
 * Copyright (c) 2018 by Vy Nguyen
 * BSD License
 *
 * @author vynguyen
 */
package ethcore

import (
	"io/ioutil"
	"tudo/kstore"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/node"
)

func makeAccountManager(conf *node.Config,
	tdcfg *TudoConfig) (AmInterface, kstore.KStoreIface, error) {
	scryptN, scryptP, keydir, err := conf.AccountConfig()

	if keydir == "" {
		keydir, err = ioutil.TempDir("", "go-eth-keystore")
	}
	if err != nil {
		return nil, nil, err
	}
	ksIface := kstore.NewKeyStore(keydir, scryptN, scryptP)
	ksArray := []keystore.KeyStore{
		ksIface,
	}
	return NewManager(tdcfg, ksArray...), ksIface, nil
}
