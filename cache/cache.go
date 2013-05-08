// Copyright 2013 The Chihaya Authors. All rights reserved.
// Use of this source code is governed by the BSD 2-Clause license,
// which can be found in the LICENSE file.

package cache

import (
	m "github.com/kotokoko/chihaya/models"
	"github.com/kotokoko/storage"
)

type Cache interface {
	LoadUsers(s Storage) error
	LoadTorrents(s Storage) error
	LoadWhitelist(s Storage) error

	FindTorrent(infohash string) (*m.Torrent, bool, error)
	FindUser(passkey string) (*m.User, bool, error)
	PeerWhitelisted(peerId *m.Peer) (bool, error)

	SaveTorrent(t *m.Torrent) error
	SaveUser(u *m.User) error

	m.StatCollector
}
