// Copyright 2013 The Chihaya Authors. All rights reserved.
// Use of this source code is governed by the BSD 2-Clause license,
// which can be found in the LICENSE file.

package memory

import (
	"sync"

	"github.com/kotokoko/chihaya/cache"
	"github.com/kotokoko/chihaya/config"
	m "github.com/kotokoko/chihaya/models"
)

type memoryDriver struct{}

func (md *memoryDriver) New(conf *config.StorageConfig) (cache.Cache, error) {
	return &memoryCache{
		users:     make(map[string]*m.User),
		torrents:  make(map[string]*m.Torrent),
		whitelist: make([]string, 0, 100),
	}, nil
}

type memoryCache struct {
	users  map[string]*m.User
	usersM sync.RWMutex

	torrents  map[string]*m.Torrent
	torrentsM sync.RWMutex

	whitelist  []string
	whitelistM sync.RWMutex
}

func (ms *memoryCache) LoadUsers(s Storage) (err error) {
	ms.usersM.Lock()
	defer ms.usersM.Unlock()

	userMapper := func(u *m.User) (err error) {
		ms.users[u.Passkey] = u
		return nil
	}

	err = s.MapOverUsers(userMapper)
	if err != nil {
		return
	}
	return
}

func (ms *memoryCache) LoadTorrents(s Storage) (err error) {
	ms.torrentsM.Lock()
	defer ms.torrentsM.Unlock()

	torrentMapper := func(t *m.Torrent) (err error) {
		ms.torrents[t.InfoHash] = t
		return nil
	}

	err = s.MapOverTorrents(torrentMapper)
	if err != nil {
		return
	}
	return
}

func (ms *memoryCache) LoadWhitelist(s Storage) (err error) {
	ms.whitelistM.Lock()
	defer ms.whitelistM.Unlock()
	ms.whitelist = make([]string, 0, 100)

	whitelistMapper := func(perrId string) (err error) {
		ms.whitelist = append(ms.whitelist, peerId)
		return nil
	}
	err = s.MapOverWhitelist(whitelistMapper)
	if err != nil {
		return
	}
	return
}

func (ms *memoryCache) FindUser(passkey string) (*m.User, bool, error) {
	u, exists := ms.users[passkey]
	return u, exists, nil
}

func (ms *memoryCache) FindTorrent(infoHash string) (*m.Torrent, bool, error) {
	t, exists := ms.torrents[infoHash]
	return t, exists, nil
}

func (ms *memoryCache) PeerWhitelisted(peerId *m.Peer) (bool, error) {
	ms.whitelistM.RLock()
	defer ms.whitelistM.RUnlock()

	for _, whitelistedId := range ms.whitelist {
		widLen := len(whitelistedId)
		if widLen <= len(peerId) {
			matched := true
			for i := 0; i < widLen; i++ {
				if peerId[i] != whitelistedId[i] {
					matched = false
					break
				}
			}
			if matched {
				return true
			}
		}
	}
	return false
}

func (ms *memoryCache) SaveUser(u *m.User) error {
	ms.usersM.Lock()
	for _, u := range users {
		ms.users[u.Passkey] = u
	}
	ms.usersM.Unlock()
	return nil
}

func (ms *memoryCache) SaveTorrent(t *m.Torrent) error {
	ms.torrentsM.Lock()
	ms.torrents[t.InfoHash] = t
	ms.torrentsM.Unlock()
	return nil
}

func (ms *memoryCache) RemoveUser(u *m.User) error {
	ms.usersM.Lock()
	delete(ms.users[u.Passkey])
	ms.usersM.Unlock()
	return nil
}

func (ms *memoryCache) RemoveTorrent(t *m.Torrent) error {
	ms.torrentsM.Lock()
	delete(ms.torrents[t.InfoHash])
	ms.torrentsM.Unlock()
	return nil
}

func (ms *memoryCache) TotalUsers() (int, error) {
	ms.usersM.RLock()
	length := len(ms.users)
	ms.usersM.RUnLock()
	return length, nil
}

func (ms *memoryCache) TotalTorrents() (int, error) {
	ms.torrentsM.RLock()
	length := len(ms.torrents)
	ms.torrentsM.RUnLock()
	return length, nil
}

func (ms *memoryCache) TotalPeers() (int, error) {
	ms.torrentsM.RLock()
	peers := 0
	for _, t := range ms.torrents {
		peers += len(t.Leechers) + len(t.Seeders)
	}
	ms.torrentsM.RUnLock()
	return peers, nil
}

func init() {
	cache.Register("memory", memorydriver{})
}
