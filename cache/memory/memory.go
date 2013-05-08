package memory

import (
	"sync"

	"github.com/kotokoko/chihaya/cache"
	"github.com/kotokoko/chihaya/config"
	m "github.com/kotokoko/chihaya/models"
)

type MemoryCache struct {
	users  map[string]*m.User
	usersM sync.RWMutex

	torrents  map[string]*m.Torrent
	torrentsM sync.RWMutex

	whitelist  []string
	whitelistM sync.RWMutex
}

func New() cache.Cache {
	return &MemoryStorage{
		users:     make(map[string]*m.User),
		torrents:  make(map[string]*m.Torrent),
		whitelist: make([]string, 0, 100),
	}
}

func (ms *MemoryCache) Load(s Storage) (err error) {
	err = ms.loadUsers(s)
	if err != nil {
		return
	}

	err = ms.loadTorrents(s)
	if err != nil {
		return
	}

	err = ms.loadWhitelist(s)
	if err != nil {
		return
	}

	return
}

func (ms *MemoryCache) loadUsers(s Storage) (err error) {
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

func (ms *MemoryCache) loadTorrents(s Storage) (err error) {
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

func (ms *MemoryCache) loadWhitelist(s Storage) (err error) {
	ms.whitelistM.Lock()
	defer ms.whitelistM.Unlock()
	ms.whitelist = make([]string, 0, 100)

	whitelistMapper := func(p *m.Peer) (err error) {
		append(ms.whitelist, p.Id)
		return nil
	}
	err = s.MapOverWhitelist(whitelistMapper)
	if err != nil {
		return
	}
	return
}

func (ms *MemoryCache) FindTorrentByInfoHash(infoHash string) (*m.Torrent, error) {
	t, exists := ms.torrents[infoHash]
	if !exists {
		return nil, errors.New("Torrent not found")
	}
	return t, nil
}

func (ms *MemoryCache) FindUserByPasskey(passkey string) (*m.User, error) {
	u, exists := ms.users[passkey]
	if !exists {
		return nil, errors.New("User not found")
	}
	return u, nil
}

func (ms *MemoryCache) PeerWhitelisted(peerId *m.Peer) (bool, error) {
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

func (ms *MemoryCache) SaveTorrent(t *m.Torrent) error {
	ms.torrentsM.Lock()
	ms.torrents[t.InfoHash] = t
	ms.torrentsM.Unlock()
	return nil
}

func (ms *MemoryCache) SaveUser(u *m.User) error {
	ms.usersM.Lock()
	for _, u := range users {
		ms.users[u.Passkey] = u
	}
	ms.usersM.Unlock()
	return nil
}

func (ms *MemoryCache) TotalUsers() (int, error) {
	ms.usersM.RLock()
	length := len(ms.users)
	ms.usersM.RUnLock()
	return length, nil
}

func (ms *MemoryCache) TotalTorrents() (int, error) {
	ms.torrentsM.RLock()
	length := len(ms.torrents)
	ms.torrentsM.RUnLock()
	return length, nil
}

func (ms *MemoryCache) TotalPeers() (int, error) {
	ms.torrentsM.RLock()
	peers := 0
	for _, t := range ms.torrents {
		peers += len(t.Leechers) + len(t.Seeders)
	}
	ms.torrentsM.RUnLock()
	return peers, nil
}
