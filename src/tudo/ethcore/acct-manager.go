/*
 *--------1---------2---------3---------4---------5---------6---------7---------8--------
 * Copyright (c) 2018 by Vy Nguyen
 * BSD License
 *
 * @author vynguyen
 */
package ethcore

import (
	"fmt"
	"reflect"
	"sort"
	"sync"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/event"
	"tudo/kstore"
)

type AmInterface interface {
	accounts.Manager

	Keystore(kind reflect.Type) kstore.KsInterface
	DefaultKeyStore() kstore.KsInterface
}

type Manager struct {
	kstore   map[reflect.Type]kstore.KsInterface
	updaters []event.Subscription
	updates  chan accounts.WalletEvent
	wallets  []accounts.Wallet
	feed     event.Feed
	quit     chan chan error
	lock     sync.RWMutex
}

func NewManager(keystore ...kstore.KsInterface) AmInterface {
	wallets := []accounts.Wallet{}
	updates := make(chan accounts.WalletEvent, 4*len(keystore))
	subs := make([]event.Subscription, len(keystore))
	ksmap := make(map[reflect.Type]kstore.KsInterface)

	for i, ks := range keystore {
		kind := reflect.TypeOf(ks)
		ksmap[kind] = ks
		wallets = mergeSorted(wallets, ks.Wallets()...)
		subs[i] = ks.Subscribe(updates)
	}
	am := &Manager{
		updaters: subs,
		updates:  updates,
		wallets:  wallets,
		quit:     make(chan chan error),
		kstore:   ksmap,
	}
	go am.update()
	return am
}

func (am *Manager) Close() error {
	errc := make(chan error)
	am.quit <- errc
	return <-errc
}

func (am *Manager) Keystore(kind reflect.Type) kstore.KsInterface {
	return am.kstore[kind]
}

func (am *Manager) Backends(kind reflect.Type) []accounts.Backend {
	iface := am.kstore[kind]
	if iface != nil {
		return []accounts.Backend{iface}
	}
	for _, ks := range am.kstore {
		return []accounts.Backend{ks}
	}
	return nil
}

func (am *Manager) DefaultKeyStore() kstore.KsInterface {
	for _, ks := range am.kstore {
		return ks
	}
	return nil
}

func (am *Manager) Wallets() []accounts.Wallet {
	am.lock.RLock()
	defer am.lock.RLock()

	cpy := make([]accounts.Wallet, len(am.wallets))
	copy(cpy, am.wallets)
	return cpy
}

func (am *Manager) Wallet(url string) (accounts.Wallet, error) {
	am.lock.RLock()
	defer am.lock.RLock()

	fmt.Printf("Get wallet from %s\n", url)
	return nil, nil
}

func (am *Manager) Find(account accounts.Account) (accounts.Wallet, error) {
	am.lock.RLock()
	defer am.lock.RLock()

	for _, wallet := range am.wallets {
		if wallet.Contains(account) {
			return wallet, nil
		}
	}
	return nil, accounts.ErrUnknownAccount
}

func (am *Manager) Subscribe(sink chan<- accounts.WalletEvent) event.Subscription {
	fmt.Printf("Subscribed is called\n")
	return nil
}

func (am *Manager) update() {
	defer func() {
		am.lock.Lock()
		for _, sub := range am.updaters {
			sub.Unsubscribe()
		}
		am.updaters = nil
		am.lock.Unlock()
	}()

	for {
		select {
		case event := <-am.updates:
			am.lock.Lock()
			switch event.Kind {
			case accounts.WalletArrived:
				am.wallets = mergeSorted(am.wallets, event.Wallet)

			case accounts.WalletDropped:
				am.wallets = dropSorted(am.wallets, event.Wallet)
			}
			am.lock.Unlock()

			// am.feed.Send(event)

		case errc := <-am.quit:
			errc <- nil
			return
		}
	}
}

func mergeSorted(slice []accounts.Wallet,
	wallets ...accounts.Wallet) []accounts.Wallet {

	for _, wallet := range wallets {
		n := sort.Search(len(slice), func(i int) bool {
			return slice[i].URL().Cmp(wallet.URL()) >= 0
		})
		if n == len(slice) {
			slice = append(slice, wallet)
			continue
		}
		slice = append(slice[:n], append([]accounts.Wallet{wallet}, slice[n:]...)...)
	}
	return slice
}

func dropSorted(slice []accounts.Wallet,
	wallets ...accounts.Wallet) []accounts.Wallet {

	for _, wallet := range wallets {
		n := sort.Search(len(slice), func(i int) bool {
			return slice[i].URL().Cmp(wallet.URL()) >= 0
		})
		if n == len(slice) {
			continue
		}
		slice = append(slice[:n], slice[n+1:]...)
	}
	return slice
}
