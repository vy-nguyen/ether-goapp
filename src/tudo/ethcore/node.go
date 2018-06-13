/*
 *--------1---------2---------3---------4---------5---------6---------7---------8--------
 * Copyright (c) 2018 by Vy Nguyen
 * BSD License
 *
 * @author vynguyen
 */

package ethcore

import (
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/rpc"
	"tudo/kstore"
)

type TudoNode struct {
	*node.Node
	kstore kstore.KStoreIface
}

func NewTudoNode(conf *node.Config) (*node.Node, error) {
	n, err := node.New(conf)
	if err != nil {
		return nil, nil
	}
	tudo := &TudoNode{n, nil}
	tudo.NodeIf = tudo

	accman, ksIface, err := makeAccountManager(conf)
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
