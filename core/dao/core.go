package dao

import (
	"database/sql"
	"fmt"
	"github.com/gnasnik/titan-workerd-api/config"
	_ "github.com/go-sql-driver/mysql"
	logging "github.com/ipfs/go-log"
	"github.com/jmoiron/sqlx"
	"time"
)

var (
	// DB reference to database
	DB *sqlx.DB
)

const (
	maxOpenConnections = 60
	connMaxLifetime    = 120
	maxIdleConnections = 30
	connMaxIdleTime    = 20
)

var ErrNoRow = fmt.Errorf("no matching row found")
var log = logging.Logger("device_info")

// StringFrom convert string to sql.NullString
func StringFrom(str string) sql.NullString {
	return sql.NullString{String: str, Valid: true}
}

// Int32From convert int32 to sql.NullInt32
func Int32From(i int32) sql.NullInt32 {
	return sql.NullInt32{Int32: i, Valid: true}
}

// Int64From convert int64 to sql.NullInt64
func Int64From(i int64) sql.NullInt64 {
	return sql.NullInt64{Int64: i, Valid: true}
}

func Init(cfg *config.Config) error {
	if cfg.DatabaseURL == "" {
		return fmt.Errorf("database url not setup")
	}

	db, err := sqlx.Connect("mysql", cfg.DatabaseURL)
	if err != nil {
		return err
	}

	db.SetMaxOpenConns(maxOpenConnections)
	db.SetConnMaxLifetime(connMaxLifetime * time.Second)
	db.SetMaxIdleConns(maxIdleConnections)
	db.SetConnMaxIdleTime(connMaxIdleTime * time.Second)

	DB = db
	return nil
}

type QueryOption struct {
	Page       int    `json:"page"`
	PageSize   int    `json:"page_size"`
	Order      string `json:"order"`
	OrderField string `json:"order_field"`
	StartTime  string `json:"start_time"`
	EndTime    string `json:"end_time" `
	UserID     string `json:"user_id"`
}
