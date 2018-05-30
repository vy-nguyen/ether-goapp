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
	"io"

	"github.com/astaxie/beego/orm"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/pborman/uuid"
	"tudo/models"
)

func NewKeyStore(keydir string, scryptN, scryptP int) keystore.KeyStore {
	ksif := keystore.NewKeyStore(keydir, scryptN, scryptP)
	kstore := &KStore{
		KeyStoreObj: *ksif.GetKeyStoreObj(),
	}
	kstore.kstoreIf = NewSqlKeyStore(scryptN, scryptP)
	kstore.KeyStoreObj.Storage = kstore.kstoreIf
	return kstore
}

func (ks *KStore) NewAccount(passphrase string) (accounts.Account, error) {
	key, account, err := storeNewKey(ks.KeyStoreObj.Storage, crand.Reader, passphrase)
	if err != nil {
		return accounts.Account{}, err
	}
	err = ks.SaveAccountKey(key, passphrase, nil, nil)
	return account, err
}

func (ks *KStore) SaveAccountKey(key *keystore.Key, passphase string,
	ownerUuid *uuid.UUID, walletUuid *uuid.UUID) error {
	if walletUuid == nil {
		uid := uuid.NewRandom()
		walletUuid = &uid
	}
	if ownerUuid == nil {
		ownerUuid = &key.Id
	}
	orm := ks.kstoreIf.GetOrm()
	acctRec := &models.Account{
		OwnerUuid:  ownerUuid.String(),
		WalletUuid: walletUuid.String(),
		PublicName: "Anonymous",
		Account:    key.Address.Hex(),
	}
	_, err := orm.Insert(acctRec)
	return err
}

func newKeyFromECDSA(privateKeyECDSA *ecdsa.PrivateKey) *keystore.Key {
	id := uuid.NewRandom()
	key := &keystore.Key{
		Id:         id,
		Address:    crypto.PubkeyToAddress(privateKeyECDSA.PublicKey),
		PrivateKey: privateKeyECDSA,
	}
	return key
}

func newKey(rand io.Reader) (*keystore.Key, error) {
	privKeyECDSA, err := ecdsa.GenerateKey(crypto.S256(), rand)
	if err != nil {
		return nil, err
	}
	return newKeyFromECDSA(privKeyECDSA), nil
}

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
