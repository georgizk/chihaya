/*
 * This file is part of Chihaya.
 *
 * Chihaya is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * Chihaya is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with Chihaya.  If not, see <http://www.gnu.org/licenses/>.
 */

package server

import (
	"bytes"
	"chihaya/config"
	cdb "chihaya/database"
	"chihaya/util"
	"fmt"
	"log"
	"strconv"
	"time"
)

func whitelisted(peerId string, db *cdb.Database) bool {
	db.WhitelistMutex.RLock()
	defer db.WhitelistMutex.RUnlock()

	var widLen int
	var i int
	var matched bool

	for _, whitelistedId := range db.Whitelist {
		widLen = len(whitelistedId)
		if widLen <= len(peerId) {
			matched = true
			for i = 0; i < widLen; i++ {
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

func announce(params *queryParams, user *cdb.User, ip string, db *cdb.Database, buf *bytes.Buffer) {
	var exists bool

	// Mandatory parameters
	infoHash, _ := params.get("info_hash")
	peerId, _ := params.get("peer_id")
	port, portExists := params.getUint64("port")
	uploaded, uploadedExists := params.getUint64("uploaded")
	downloaded, downloadedExists := params.getUint64("downloaded")
	left, leftExists := params.getUint64("left")

	if !(infoHash != "" && peerId != "" && portExists && uploadedExists && downloadedExists && leftExists) {
		failure("Malformed request", buf)
		return
	}

	if !whitelisted(peerId, db) {
		failure("Your client is not approved", buf)
		return
	}

	// TODO: better synchronization strategy for announces (like per user mutexes)
	db.TorrentsMutex.Lock()
	defer db.TorrentsMutex.Unlock()

	torrent, exists := db.Torrents[infoHash]
	if !exists {
		failure("This torrent does not exist", buf)
		return
	}

	if torrent.Status == 1 && left == 0 {
		log.Printf("Unpruning torrent %d", torrent.Id)
		db.UnPrune(torrent)
		torrent.Status = 0
	} else if torrent.Status != 0 {
		failure(fmt.Sprintf("This torrent does not exist (status: %d, left: %d)", torrent.Status, left), buf)
		return
	}

	now := time.Now().Unix()

	// Optional parameters
	event, _ := params.get("event")
	shouldFlushAddr := false

	var numWantStr string
	var numWant int
	numWantStr, exists = params.get("numwant")
	if !exists {
		numWant = 50
	} else {
		numWant64, _ := strconv.ParseInt(numWantStr, 10, 32)
		numWant = int(numWant64)
		if numWant > 50 || numWant < 0 {
			numWant = 50
		}
	}

	// Match or create peer
	var peer *cdb.Peer
	newPeer := false
	seeding := false
	active := true
	completed := event == "completed"

	if left > 0 {
		if user.DisableDownload {
			failure("Your download privileges are disabled.", buf)
			return
		}
		peer, exists = torrent.Leechers[peerId]
		if !exists {
			newPeer = true
			peer = &cdb.Peer{}
			torrent.Leechers[peerId] = peer
		}
	} else if completed {
		peer, exists = torrent.Leechers[peerId]
		if !exists {
			newPeer = true
			peer = &cdb.Peer{}
			torrent.Seeders[peerId] = peer
		} else {
			// They're a seeder now
			torrent.Seeders[peerId] = peer
			delete(torrent.Leechers, peerId)
		}
		seeding = true
	} else { // Previously completed (probably)
		peer, exists = torrent.Seeders[peerId]
		if !exists {
			peer, exists = torrent.Leechers[peerId]
			if !exists {
				newPeer = true
				peer = &cdb.Peer{}
				torrent.Seeders[peerId] = peer
			} else {
				// They're a seeder now.. Broken client? Unreported snatch?
				torrent.Seeders[peerId] = peer
				delete(torrent.Leechers, peerId)
				// completed = true // TODO: not sure if this will result in over-reported snatches
			}
		}
		seeding = true
	}

	// Update peer info/stats
	if newPeer {
		peer.Id = peerId
		peer.UserId = user.Id
		peer.TorrentId = torrent.Id
		peer.StartTime = now
		peer.LastAnnounce = now
		peer.Uploaded = uploaded
		peer.Downloaded = downloaded
	}

	rawDeltaUpload := int64(uploaded) - int64(peer.Uploaded)
	rawDeltaDownload := int64(downloaded) - int64(peer.Downloaded)

	// If a user restarts a torrent, their delta may be negative, attenuating this to 0 should be fine for stats purposes
	if rawDeltaUpload < 0 {
		rawDeltaUpload = 0
	}
	if rawDeltaDownload < 0 {
		rawDeltaDownload = 0
	}

	var deltaDownload int64
	if !config.GlobalFreeleech {
		deltaDownload = int64(float64(rawDeltaDownload) * user.DownMultiplier * torrent.DownMultiplier)
	}
	deltaUpload := int64(float64(rawDeltaUpload) * user.UpMultiplier * torrent.UpMultiplier)

	peer.Uploaded = uploaded
	peer.Downloaded = downloaded
	peer.Left = left
	peer.Seeding = seeding

	var deltaTime int64
	if seeding {
		deltaTime = now - peer.LastAnnounce
	}
	peer.LastAnnounce = now
	torrent.LastAction = now

	// Handle events
	var deltaSnatch uint64
	if event == "stopped" || event == "paused" {
		/*  We can remove the peer from the list and still have their stats be recorded,
		since we still have a reference to their object. After flushing, all references
		should be gone, allowing the peer to be GC'd.  */
		if seeding {
			delete(torrent.Seeders, peerId)
		} else {
			delete(torrent.Leechers, peerId)
		}

		active = false
	} else if completed {
		db.RecordSnatch(peer, now)
		deltaSnatch = 1
	}

	/*
	 * Generate compact ip/port
	 * Future TODO: possible IPv6 support
	 */
	if active && ip != peer.Ip || uint(port) != peer.Port {
		peer.Addr = []byte{0, 0, 0, 0, 0, 0}
		peer.Port = uint(port)
		peer.Ip = ip
		var val byte
		val = 0
		k := 0
		for i := 0; i < len(ip); i++ {
			if ip[i] == '.' {
				if k > 2 {
					failure("Malformed IP address", buf)
					return
				}
				peer.Addr[k] = val
				val = 0
				k++
			} else if ip[i] >= '0' && ip[i] <= '9' {
				val = val*10 + ip[i] - '0'
			} else {
				failure("IPv4 address required (sorry!)", buf)
				return
			}
		}
		if k != 3 {
			failure("Malformed IP address", buf)
			return
		}
		peer.Addr[3] = val
		peer.Addr[4] = byte(port >> 8)
		peer.Addr[5] = byte(port & 0xff)
		shouldFlushAddr = true
	}

	// If the channels are already full, record* blocks until a flush occurs
	db.RecordTorrent(torrent, deltaSnatch)
	db.RecordTransferHistory(peer, rawDeltaUpload, rawDeltaDownload, deltaTime, deltaSnatch, active)
	db.RecordUser(user, rawDeltaUpload, rawDeltaDownload, deltaUpload, deltaDownload)

	if shouldFlushAddr {
		db.RecordTransferIp(peer)
	}

	// Generate response
	seedCount := len(torrent.Seeders)
	leechCount := len(torrent.Leechers)

	buf.WriteRune('d')
	util.Bencode("complete", buf)
	util.Bencode(seedCount, buf)
	util.Bencode("incomplete", buf)
	util.Bencode(leechCount, buf)
	util.Bencode("interval", buf)
	util.Bencode(config.AnnounceInterval, buf)
	util.Bencode("min interval", buf)
	util.Bencode(config.MinAnnounceInterval, buf)

	if numWant > 0 && active {
		util.Bencode("peers", buf)

		compactString, exists := params.get("compact")
		compact := exists && compactString == "1"

		var peerCount int
		if seeding {
			peerCount = util.Min(numWant, leechCount)
		} else {
			peerCount = util.Min(numWant, leechCount+seedCount-1)
		}

		peersToSend := make([]*cdb.Peer, 0, peerCount)

		if seeding {
			for _, leech := range torrent.Leechers {
				if len(peersToSend) >= numWant {
					break
				}
				if leech.UserId == peer.UserId {
					continue
				}
				peersToSend = append(peersToSend, leech)
			}
		} else {
			/*
			 * The iteration is already "random" as of Go 1 (so we don't need to randomize ourselves):
			 * Each time an element is inserted into the map, it gets a some arbitrary position for iteration
			 * Each time you range over the map, it starts at a random offset into the map's elements
			 * See http://code.google.com/p/go/source/browse/src/pkg/runtime/hashmap.c?name=release-branch.go1#614
			 *
			 * Their fastrand1 function (for the random offset) is somewhat shitty though,
			 * so I'm not 100% sure if this randomness is sufficient for rotating seeds
			 * TODO: May want to look into / test this more though
			 */

			for _, seed := range torrent.Seeders {
				if len(peersToSend) >= numWant {
					break
				}
				if seed.UserId == peer.UserId {
					continue
				}
				peersToSend = append(peersToSend, seed)
			}

			for _, leech := range torrent.Leechers {
				if len(peersToSend) >= numWant {
					break
				}
				if leech.UserId == peer.UserId {
					continue
				}
				peersToSend = append(peersToSend, leech)
			}
		}
		if compact {
			buf.WriteString(strconv.Itoa(len(peersToSend) * 6))
			buf.WriteRune(':')
			for _, other := range peersToSend {
				buf.Write(other.Addr)
			}
		} else {
			buf.WriteRune('l')
			for _, other := range peersToSend {
				buf.WriteRune('d')
				util.Bencode("ip", buf)
				util.Bencode(other.Ip, buf)
				util.Bencode("peer id", buf)
				util.Bencode(other.Id, buf)
				util.Bencode("port", buf)
				util.Bencode(other.Port, buf)
				buf.WriteRune('e')
			}
			buf.WriteRune('e')
		}
	}

	buf.WriteRune('e')
}
