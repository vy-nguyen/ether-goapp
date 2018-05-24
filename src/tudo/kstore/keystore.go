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

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/event"
	"github.com/pborman/uuid"
)

func NewKeyStore(scryptN, scryptP int) KsInterface {
	ks := &KStore{
		StoreIf: &SqlKeyStore{
			BaseKeyStore: BaseKeyStore{scryptN, scryptP},
		},
	}
	ks.init()
	return ks
}

/**
 * Base keystore.
 */
func (ks *KStore) init() {
	ks.mu.Lock()
	defer ks.mu.Unlock()

	ks.unlocked = make(map[common.Address]*keystore.Key)
	ks.wallets = make([]accounts.Wallet, 1)
}

func (ks *KStore) Wallets() []accounts.Wallet {
	fmt.Printf("Wallets is called\n")
	return nil
}

func (ks *KStore) Subscribe(sink chan<- accounts.WalletEvent) event.Subscription {
	fmt.Printf("Subscribe is called\n")
	return nil
}

func (ks *KStore) GetKey(addr common.Address,
	owner uuid.UUID, auth string) (*keystore.Key, error) {
	return nil, nil
}

func (ks *KStore) StoreKey(k *keystore.Key, owner uuid.UUID, auth string) error {
	return nil
}

/**
 * SQL based keystore.
 */
func (ks *SqlKeyStore) GetKey(addr common.Address,
	owner uuid.UUID, auth string) (*keystore.Key, error) {
	fmt.Printf("Get key is called %v\n", addr)
	return nil, nil
}

func (ks *SqlKeyStore) StoreKey(k *keystore.Key, owner uuid.UUID, auth string) error {
	fmt.Printf("Save key %v\n", k)
	return nil
}

func (ks *SqlKeyStore) Wallets() []accounts.Wallet {
	return nil
}

func (ks *SqlKeyStore) Subscribe(sink chan<- accounts.WalletEvent) event.Subscription {
	return nil
}
