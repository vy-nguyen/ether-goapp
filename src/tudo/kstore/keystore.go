/*
 *--------1---------2---------3---------4---------5---------6---------7---------8--------
 * Copyright (c) 2018 by Vy Nguyen
 * BSD License
 *
 * @author vynguyen
 */
package kstore

import (
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pborman/uuid"
)

func NewKeyStore(keydir string, scryptN, scryptP int) keystore.KeyStore {
	ksif := keystore.NewKeyStore(keydir, scryptN, scryptP)
	ks := ksif.GetKeyStoreObj()
	ks.Storage = &SqlKeyStore{
		BaseKeyStore: BaseKeyStore{scryptN, scryptP},
	}
	return ks
}

func (ks *BaseKeyStore) GetKey(addr common.Address,
	filename string, auth string) (*keystore.Key, error) {
	fmt.Printf("Base GetKey\n")
	return nil, nil
}

func (ks *BaseKeyStore) StoreKey(filename string, k *keystore.Key, auth string) error {
	fmt.Printf("Base store key %v, name %s\n", k, filename)
	return nil
}

func (ks *BaseKeyStore) JoinPath(filename string) string {
	return filename
}

func (ks *BaseKeyStore) GetKeyUuid(addr common.Address,
	owner uuid.UUID, auth string) (*keystore.Key, error) {
	return nil, nil
}

func (ks *BaseKeyStore) StoreKeyUuid(k *keystore.Key,
	owner uuid.UUID, auth string) error {
	return nil
}

/**
 * SQL based keystore.
 */
func (ks *SqlKeyStore) GetKeyUuid(addr common.Address,
	owner uuid.UUID, auth string) (*keystore.Key, error) {
	fmt.Printf("Get key is called %v\n", addr)
	return nil, nil
}

func (ks *SqlKeyStore) StoreKeyUuid(k *keystore.Key, owner uuid.UUID, auth string) error {
	fmt.Printf("Save key %v\n", k)
	return nil
}
