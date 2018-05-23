/*
 *--------1---------2---------3---------4---------5---------6---------7---------8--------
 * Copyright (c) 2018 by Vy Nguyen
 * BSD License
 *
 * @author vynguyen
 */
package kstore

import (
	"sync"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/event"
	"github.com/pborman/uuid"
)

type Interface interface {
	GetKey(addr common.Address, owner uuid.UUID, auth string) (*keystore.Key, error)
	StoreKey(k *keystore.Key, owner uuid.UUID, auth string) error

	Wallets() []accounts.Wallet
	Subscribe(sink chan<- accounts.WalletEvent) event.Subscription
}

type KStore struct {
	StoreIf    Interface
	changes    chan struct{}
	unlocked   map[common.Address]*keystore.Key
	wallets    []accounts.Wallet
	updateFeed event.Feed
	updateSope event.SubscriptionScope
	updating   bool
	mu         sync.RWMutex
}

type BaseKeyStore struct {
	scryptN int
	scryptP int
}

type SqlKeyStore struct {
	BaseKeyStore
}
