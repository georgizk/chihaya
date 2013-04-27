// Copyright 2013 The Chihaya Authors. All rights reserved.
// Use of this source code is governed by the BSD 2-Clause license,
// which can be found in the LICENSE file.

package database

import (
	"log"
	"strconv"

	"github.com/garyburd/redigo/redis"
	"github.com/kotokoko/chihaya/config"
)

// subscribe connects to Redis and listens for updates to the database
func (db *Database) subscribe() {
	go func() {
		rc, err := redis.Dial("tcp", config.Loaded.RedisAddress)
		if err != nil {
			log.Println("!!! WARNING !!! Couldn't connect to Redis!")
			return
		}
		defer rc.Close()
		psc := redis.PubSubConn{rc}

		channels := map[string]func(int){
			"torrent_created": db.loadTorrent,
			"torrent_modified": db.loadTorrent,
			"torrent_deleted": db.unloadTorrent,
			"user_created": db.loadUser,
			"user_modified": db.loadUser,
			"user_deleted": db.unloadUser,
		}

		for channel, _ := range channels {
			psc.Subscribe(channel)
		}

		for !db.terminate {
			switch message := psc.Receive().(type) {
			case redis.Message:
				id, err := strconv.Atoi(string(message.Data))
				if err != nil {
					log.Printf("!!! WARNING !!! Got an invalid ID from Redis (%s)", message.Data)
					break
				}
				function, exists := channels[message.Channel]
				if exists {
					function(id)
				} else {
					log.Printf("!!! WARNING !!! Received a message from an unsubscribed Redis channel ???")
				}
			case error:
				log.Printf("!!! WARNING !!! Redis returned error: %s", message)
			}
		}
	}()
}

// loadTorrent retrieves the current state of a torrent from the database and updates it in the cache
func (db *Database) loadTorrent(id int) {
	log.Println("Loading torrent", id)
}

// unloadTorrent removes a torrent from the cache
func (db *Database) unloadTorrent(id int) {
	log.Println("Unloading torrent", id)
}

// loadUser retrieves the current state of a userfrom the database and updates it in the cache
func (db *Database) loadUser(id int) {
	log.Println("Loading user", id)
}

// unloadUser removes a user from the cache
func (db *Database) unloadUser(id int) {
	log.Println("Unloading user", id)
}

