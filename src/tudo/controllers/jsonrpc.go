package controllers

import (
	"fmt"
	"net/http"

	"github.com/astaxie/beego"
	"github.com/gorilla/rpc/v2"
	"github.com/gorilla/rpc/v2/json"
)

type KeyArgs struct {
	PublicKey string
	PrivKey   string
}

type KeyReply struct {
	Status string
}

type KeyStore struct{}

func (t *KeyStore) SaveKey(r *http.Request, arg *KeyArgs, reply *KeyReply) error {
	fmt.Printf("Invoke key args %v\n", *arg)
	reply.Status = "Ok"
	return nil
}

func init() {
	fmt.Printf("Init jsonrpc...\n")
	srv := rpc.NewServer()
	srv.RegisterCodec(json.NewCodec(), "application/json")
	srv.RegisterCodec(json.NewCodec(), "application/json;charset=UTF-8")

	srv.RegisterService(new(KeyStore), "")
	beego.Handler("/rpc", srv)
}
