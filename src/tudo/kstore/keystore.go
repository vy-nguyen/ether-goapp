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
	"errors"
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
func NewKeyStore(keydir string, scryptN, scryptP int) KStoreIface {
	storage := NewSqlKeyStore(scryptN, scryptP)
	kstore := &KStore{
		Storage: storage,
	}
	storage.SetKeyStoreRef(kstore)
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
			wallet = NewWallet(ks, acct.OwnerUuid)
			ks.wallets[acct.OwnerUuid] = wallet
		}
		key = nil
		acctRec := NewAccount(acct.Account, acct.OwnerUuid)
		if acct.PassKey != "" {
			var err error

			key, err = getDecryptedKey(acct, "")
			if err == nil {
				wallet.unlock = true
			}
		}
		wallet.Add(acct, key, acctRec)
	}
}

/**
 * GetStorageIf
 * ------------
 */
func (ks *KStore) GetStorageIf() KsInterface {
	return ks.Storage
}

/**
 * AddWallet
 * ---------
 */
func (ks *KStore) AddWallet(wallet *Wallet) {
	ownerUuid := wallet.OwnerUuid.String()

	ks.mu.Lock()
	if ks.wallets[ownerUuid] == nil {
		ks.wallets[wallet.OwnerUuid.String()] = wallet
	}
	ks.mu.Unlock()
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
	acctKey, _ := ks.getAccountKey(acct)
	return acctKey
}

func (ks *KStore) getAccountKey(a accounts.Account) (*AccountKey, *Wallet) {
	ks.mu.RLock()
	defer ks.mu.RUnlock()

	for _, wallet := range ks.wallets {
		acctKey := wallet.Find(a)
		if acctKey != nil {
			return acctKey, wallet
		}
	}
	return nil, nil
}

/**
 * Delete
 * ------
 */
