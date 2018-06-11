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

func makeAccountManager(conf *node.Config) (AmInterface, error) {
	scryptN, scryptP, keydir, err := conf.AccountConfig()

	if keydir == "" {
		keydir, err = ioutil.TempDir("", "go-eth-keystore")
	}
	if err != nil {
		return nil, err
	}
	kstore := []keystore.KeyStore{
		kstore.NewKeyStore(keydir, scryptN, scryptP),
	}
	return NewManager(kstore...), nil
}
