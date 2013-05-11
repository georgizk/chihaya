// Copyright 2013 The Chihaya Authors. All rights reserved.
// Use of this source code is governed by the BSD 2-Clause license,
// which can be found in the LICENSE file.

package mysql

import (
	"database/sql"
	"fmt"

	m "github.com/kotokoko/chihaya/models"
	"github.com/kotokoko/chihaya/storage"
	"github.com/kotokoko/config"

	_ "github.com/go-sql-driver/mysql"
)

type MySQLStorage struct {
	database *sql.DB
}

func New(conf *config.StorageConfig) (storage.Storage, error) {
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

	ms := &MySQLStorage{
		database: db,
	}

	err := ms.createSchema()
	if err != nil {
		return nil, err
	}

	return ms, nil
}

// createSchema() creates the schema if necessary.
func (ms *MySQLStorage) createSchema() error {
}

func (ms *MySQLStorage) FreeLeechEnabled() (bool, error) {
}