func (ks *KStore) Delete(a accounts.Account, passpharse string) error {
	acctKey, wallet := ks.getAccountKey(a)
	if acctKey == nil {
		return accounts.ErrUnknownAccount
	}
	orm := ks.Storage.GetOrm()
	sql := "DELETE FROM account_key WHERE account = ?"
	res, err := orm.Raw(sql, acctKey.Account).Exec()

	if err == nil {
		num, _ := res.RowsAffected()
		wallet.Remove(a)

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
	acctKey, _ := ks.getAccountKey(a)
	if acctKey != nil && acctKey.Key != nil {
		return crypto.Sign(hash, acctKey.Key.PrivateKey)
	}
	return nil, accounts.ErrUnknownAccount
}

/**
 * SignTx
 * ------
 */
func (ks *KStore) SignTx(a accounts.Account, tx *types.Transaction,
	chainId *big.Int) (*types.Transaction, error) {
	acctKey, _ := ks.getAccountKey(a)

	if acctKey != nil && acctKey.Key != nil {
		privKey := acctKey.Key.PrivateKey
		if chainId != nil {
			return types.SignTx(tx, types.NewEIP155Signer(chainId), privKey)
		}
		return types.SignTx(tx, types.HomesteadSigner{}, privKey)
	}
	return nil, accounts.ErrUnknownAccount
}

/**
 * LogTx
 * -----
 */
func (ks *KStore) LogTx(tx *types.Transaction) error {
	var signer types.Signer = types.FrontierSigner{}

	if tx.Protected() {
		signer = types.NewEIP155Signer(tx.ChainId())
	}
	from, _ := types.Sender(signer, tx)
	to := tx.To()
	if to == nil {
		to = &from
	}
	peerUuid := "Anonymous"
	ownerUuid := "Anonymous"
	if fromAcct := ks.GetAccountKey(from); fromAcct != nil {
		ownerUuid = fromAcct.OwnerUuid
	}
	if toAcct := ks.GetAccountKey(*to); toAcct != nil {
		peerUuid = toAcct.OwnerUuid
	}
	trans := models.Transaction{
		FromUuid: ownerUuid,
		ToUuid:   peerUuid,
		FromAcct: from.Hex(),
		ToAcct:   to.Hex(),
		TxHash:   tx.Hash().Hex(),
	}
	orm := ks.Storage.GetOrm()
	_, err := orm.Insert(&trans)
	return err
}

/**
 * SignHashWithPassphrase
 * ----------------------
 */
func (ks *KStore) SignHashWithPassphrase(a accounts.Account, passphrase string,
	hash []byte) ([]byte, error) {
	acctKey, _ := ks.getAccountKey(a)
	if acctKey == nil {
		return nil, accounts.ErrUnknownAccount
	}
	key, err := getDecryptedKey(acctKey.AccountKey, passphrase)
	if err != nil {
		return nil, err
	}
	return crypto.Sign(hash, key.PrivateKey)
}

/**
 * SignTxWithPassphrase
 * --------------------
 */
func (ks *KStore) SignTxWithPassphrase(a accounts.Account, passphrase string,
	tx *types.Transaction, chainId *big.Int) (*types.Transaction, error) {
	acctKey, _ := ks.getAccountKey(a)
	if acctKey == nil {
		return nil, accounts.ErrUnknownAccount
	}
	key, err := getDecryptedKey(acctKey.AccountKey, passphrase)
	if key != nil {
		privKey := acctKey.Key.PrivateKey
		if chainId != nil {
			return types.SignTx(tx, types.NewEIP155Signer(chainId), privKey)
		}
		return types.SignTx(tx, types.HomesteadSigner{}, privKey)
	}
	return nil, err
}

/**
 * Unlock
 * ------
 */
func (ks *KStore) Unlock(a accounts.Account, passphrase string) error {
	fmt.Printf("Key store unlock account %v, pass %s\n", a, passphrase)
	return ks.TimedUnlock(a, passphrase, 0)
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
func (ks *KStore) TimedUnlock(a accounts.Account, passphrase string,
	timeout time.Duration) error {
	acctKey, _ := ks.getAccountKey(a)
	if acctKey != nil {
		key, err := getDecryptedKey(acctKey.AccountKey, "")
		fmt.Printf("Got key %v, timeout %v\n", key, timeout)
		if err != nil {
			return err
		}
		acctKey.Key = key
		go ks.expire(acctKey, timeout)
	}
	return nil
}

func (ks *KStore) expire(acctKey *AccountKey, timeout time.Duration) {
	t := time.NewTimer(timeout)
	defer t.Stop()
	select {
	case <-acctKey.abort:
		// Exit out
	case <-t.C:
		// Timeout, encrypt back the key.
	}
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
	key, account, err := storeNewKey(ks.Storage, crand.Reader, "", passphrase)
	if err != nil {
		return accounts.Account{}, err
	}
	_, err = ks.Storage.StoreAccount(key, "annon", "normal", nil, nil)
	return account, err
}

func (ks *KStore) NewAccountOwner(ownerUuid, walletUuid,
	name, passphrase, actType string) (*accounts.Account, *models.Account, error) {
	owner := uuid.Parse(ownerUuid)
	if owner == nil {
		owner = uuid.NewRandom()
		ownerUuid = owner.String()
	}
	wallet := uuid.Parse(walletUuid)
	if wallet == nil {
		wallet = uuid.NewRandom()
	}
	key, account, err := storeNewKey(ks.Storage, crand.Reader, ownerUuid, passphrase)
	if err != nil {
		return nil, nil, err
	}
	model, err := ks.Storage.StoreAccount(key, name, actType, &owner, &wallet)
	return &account, model, err
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
	fmt.Printf("Import json key %v, pass %s %s\n", keyJson, passphase, newPassphase)
	return acct, nil
}

/**
 * ImportECDSA
 * -----------
 */
func (ks *KStore) ImportECDSA(priv *ecdsa.PrivateKey,
	passphrase string) (accounts.Account, error) {
	key := newKeyFromECDSA(priv)
	acct := accounts.Account{
		Address: key.Address,
		URL:     NewURL(key.Id.String()),
	}
	_, wallet := ks.getAccountKey(acct)
	if wallet != nil {
		return accounts.Account{}, fmt.Errorf("account already exists")
	}
	ownerUuid := acct.URL.Path
	if err := ks.Storage.StoreKey(ownerUuid, key, passphrase); err != nil {
		return accounts.Account{}, err
	}
	wallet = NewWallet(ks, ownerUuid)
	wallet.Add(&models.AccountKey{
		Account:   key.Address.Hex(),
		OwnerUuid: acct.URL.Path,
		PassKey:   passphrase,
		PrivKey:   hex.EncodeToString(crypto.FromECDSA(priv)),
	}, key, &acct)

	ks.AddWallet(wallet)
	// Send event to update account manager
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

func getDecryptedKey(acctKey *models.AccountKey, passwd string) (*keystore.Key, error) {
	s := acctKey.Account
	if s[0] == '0' {
		s = s[1:]
	}
	if s[0] == 'x' || s[0] == 'X' {
		s = s[1:]
	}
	addr, err := hex.DecodeString(s)
	if err != nil {
		return nil, err
	}
	privKey, err := crypto.HexToECDSA(acctKey.PrivKey)
	if err != nil {
		return nil, err
	}
	return &keystore.Key{
		Id:         uuid.Parse(acctKey.OwnerUuid),
		Address:    common.BytesToAddress(addr),
		PrivateKey: privKey,
	}, nil
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
	ownerUuid, auth string) (*keystore.Key, accounts.Account, error) {
	key, err := newKey(rand)
	if err != nil {
		return nil, accounts.Account{}, err
	}
	owner := key.Id
	if ownerUuid != "" {
		if owner = uuid.Parse(ownerUuid); owner != nil {
			key.Id = owner
		}
	}
	ownerStr := owner.String()
	a := accounts.Account{
		Address: key.Address,
		URL: accounts.URL{
			Scheme: "sql",
			Path:   ownerStr,
		},
	}
	if err := ks.StoreKey(ownerStr, key, auth); err != nil {
		return nil, a, err
	}
	return key, a, err
}

/**
 * Base KeyStore
 */
func (ks *BaseKeyStore) GetOrm() orm.Ormer {
	return ks.ormHandler
}

func (ks *BaseKeyStore) JoinPath(filename string) string {
	return filename
}

func (ks *BaseKeyStore) SetKeyStoreRef(kstore *KStore) {
	ks.kstore = kstore
}

/**
 * SQL based keystore.
 */
func NewSqlKeyStore(scryptN, scryptP int) *SqlKeyStore {
	return &SqlKeyStore{
		BaseKeyStore: BaseKeyStore{scryptN, scryptP, nil, orm.NewOrm()},
	}
}

/**
 * StoreKey
 * --------
 */
func (ks *SqlKeyStore) StoreKey(filename string, key *keystore.Key, auth string) error {
	return ks.StoreKeyUuid(key, key.Id, auth)
}

/**
 * GetKey
 * ------
 */
func (ks *SqlKeyStore) GetKey(addr common.Address,
	path, auth string) (*keystore.Key, error) {
	return ks.GetKeyUuid(addr, uuid.Parse(path), auth)
}

/**
 * GetAccount
 * ----------
 */
func (ks *SqlKeyStore) GetAccount(addr common.Address) ([]models.Account, error) {
	sql := fmt.Sprintf(
		"SELECT * from account WHERE account=\"%s\"", addr.Hex())

	return ks.getAccountQuery(sql)
}

/**
 * GetUserAccount
 * --------------
 */
func (ks *SqlKeyStore) GetUserAccount(ownerUuid uuid.UUID) ([]models.Account, error) {
	sql := fmt.Sprintf(
		"SELECT * from account WHERE owner_uuid=\"%s\"", ownerUuid.String())

	return ks.getAccountQuery(sql)
}

/**
 * GetWallet
 * ---------
 */
func (ks *SqlKeyStore) GetWallet(walletUuid uuid.UUID) ([]models.Account, error) {
	sql := fmt.Sprintf(
		"SELECT * from account WHERE wallet_uuid=\"%s\"", walletUuid.String())

	return ks.getAccountQuery(sql)
}

/**
 * GetTransaction
 * --------------
 */
func (ks *SqlKeyStore) GetTransaction(addr *common.Address, owner *uuid.UUID,
	from bool, offset, limit int) ([]models.Transaction, error) {
	var sql string
	acct := "from_acct"
	uuid := "from_uuid"

	if from == false {
		acct = "to_acct"
		uuid = "to_uuid"
	}
	if addr != nil && owner != nil {
		sql = fmt.Sprintf(
			"SELECT * from transaction where %s=\"%s\" AND %s=\"%s\"",
			acct, addr.Hex(), uuid, owner.String())
	} else if addr != nil {
		sql = fmt.Sprintf("SELECT * from transaction where %s=\"%s\"", acct, addr.Hex())
	} else if owner != nil {
		sql = fmt.Sprintf(
			"SELECT * from transaction where %s=\"%s\"",
			uuid, owner.String())
	} else {
		return nil, errors.New("Invalid arguments")
	}
	if limit != 0 {
		sql = fmt.Sprintf("%s LIMIT %d OFFSET %d", sql, limit, offset)
	}
	return ks.getTransQuery(sql)
}

/**
 * getAccountQuery
 * ---------------
 */
func (ks *SqlKeyStore) getAccountQuery(sql string) ([]models.Account, error) {
	var results []models.Account
	orm := ks.GetOrm()

	orm.Raw(sql).QueryRows(&results)
	if len(results) > 0 {
		return results, nil
	}
	return nil, errors.New("No account record found")
}

/**
 * getTransQuery
 * -------------
 */
func (ks *SqlKeyStore) getTransQuery(sql string) ([]models.Transaction, error) {
	var results []models.Transaction
	orm := ks.GetOrm()

	orm.Raw(sql).QueryRows(&results)
	if len(results) > 0 {
		return results, nil
	}
	return nil, errors.New("No transaction record found")
}

/**
 * UpdateAccount
 * -------------
 */
func (ks *SqlKeyStore) UpdateAccount(addr common.Address,
	name, actType string, ownerUuid, walletUuid uuid.UUID) error {
	sql := fmt.Sprintf(
		"SELECT * from account WHERE account=\"%s\" and owner_uuid=\"%s\"",
		addr.Hex(), ownerUuid.String())

	results, err := ks.getAccountQuery(sql)
	if err != nil {
		acctRec := &models.Account{
			OwnerUuid:  ownerUuid.String(),
			WalletUuid: walletUuid.String(),
			PublicName: name,
			Account:    addr.Hex(),
			Type:       actType,
		}
		_, err := ks.GetOrm().Insert(acctRec)
		return err
	}
	obj := &results[0]
	if walletUuid != nil {
		obj.WalletUuid = walletUuid.String()
	}
	obj.PublicName = name
	obj.Type = actType

	_, err = ks.ormHandler.Update(obj)
	return err
}

/**
 * GetKeyUuid
 * ----------
 */
func (ks *SqlKeyStore) GetKeyUuid(addr common.Address,
	owner uuid.UUID, auth string) (*keystore.Key, error) {
	var results []models.AccountKey

	orm := ks.GetOrm()
	sql := fmt.Sprintf("SELECT * FROM account_key WHERE Account = %s", addr.Hex())

	orm.Raw(sql).QueryRows(&results)
	if len(results) > 0 {
		acct := &results[0]
		return getDecryptedKey(acct, auth)
	}
	return nil, accounts.ErrUnknownAccount
}

/**
 * StoreAccountKey
 * ---------------
 */
func (ks *SqlKeyStore) StoreAccount(key *keystore.Key, name, actType string,
	ownerUuid *uuid.UUID, walletUuid *uuid.UUID) (*models.Account, error) {
	if walletUuid == nil {
		uid := uuid.NewRandom()
		walletUuid = &uid
	}
	if ownerUuid == nil {
		ownerUuid = &key.Id
	}
	if name == "" {
		name = "Anonymous"
	}
	orm := ks.ormHandler
	acctRec := &models.Account{
		OwnerUuid:  ownerUuid.String(),
		WalletUuid: walletUuid.String(),
		PublicName: name,
		Account:    key.Address.Hex(),
		Type:       actType,
	}
	_, err := orm.Insert(acctRec)
	return acctRec, err
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
