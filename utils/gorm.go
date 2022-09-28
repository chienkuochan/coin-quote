package utils

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"strings"
)

var Db *gorm.DB

type GormDB struct {
	*gorm.DB
	gdbDone bool
}

func InitDB() {
	config := getDatabaseConfig()

	var connstring string
	connstring = getConnectionString(config)
	db, err := gorm.Open("mysql", connstring)
	if err != nil {
		panic(err)
	}
	db.DB().SetMaxIdleConns(config.GetInt("production.pool", 5))
	db.DB().SetMaxOpenConns(config.GetInt("production.maxopen", 0))
	Db = db
}

func CloseDB() {
	Db.Close()
}

func getConnectionString(config *ConfigEnv) string {
	host := config.Get("production.host", "")
	port := config.Get("production.port", "3306")
	user := config.Get("production.username", "")
	pass := config.Get("production.password", "")
	dbname := config.Get("production.database", "")
	protocol := config.Get("production.protocol", "tcp")
	dbargs := config.Get("production.dbargs", " ")

	if strings.Trim(dbargs, " ") != "" {
		dbargs = "?" + dbargs
	} else {
		dbargs = ""
	}
	return fmt.Sprintf("%s:%s@%s([%s]:%s)/%s%s", user, pass, protocol, host, port, dbname, dbargs)
}

func DbBegin() *GormDB {
	txn := Db.Begin()
	if txn.Error != nil {
		panic(txn.Error)
	}
	return &GormDB{txn, false}
}
func (c *GormDB) DbCommit() {
	if c.gdbDone {
		return
	}
	tx := c.Commit()
	c.gdbDone = true
	if err := tx.Error; err != nil && err != sql.ErrTxDone {
		panic(err)
	}
}

func (c *GormDB) DbRollback() {
	if c.gdbDone {
		return
	}
	tx := c.Rollback()
	c.gdbDone = true
	if err := tx.Error; err != nil && err != sql.ErrTxDone {
		panic(err)
	}
}
