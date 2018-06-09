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
	"math/big"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/pborman/uuid"
	"tudo/models"
)

func NewURL(path string) accounts.URL {
	return accounts.URL{
		Scheme: "sql",
		Path:   path,
	}
}

func NewWallet(ks *KStore, ownerUuid string) *Wallet {
	return &Wallet{
		unlock:    false,
		AcctMap:   make(map[string]*AccountKey),
		OwnerUuid: uuid.Parse(ownerUuid),
		KsIface:   ks,
	}
}

/**
 * URL
 * ---
 */
func (w *Wallet) URL() accounts.URL {
	return NewURL(w.OwnerUuid.String())
}

/**
 * Status
 * ------
 */
func (w *Wallet) Status() (string, error) {
	if w.unlock == true {
		return "Unlocked", nil
	}
	return "Locked", nil
}

/**
 * Open
 * ----
 * If the wallet was encrypted, open it with the passphrase.
 */
func (w *Wallet) Open(passphrase string) error {
	return nil
}

/**
 * Close
 * -----
 * Close the decrypted wallet.
 */
func (w *Wallet) Close() error {
	return nil
}

/**
 * Accounts
 * --------
 * Return the list of accounts associated with this wallet.
 */
func (w *Wallet) Accounts() []accounts.Account {
	result := make([]accounts.Account, 0, len(w.AcctMap))
	for _, actKey := range w.AcctMap {
		result = append(result, accounts.Account{
			Address: actKey.Account.Address,
			URL:     w.URL(),
		})
	}
	return result
}

/**
 * Contains
 * --------
 */
func (w *Wallet) Contains(account accounts.Account) bool {
	if w.Find(account) != nil {
		return true
	}
	return false
}

/**
 * Find
 * ----
 */
func (w *Wallet) Find(account accounts.Account) *AccountKey {
	acctKey := w.AcctMap[account.Address.Hex()]
	if acctKey == nil {
		return nil
	}
	if w.URL() != account.URL {
		if account.URL == (accounts.URL{}) {
			return acctKey
		}
		return nil
	}
	return acctKey
}

/**
 * Remove
 * ------
 * Remove account from the wallet.
 */
func (w *Wallet) Remove(account accounts.Account) {
	w.mu.Lock()
	delete(w.AcctMap, account.Address.Hex())
	w.mu.Unlock()
}

/**
 * Add
 * ---
 * Add account to the wallet.
 */
func (w *Wallet) Add(acctRec *models.AccountKey,
	key *keystore.Key, account *accounts.Account) {
	w.mu.Lock()
	w.AcctMap[account.Address.Hex()] = &AccountKey{
		AccountKey: acctRec,
		Key:        key,
		Account:    account,
		abort:      make(chan struct{}),
	}
	w.mu.Unlock()
}

/**
 * Derive
 * ------
 */
func (w *Wallet) Derive(path accounts.DerivationPath,
	pin bool) (accounts.Account, error) {
	return accounts.Account{}, nil
}

/**
 * SelfDerive
 * ----------
 */
func (w *Wallet) SelfDerive(base accounts.DerivationPath,
	chain ethereum.ChainStateReader) {
}

/**
 * SignHash
 * --------
 */
func (w *Wallet) SignHash(account accounts.Account, hash []byte) ([]byte, error) {
	if w.Contains(account) == false {
		return nil, accounts.ErrUnknownAccount
	}
	return w.KsIface.SignHash(account, hash)
}

/**
 * SignTx
 * ------
 */
func (w *Wallet) SignTx(account accounts.Account,
	tx *types.Transaction, chainID *big.Int) (*types.Transaction, error) {
	if w.Contains(account) == false {
		return nil, accounts.ErrUnknownAccount
	}
	return w.KsIface.SignTx(account, tx, chainID)
}

func (w *Wallet) LogTx(tx *types.Transaction) {
	w.KsIface.LogTx(tx)
	fmt.Printf("Log tx hash %s\n", tx.Hash())
}

/**
 * SignHashWithPassphrase
 * ----------------------
 */
func (w *Wallet) SignHashWithPassphrase(account accounts.Account,
	passphrase string, hash []byte) ([]byte, error) {
	if w.Contains(account) == false {
		return nil, accounts.ErrUnknownAccount
	}
	return w.KsIface.SignHashWithPassphrase(account, passphrase, hash)
}

/**
 * SignTxWithPassphrase
 * --------------------
 */
func (w *Wallet) SignTxWithPassphrase(account accounts.Account, passphrase string,
	tx *types.Transaction, chainID *big.Int) (*types.Transaction, error) {
	if w.Contains(account) == false {
		return nil, accounts.ErrUnknownAccount
	}
	return w.KsIface.SignTxWithPassphrase(account, passphrase, tx, chainID)
}
