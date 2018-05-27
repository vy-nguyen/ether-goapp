/*
 *--------1---------2---------3---------4---------5---------6---------7---------8--------
 * Copyright (c) 2018 by Vy Nguyen
 * BSD License
 *
 * @author vynguyen
 */
package kstore

import (
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pborman/uuid"
)

type KsInterface interface {
	keystore.KeyStoreIf

	GetKeyUuid(addr common.Address, owner uuid.UUID, auth string) (*keystore.Key, error)
	StoreKeyUuid(k *keystore.Key, owner uuid.UUID, auth string) error
}

type KStore struct {
	keystore.KeyStoreObj
}

type BaseKeyStore struct {
	scryptN int
	scryptP int
}

type SqlKeyStore struct {
	BaseKeyStore
}
