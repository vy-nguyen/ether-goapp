/*
 *--------1---------2---------3---------4---------5---------6---------7---------8--------
 * Copyright (c) 2018 by Vy Nguyen
 * BSD License
 *
 * @author vynguyen
 */
package ethcore

import (
	"math/big"
	"tudo/models"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

type UpdateAccountReqt struct {
	Address    string
	Name       string
	Password   string
	OwnerUuid  string
	WalletUuid string
}

type UpdateAcountResp struct {
	Error     string
	Address   string
	OwnerUuid string
}

type ListUserTransReqt struct {
	Address  string
	UserUuid string
	From     bool
	Start    int
	Limit    int
}

type ListUserTransResp struct {
	Error       string
	Transaction []models.Transaction
}

type AccountInfo struct {
	Account string
	Balance big.Int
}

type RPCTransaction struct {
	BlockHash        common.Hash     `json:"blockHash"`
	BlockNumber      *hexutil.Big    `json:"blockNumber"`
	From             common.Address  `json:"from"`
	Gas              hexutil.Uint64  `json:"gas"`
	GasPrice         *hexutil.Big    `json:"gasPrice"`
	Hash             common.Hash     `json:"hash"`
	Input            hexutil.Bytes   `json:"input"`
	Nonce            hexutil.Uint64  `json:"nonce"`
	To               *common.Address `json:"to"`
	TransactionIndex hexutil.Uint    `json:"transactionIndex"`
	Value            *hexutil.Big    `json:"value"`
	V                *hexutil.Big    `json:"v"`
	R                *hexutil.Big    `json:"r"`
	S                *hexutil.Big    `json:"s"`
}
