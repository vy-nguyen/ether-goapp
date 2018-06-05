/*
 *--------1---------2---------3---------4---------5---------6---------7---------8--------
 * Copyright (c) 2018 by Vy Nguyen
 * BSD License
 *
 * @author vynguyen
 */
package kstore

import (
	"crypto/ecdsa"
	crand "crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"math/big"
	"time"

	"github.com/astaxie/beego/orm"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/event"
	"github.com/pborman/uuid"
	"tudo/models"
)

/**
 * NewKeyStore
 * -----------
 */
func NewKeyStore(keydir string, scryptN, scryptP int) keystore.KeyStore {
	kstore := &KStore{
		Storage: NewSqlKeyStore(scryptN, scryptP),
	}
	kstore.init()
	return kstore
}

/**
 * NewAccount
 * ----------
 */
func NewAccount(acct, ownerUuid string) *accounts.Account {
	return &accounts.Account{
		Address: common.HexToAddress(acct),
		URL:     NewURL(ownerUuid),
	}
}

func (ks *KStore) init() {
	var key *keystore.Key
	var accounts []models.AccountKey

	orm := ks.Storage.GetOrm()
	sql := "SELECT * FROM account_key LIMIT 1000 OFFSET 0"

	// execute the raw query string
	orm.Raw(sql).QueryRows(&accounts)

	ks.wallets = make(map[string]*Wallet)
	for idx, _ := range accounts {
		acct := &accounts[idx]
		wallet := ks.wallets[acct.OwnerUuid]
		if wallet == nil {
			wallet = &Wallet{
				unlock:    false,
				AcctMap:   make(map[string]*AccountKey),
				OwnerUuid: uuid.Parse(acct.OwnerUuid),
				KsIface:   ks,
			}
			ks.wallets[acct.OwnerUuid] = wallet
		}
		key = nil
		acctRec := NewAccount(acct.Account, acct.OwnerUuid)
		if acct.PassKey != "" {
			var err error

			key, err = ks.getDecryptedKey(acctRec, acct.PassKey)
			if err == nil {
				wallet.unlock = true
			}
		}
		wallet.AcctMap[acctRec.Address.Hex()] = &AccountKey{
			AccountKey: acct,
			Key:        key,
			Account:    acctRec,
			abort:      make(chan struct{}),
		}
	}
}

/**
 * SaveAccountKey
 * --------------
 */
func (ks *KStore) SaveAccountKey(key *keystore.Key, passphase string,
	ownerUuid *uuid.UUID, walletUuid *uuid.UUID) error {
	if walletUuid == nil {
		uid := uuid.NewRandom()
		walletUuid = &uid
	}
	if ownerUuid == nil {
		ownerUuid = &key.Id
	}
	orm := ks.Storage.GetOrm()
	acctRec := &models.Account{
		OwnerUuid:  ownerUuid.String(),
		WalletUuid: walletUuid.String(),
		PublicName: "Anonymous",
		Account:    key.Address.Hex(),
	}
	_, err := orm.Insert(acctRec)
	return err
}

/**
 * Wallets
 * -------
 */
func (ks *KStore) Wallets() []accounts.Wallet {
	wallets := make([]accounts.Wallet, 0, len(ks.wallets))
	for _, w := range ks.wallets {
		wallets = append(wallets, w)
	}
	return wallets
}

/**
 * Subscribe
 * ---------
 */
func (ks *KStore) Subscribe(sink chan<- accounts.WalletEvent) event.Subscription {
	ks.mu.Lock()
	defer ks.mu.Unlock()

	sub := ks.updateScope.Track(ks.updateFeed.Subscribe(sink))

	if !ks.updating {
		ks.updating = true
		go ks.updater()
	}
	return sub
}

func (ks *KStore) updater() {
	for {
		select {
		case <-ks.changes:
		case <-time.After(3 * time.Second):
		}

		ks.mu.Lock()
		if ks.updateScope.Count() == 0 {
			ks.updating = false
			ks.mu.Unlock()
			return
		}
		ks.mu.Unlock()
	}
}

/**
 * HasAddress
 * ----------
 */
func (ks *KStore) HasAddress(addr common.Address) bool {
	if ks.GetAccountKey(addr) == nil {
		return false
	}
	return true
}

