package mysql

import (
	m "github.com/kotokoko/chihaya/models"
	"github.com/kotokoko/chihaya/storage"
	"github.com/ziutek/mymysql/mysql"
)

type MySQLStorage struct {
	conn  mysql.Conn
	connM sync.Mutex
}

func New() (storage.Storage, error) {
	ms := &MySQLStorage{}

	err := ms.createSchema()
	if err != nil {
		return nil, err
	}

	return ms, nil
}

// createSchema() creates the schema if necessary.
func (ms *MySQLStorage) createSchema() error {
}
