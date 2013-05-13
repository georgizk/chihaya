// Copyright 2013 The Chihaya Authors. All rights reserved.
// Use of this source code is governed by the BSD 2-Clause license,
// which can be found in the LICENSE file.

package cache

import (
	"fmt"

	m "github.com/kotokoko/chihaya/models"
	"github.com/kotokoko/storage"
)

var drivers = make(map[string]CacheDriver)

type CacheDriver interface {
	New(*config.StorageConfig) (Cache, error)
}

func Register(name string, driver CacheDriver) {
	if driver == nil {
		panic("cache: Register driver is nil")
	}
	if _, dup := drivers[name]; dup {
		panic("cache: Register called twice for driver " + name)
	}
	drivers[name] = driver
}

func New(driverName string, conf *config.StorageConfig) (Cache, error) {
	driver, ok := drivers[driverName]
	if !ok {
		return nil, fmt.Errorf("cache: unknown driver %q (forgotten import?)", driverName)
	}
	store, err := driver.New(conf)
	if err != nil {
		return nil, err
	}
	return store, nil
}

type Cache interface {
	Shutdown() error

	LoadUsers(s storage.Storage) error
	LoadTorrents(s storage.Storage) error
	LoadWhitelist(s storage.Storage) error

	FindUser(passkey string) (*m.User, bool, error)
	FindTorrent(infohash string) (*m.Torrent, bool, error)
	PeerWhitelisted(peerId string) (bool, error)

	SaveUser(u *m.User) error
	SaveTorrent(t *m.Torrent) error

	RemoveUser(u *m.User) error
	RemoveTorrent(t *m.Torrent) error

	m.StatCollector
}
