package models

import (
	"math/big"
	"time"
)

var (
	TEN       = big.NewInt(10)
	XU_UNIT   = new(big.Int).Exp(TEN, big.NewInt(14), nil)
	HAO_UNIT  = new(big.Int).Exp(TEN, big.NewInt(16), nil)
	DONG_UNIT = new(big.Int).Exp(TEN, big.NewInt(18), nil)
)

type Account struct {
	Account    string `orm:"pk;size(128)"`
	OwnerUuid  string `orm:"index;size(64)"`
	WalletUuid string `orm:"index;size(64)"`
	PublicName string `orm:"size(64)"`
	Type       string `orm:"size(64)"`
}

type AccountKey struct {
	Account   string `orm:"pk;size(64)"`
	OwnerUuid string `orm:"index;size(64)"`
	PassKey   string `orm:"size(128)"`
	PrivKey   string `orm:"size(512)"`
}

type Transaction struct {
	TxHash   string    `orm:"pk;size(128)"`
	FromUuid string    `orm:"index;size(64)"`
	ToUuid   string    `orm:"index;size(64)"`
	FromAcct string    `orm:"index;size(64)"`
	ToAcct   string    `orm:"index;size(64)"`
	XuAmount uint64    `orm:"bigint unsigned"`
	Created  time.Time `orm:"auto_now_add;type(date)"`
}
