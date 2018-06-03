/*
 *--------1---------2---------3---------4---------5---------6---------7---------8--------
 * Copyright (c) 2018 by Vy Nguyen
 * BSD License
 *
 * @author vynguyen
 */
package kstore

import (
	"math/big"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/core/types"
)

/**
 * URL
 * ---
 */
func (w *Wallet) URL() accounts.URL {
	return accounts.URL{
		Scheme: "sql",
		Path:   w.OwnerUuid.String(),
	}
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
			Address: actKey.Address,
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
	if w.AcctMap[account.Address.Hex()] == nil {
		return false
	}
	if w.URL() != account.URL {
		if account.URL == (accounts.URL{}) {
			return true
		}
		return false
	}
	return true
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
