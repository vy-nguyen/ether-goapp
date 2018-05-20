package keystore

import (
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
)

type KeyStore struct {
	keystore.KeyStore
}

type KeyStoreDb struct {
	userUuid string
	scryptN  int
	scryptP  int
}

func NewKeyStore(scryptN, scryptP int) *KeyStore {
	ks := &KeyStore{
		KeyStore: keystore.KeyStore{
			Storage: &KeyStoreDb{"abc", scryptN, scryptP},
		},
	}
	return ks
}

func (ks *KeyStoreDb) GetKey(addr common.Address,
	key string, auth string) (*keystore.Key, error) {
	return nil, nil
}

func (ks *KeyStoreDb) StoreKey(key string, k *keystore.Key, auth string) error {
	return nil
}

func (ks *KeyStoreDb) JoinPath(key string) string {
	return "Hello"
}
