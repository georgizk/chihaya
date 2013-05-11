// Copyright 2013 The Chihaya Authors. All rights reserved.
// Use of this source code is governed by the BSD 2-Clause license,
// which can be found in the LICENSE file.

package deprecated

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"

	m "github.com/kotokoko/chihaya/models"
	"github.com/kotokoko/chihaya/storage"
	"github.com/kotokoko/config"
)

type DeprecatedStorage struct {
	database *sql.DB

	freeleechEnabledStmt sql.Stmt
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

	ds := &DeprecatedStorage{
		database: db,
	}

	err = ds.prepareStmts()
	if err != nil {
		return nil, err
	}

	err := ds.createSchema()
	if err != nil {
		return nil, err
	}

	return ds, nil
}

func (ds *DeprecatedStorage) prepareStmts() error {
  ds.freeleechEnabledStmt, err := ds.database.Prepare(
		"SELECT mod_setting FROM mod_core WHERE mod_option='global_freeleech'",
	)
  return err
}

// createSchema() creates the schema if necessary.
func (ds *DeprecatedStorage) createSchema() error {
}

func (ds *DeprecatedStorage) FreeLeechEnabled() (enabled bool, err error) {
  err = ds.freeleechEnabledStmt.QueryRow(/*TODO*/).Scan(&enabled)
  if err != nil {
    return false error
  }

}
