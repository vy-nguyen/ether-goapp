/*
 *--------1---------2---------3---------4---------5---------6---------7---------8--------
 * Copyright (c) 2018 by Vy Nguyen
 * BSD License
 *
 * @author vynguyen
 */

package ethcore

import (
	"reflect"

	"github.com/astaxie/beego/orm"
	"github.com/ethereum/go-ethereum/eth"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/rpc"
	"tudo/kstore"
)

type TudoNode struct {
	*node.Node
	ether    *eth.Ethereum
	bcEthApi *eth.EthApiBackend
	kstore   kstore.KStoreIface
}

type TudoConfig struct {
	AdminAccounts []string
	PeerCfgFile   string
}

func NewTudoNode(conf *node.Config, tdcfg *TudoConfig) (*node.Node, error) {
	n, err := node.New(conf)
	if err != nil {
		return nil, nil
	}
	tudo := &TudoNode{n, nil, nil, nil}
	tudo.NodeIf = tudo

	accman, ksIface, err := makeAccountManager(conf, tdcfg)
	if err != nil {
		return tudo.Node, nil
	}
	tudo.kstore = ksIface
	tudo.Node.SetAccountManager(accman)
	return tudo.Node, err
}

func (n *TudoNode) GetApis() []rpc.API {
	apis := n.Node.GetApis()
	return append(apis, rpc.API{
		Namespace: "tudo",
		Version:   "1.0",
		Service:   NewTudoNodeAPI(n),
		Public:    true,
	})
	return apis
}

func (n *TudoNode) GetEthereum() *eth.Ethereum {
	if n.ether == nil {
		e := n.GetService(reflect.TypeOf((*eth.Ethereum)(nil)))
		n.ether = e.(*eth.Ethereum)
		n.bcEthApi = n.ether.ApiBackend
	}
	return n.ether
}

func (n *TudoNode) GetOrm() orm.Ormer {
	return n.kstore.GetStorageIf().GetOrm()
}
