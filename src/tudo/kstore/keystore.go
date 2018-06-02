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

func (ks *KStore) init() {
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
				AcctMap:   make(map[string]AccountKey),
				OwnerUuid: uuid.Parse(acct.OwnerUuid),
				KsIface:   ks,
			}
			ks.wallets[acct.OwnerUuid] = wallet
		}
		if acct.PassKey != "" {
			wallet.unlock = true
		}
		address := common.HexToAddress(acct.Account)
		wallet.AcctMap[address.Hex()] = AccountKey{
			AccountKey: acct,
			Address:    address,
		}
	}
	fmt.Printf("Wallet %v\n", ks.wallets)
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
		fmt.Printf("Wallet %s, value %v\n", w.OwnerUuid.String(), w)
	}
	fmt.Printf("Get wallets from keystore %v\n", wallets)
	return wallets
}

/**
 * Subscribe
 * ---------
 */
func (ks *KStore) Subscribe(sink chan<- accounts.WalletEvent) event.Subscription {
	return nil
}

/**
 * HasAddress
 * ----------
 */
func (ks *KStore) HasAddress(addr common.Address) bool {
	return false
}

/**
 * Accounts
 * --------
 */
func (ks *KStore) Accounts() []accounts.Account {
	fmt.Println("Get accounts from keystore")
	return nil
}

/**
 * Delete
 * ------
 */
func (ks *KStore) Delete(a accounts.Account, passpharse string) error {
	return nil
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
func (ks *KStore) Unlock(a accounts.Account, passphase string) error {
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

func (ks *SqlKeyStore) GetKeyUuid(addr common.Address,
	owner uuid.UUID, auth string) (*keystore.Key, error) {
	return nil, nil
}

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
