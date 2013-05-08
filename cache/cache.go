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

	FindUser(passkey string) (*m.User, bool, error)
	FindTorrent(infohash string) (*m.Torrent, bool, error)
	PeerWhitelisted(peerId *m.Peer) (bool, error)

	SaveUser(u *m.User) error
	SaveTorrent(t *m.Torrent) error

	RemoveUser(u *m.User) error
	RemoveTorrent(t *m.Torrent) error

	m.StatCollector
}
