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

	"github.com/astaxie/beego/orm"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/event"
	"github.com/pborman/uuid"
	"tudo/models"
)

type KsInterface interface {
	keystore.KeyStoreIf

	GetKeyUuid(addr common.Address, owner uuid.UUID, auth string) (*keystore.Key, error)
	StoreKeyUuid(k *keystore.Key, owner uuid.UUID, auth string) error
	GetOrm() orm.Ormer
}

type KStore struct {
	Storage     KsInterface
	changes     chan struct{}
	updateFeed  event.Feed
	updateScope event.SubscriptionScope
	updating    bool
	wallets     map[string]*Wallet
	mu          sync.RWMutex
}

/**
 * Wallet for multiple accounts.
 */
type AccountKey struct {
	*models.AccountKey
	*keystore.Key

	Account *accounts.Account
	abort   chan struct{}
}

type Wallet struct {
	AcctMap   map[string]*AccountKey
	unlock    bool
	OwnerUuid uuid.UUID
	KsIface   keystore.KeyStore
	mu        sync.Mutex
}

type BaseKeyStore struct {
	scryptN    int
	scryptP    int
	ormHandler orm.Ormer
}

type SqlKeyStore struct {
	BaseKeyStore
}
