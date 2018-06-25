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
