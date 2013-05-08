// Copyright 2013 The Chihaya Authors. All rights reserved.
// Use of this source code is governed by the BSD 2-Clause license,
// which can be found in the LICENSE file.

package storage

import (
	m "github.com/kotokoko/chihaya/models"
)

type UserMapper func(u *m.User) error
type TorrentMapper func(t *m.Torrent) error
type WhitelistMapper func(p *m.Peer) error

type Storage interface {
	FreeleechEnabled() (bool, error)

	// These are used to load the cache
	MapOverUsers(f UserMapper) error
	MapOverTorrents(f TorrentMapper) error
	MapOverWhitelist(f WhitelistMapper) error

	RecordSnatch(peer *m.Peer, now int64)
	RecordTorrent(torrent *m.Torrent, deltaSnatch uint64)
	RecordTransferHistory(peer *Peer, rawDeltaUpload int64, rawDeltaDownload int64, deltaTime int64, deltaSnatch uint64, active bool)
	RecordTransferIP(peer *m.Peer)
	RecordUser(user *m.User, rawDeltaUpload int64, rawDeltaDownload int64, deltaUpload int64, deltaDownload int64)

	m.StatCollector
}
