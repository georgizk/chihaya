// Copyright 2013 The Chihaya Authors. All rights reserved.
// Use of this source code is governed by the BSD 2-Clause license,
// which can be found in the LICENSE file.

package tentacles

import (
	"database/sql"
	"fmt"

	m "github.com/kotokoko/chihaya/models"
	"github.com/kotokoko/chihaya/storage"
	"github.com/kotokoko/config"

	_ "github.com/go-sql-driver/mysql"
)

type tentaclesDriver struct{}

func (td *tentaclesDriver) New(conf *config.StorageConfig) (storage.Storage, error) {
	dsn := fmt.Sprintf(
		"%s:%s@%s/%s?charset=%s",
		conf.Username,
		conf.Password,
		conf.Address,
		conf.Database,
		conf.Encoding,
	)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	ts := &tentaclesStorage{
		database: db,
	}

	err = ts.prepareStmts()
	if err != nil {
		return nil, err
	}

	err := ts.createSchema()
	if err != nil {
		return nil, err
	}

	return ts, nil
}

type tentaclesStorage struct {
	database *sql.DB

	loadUsersStmt        *sql.Stmt
	loadTorrentsStmt     *sql.Stmt
	loadWhitelistStmt    *sql.Stmt
	freeleechEnabledStmt *sql.Stmt
}

func (ts *tentaclesStorage) prepareStmts() (err error) {
	ts.loadUsersStmt, err = ts.database.Prepare(
		"SELECT ID, torrent_pass, DownMultiplier, UpMultiplier, Slots " +
			"FROM users_main WHERE Enabled='1'",
	)
	if err != nil {
		return
	}

	ts.loadTorrentsStmt, err = ts.database.Prepare(
		"SELECT ID, info_hash, DownMultiplier, UpMultiplier, Snatched, Status " +
			"FROM torrents",
	)
	if err != nil {
		return
	}

	ts.loadWhitelistStmt, err = ts.database.Prepare(
		"SELECT peer_id FROM xbt_client_whitelist",
	)
	if err != nil {
		return
	}

	ts.freeleechEnabledStmt, err = ts.database.Prepare(
		"SELECT mod_setting FROM mod_core WHERE mod_option='global_freeleech'",
	)
	if err != nil {
		return
	}

	return
}

// createSchema() creates the schema if necessary.
func (ts *tentaclesStorage) createSchema() error {
}

func (ts *tentaclesStorage) FreeLeechEnabled() (enabled bool, err error) {
	err = ts.freeleechEnabledStmt.QueryRow().Scan(&enabled)
	if err != nil {
		return
	}
	return
}

func init() {
	cache.Register("tentacles", tentaclesDriver{})
}