/**
 * Accounts
 * --------
 */
func (ks *KStore) Accounts() []accounts.Account {
	accounts := []accounts.Account{}
	for _, wallet := range ks.wallets {
		out := wallet.Accounts()
		if len(out) > 0 {
			accounts = append(accounts, out...)
		}
	}
	fmt.Println("Get accounts from keystore, count %d", len(accounts))
	return accounts
}

/**
 * GetAccountKey
 * -------------
 */
func (ks *KStore) GetAccountKey(addr common.Address) *AccountKey {
	acct := accounts.Account{
		Address: addr,
		URL:     accounts.URL{},
	}
	return ks.getAccountKey(acct)
}

func (ks *KStore) getAccountKey(a accounts.Account) *AccountKey {
	for _, wallet := range ks.wallets {
		acctKey := wallet.Find(a)
		if acctKey != nil {
			return acctKey
		}
	}
	return nil
}

/**
 * Delete
 * ------
 */
func (ks *KStore) Delete(a accounts.Account, passpharse string) error {
	acctKey := ks.getAccountKey(a)
	if acctKey == nil {
		return accounts.ErrUnknownAccount
	}
	orm := ks.Storage.GetOrm()
	sql := "DELETE FROM account_key WHERE account = ?"
	res, err := orm.Raw(sql, acctKey.Account).Exec()

	if err == nil {
		num, _ := res.RowsAffected()
		fmt.Printf("Deleted %d rows\n", num)
	} else {
		fmt.Printf("Error returned %v\n", err)
	}
	return err
}

/**
 * SignHash
 * --------
 */
func (ks *KStore) SignHash(a accounts.Account, hash []byte) ([]byte, error) {
	return nil, nil
}

/**
 * Sign
 * ----
 */
func (ks *KStore) SignTx(a accounts.Account, tx *types.Transaction,
	chainId *big.Int) (*types.Transaction, error) {
	return nil, nil
}

/**
 * SignHashWithPassphrase
 * ----------------------
 */
func (ks *KStore) SignHashWithPassphrase(a accounts.Account, passphase string,
	hash []byte) ([]byte, error) {
	return nil, nil
}

/**
 * SignTxWithPassphrase
 * --------------------
 */
func (ks *KStore) SignTxWithPassphrase(a accounts.Account, passphase string,
	tx *types.Transaction, chainID *big.Int) (*types.Transaction, error) {
	return nil, nil
}

/**
 * Unlock
 * ------
 */
func (ks *KStore) Unlock(a accounts.Account, passphrase string) error {
	fmt.Printf("Key store unlock account %v, pass %s\n", a, passphrase)
	return nil
}

/**
 * Lock
 * ----
 */
func (ks *KStore) Lock(addr common.Address) error {
	return nil
}

/**
 * TimedUnlock
 * -----------
 */
func (ks *KStore) TimedUnlock(a accounts.Account, passphase string,
	timeout time.Duration) error {
	return nil
}

/**
 * Find
 * ----
 */
func (ks *KStore) Find(a accounts.Account) (accounts.Account, error) {
	acct := accounts.Account{}
	return acct, nil
}

/**
 * NewAccount
 * ----------
 */
func (ks *KStore) NewAccount(passphrase string) (accounts.Account, error) {
	key, account, err := storeNewKey(ks.Storage, crand.Reader, passphrase)
	if err != nil {
		return accounts.Account{}, err
	}
	err = ks.SaveAccountKey(key, passphrase, nil, nil)
	return account, err
}

/**
 * Export
 * ------
 */
func (ks *KStore) Export(a accounts.Account,
	passphase, newPassPhase string) ([]byte, error) {
	return nil, nil
}

/**
 * Import
 * ------
 */
func (ks *KStore) Import(keyJson []byte,
	passphase, newPassphase string) (accounts.Account, error) {
	acct := accounts.Account{}
	return acct, nil
}

/**
 * ImportECDSA
 * -----------
 */
func (ks *KStore) ImportECDSA(priv *ecdsa.PrivateKey,
	passphrase string) (accounts.Account, error) {
	acct := accounts.Account{}
	return acct, nil
}

/**
 * Update
 * ------
 */
func (ks *KStore) Update(a accounts.Account, passphrase, newPassphrase string) error {
	return nil
}

