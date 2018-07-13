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

	"github.com/astaxie/beego/orm"
	"github.com/ethereum/go-ethereum/core/types"
	"tudo/models"
)

func LogTransaction(tx *types.Transaction, orm orm.Ormer) error {
	var signer types.Signer = types.FrontierSigner{}
	if tx.Protected() {
		signer = types.NewEIP155Signer(tx.ChainId())
	}
	value := tx.Value()
	from, _ := types.Sender(signer, tx)
	txLog := &models.Transaction{
		TxHash:   tx.Hash().Hex(),
		FromUuid: "Anonymous",
		ToUuid:   "Anonymous",
		FromAcct: from.Hex(),
		ToAcct:   tx.To().Hex(),
		XuAmount: (new(big.Int).Div(value, models.XU_UNIT)).Uint64(),
	}
	_, err := orm.Insert(txLog)
	return err
}
