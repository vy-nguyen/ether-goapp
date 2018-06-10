package models

import ()

type Account struct {
	OwnerUuid  string `orm:"pk;size(64)"`
	WalletUuid string `orm:"index;size(64)"`
	PublicName string `orm:"size(128)"`
	Account    string `orm:"size(128)"`
}

type AccountKey struct {
	Account   string `orm:"pk;size(64)"`
	OwnerUuid string `orm:"index;size(64)"`
	PassKey   string `orm:"size(128)"`
	PrivKey   string `orm:"size(512)"`
}

type Transaction struct {
	TxHash      string `orm:"pk;size(128)"`
	OwnerUuid   string `orm:"index;size(64)"`
	PeerUuid    string `orm:"size(64)"`
	Account     string `orm:"index;size(64)"`
	PeerAccount string `orm:"index;size(64)"`
}
