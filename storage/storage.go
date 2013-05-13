// Copyright 2013 The Chihaya Authors. All rights reserved.
// Use of this source code is governed by the BSD 2-Clause license,
// which can be found in the LICENSE file.

package storage

import (
	m "github.com/kotokoko/chihaya/models"
)

var drivers = make(map[string]StorageDriver)

type StorageDriver interface {
	New(*config.StorageConfig) (Storage, error)
}

func Register(name string, driver StorageDriver) {
	if driver == nil {
		panic("storage: Register driver is nil")
	}
	if _, dup := drivers[name]; dup {
		panic("storage: Register called twice for driver " + name)
	}
	drivers[name] = driver
}

func New(driverName string, conf *config.StorageConfig) (Storage, error) {
	driver, ok := drivers[driverName]
	if !ok {
		return nil, fmt.Errorf("storage: unknown driver %q (forgotten import?)", driverName)
	}
	store, err := driver.New(conf)
	if err != nil {
		return nil, err
	}
	return store, nil
}

type UserMapper func(u *m.User) error
type TorrentMapper func(t *m.Torrent) error
type WhitelistMapper func(p *m.Peer) error

type Storage interface {
	Shutdown() error

	FreeleechEnabled() (bool, error)

	// These are used to load the cache
	MapOverUsers(f UserMapper) error
	MapOverTorrents(f TorrentMapper) error
	MapOverWhitelist(f WhitelistMapper) error

	RecordSnatch(peer *m.Peer, now int64) error
	RecordTorrent(torrent *m.Torrent, deltaSnatch uint64) error
	RecordTransferIP(peer *m.Peer) error
	RecordTransferHistory(
		peer *Peer,
		rawDeltaUpload int64,
		rawDeltaDownload int64,
		deltaTime int64,
		deltaSnatch uint64,
		active bool,
	) error
	RecordUser(
		user *m.User,
		rawDeltaUpload int64,
		rawDeltaDownload int64,
		deltaUpload int64,
		deltaDownload int64,
	) error

	m.StatCollector
}