/**
 * ImportPreSaleKey
 * ----------------
 */
func (ks *KStore) ImportPreSaleKey(keyJSON []byte,
	passphrase string) (accounts.Account, error) {
	acct := accounts.Account{}
	return acct, nil
}

func (ks *KStore) getDecryptedKey(a *accounts.Account,
	passphrase string) (*keystore.Key, error) {
	return nil, nil
}

/**
 * newKeyFromECDSA
 * ---------------
 */
func newKeyFromECDSA(privateKeyECDSA *ecdsa.PrivateKey) *keystore.Key {
	id := uuid.NewRandom()
	key := &keystore.Key{
		Id:         id,
		Address:    crypto.PubkeyToAddress(privateKeyECDSA.PublicKey),
		PrivateKey: privateKeyECDSA,
	}
	return key
}

/**
 * newKey
 * ------
 */
func newKey(rand io.Reader) (*keystore.Key, error) {
	privKeyECDSA, err := ecdsa.GenerateKey(crypto.S256(), rand)
	if err != nil {
		return nil, err
	}
	return newKeyFromECDSA(privKeyECDSA), nil
}

/**
 * storeNewKey
 * -----------
 */
func storeNewKey(ks keystore.KeyStoreIf, rand io.Reader,
	auth string) (*keystore.Key, accounts.Account, error) {
	key, err := newKey(rand)
	if err != nil {
		return nil, accounts.Account{}, err
	}
	a := accounts.Account{
		Address: key.Address,
		URL: accounts.URL{
			Scheme: "sql",
			Path:   key.Id.String(),
		},
	}
	if err := ks.StoreKey(a.URL.Path, key, auth); err != nil {
		return nil, a, err
	}
	return key, a, err
}

/**
 * BaseKeyStore
 */
func (ks *BaseKeyStore) GetKey(addr common.Address,
	filename string, auth string) (*keystore.Key, error) {
	fmt.Println("Get key addr %v", addr)
	return nil, nil
}

func (ks *BaseKeyStore) StoreKey(filename string, k *keystore.Key, auth string) error {
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

func (ks *BaseKeyStore) GetOrm() orm.Ormer {
	return ks.ormHandler
}

/**
 * SQL based keystore.
 */
func NewSqlKeyStore(scryptN, scryptP int) *SqlKeyStore {
	return &SqlKeyStore{
		BaseKeyStore: BaseKeyStore{scryptN, scryptP, orm.NewOrm()},
	}
}

func (ks *SqlKeyStore) StoreKey(filename string, key *keystore.Key, auth string) error {
	return ks.StoreKeyUuid(key, key.Id, auth)
}

func (ks *SqlKeyStore) GetKey(addr common.Address,
	path, auth string) (*keystore.Key, error) {
	return ks.GetKeyUuid(addr, uuid.Parse(path), auth)
}

/**
 * GetKeyUuid
 * ----------
 */
func (ks *SqlKeyStore) GetKeyUuid(addr common.Address,
	owner uuid.UUID, auth string) (*keystore.Key, error) {
	var accounts []models.AccountKey

	orm := ks.GetOrm()
	sql := fmt.Sprintf("SELECT * FROM account_key WHERE Account = %s", addr.Hex())

	orm.Raw(sql).QueryRows(&accounts)
	fmt.Printf("Query returned %v\n", accounts)
	if len(accounts) > 0 {
		acct := accounts[0]
		addr, err := hex.DecodeString(acct.Account)
		if err != nil {
			return nil, err
		}
		privKey, err := crypto.HexToECDSA(acct.PrivKey)
		if err != nil {
			return nil, err
		}
		return &keystore.Key{
			Id:         owner,
			Address:    common.BytesToAddress(addr),
			PrivateKey: privKey,
		}, nil
	}
	return nil, nil
}

/**
 * StoreKeyUuid
 * ------------
 */
func (ks *SqlKeyStore) StoreKeyUuid(k *keystore.Key, owner uuid.UUID, auth string) error {
	keyRec := &models.AccountKey{
		Account:   k.Address.Hex(),
		OwnerUuid: owner.String(),
		PassKey:   auth,
		PrivKey:   hex.EncodeToString(crypto.FromECDSA(k.PrivateKey)),
	}
	_, err := ks.ormHandler.Insert(keyRec)
	return err
}
