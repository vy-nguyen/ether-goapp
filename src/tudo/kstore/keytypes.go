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

/**
 * KeyStore Storage specific interface
 */
type KsInterface interface {
	keystore.KeyStoreIf

	GetOrm() orm.Ormer
	SetKeyStoreRef(kstore *KStore)

	GetAccount(addr common.Address) ([]models.Account, error)
	GetAccountOwner(addr, ownerUuid string) (*models.Account, error)
	GetUserAccount(ownerUuid uuid.UUID) ([]models.Account, error)
	GetWallet(walletUuid uuid.UUID) ([]models.Account, error)

	GetTransaction(addr *common.Address, owner *uuid.UUID,
		from *bool, offset, limit int) ([]models.Transaction, error)
	GetKeyUuid(addr common.Address, owner uuid.UUID, auth string) (*keystore.Key, error)

	StoreAccount(k *keystore.Key, name, passwd string,
		ownerUuid *uuid.UUID, walletUuid *uuid.UUID) (*models.Account, error)
	StoreKeyUuid(k *keystore.Key, owner uuid.UUID, auth string) error
	UpdateAccount(addr common.Address, name, actType string,
		ownerUuid uuid.UUID, walletUuid uuid.UUID) error
}

/**
 * Main KeyStore interface
 */
type KStoreIface interface {
	keystore.KeyStore

	GetStorageIf() KsInterface
	NewAccountOwner(ownerUuid, walletUuid,
		name, passphrase, actType string) (*accounts.Account, *models.Account, error)
}

/**
 * SQL based keystore object.
 */
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
	kstore     *KStore
	ormHandler orm.Ormer
}

type SqlKeyStore struct {
	BaseKeyStore
}
