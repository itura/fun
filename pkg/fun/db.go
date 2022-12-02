package fun

import (
	"database/sql"
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"time"
)

func InitDb(models []interface{}, adapter gorm.Dialector, config *gorm.Config) (*gorm.DB, error) {
	db, err := gorm.Open(adapter, config)
	if err != nil {
		return nil, err
	}

	if err = db.AutoMigrate(models...); err != nil {
		return nil, err
	}

	return db, nil
}

func CloseDb(db *gorm.DB) error {
	connection, err := db.DB()
	if err != nil {
		return err
	}
	err = connection.Close()
	if err != nil {
		return err
	}
	return nil
}

func InitTestDb(models []interface{}) (*gorm.DB, error) {
	return InitSqlite(models, fmt.Sprintf("file:%s?mode=memory&cache=shared", RandString(5)))
}

func InitDevDb(models []interface{}) (*gorm.DB, error) {
	return InitSqlite(models, "dev.db")
}

func InitSqlite(models []interface{}, file string) (*gorm.DB, error) {
	db, err := InitDb(models, sqlite.Open(file), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	tx := db.Exec("PRAGMA foreign_keys = ON")
	if tx.Error != nil {
		return nil, tx.Error
	}

	return db, nil
}

func InitProdDb(models []interface{}, pgConfig *ProdDbConfig, gormConfig *gorm.Config) (*gorm.DB, error) {
	conn, err := InitDb(models, postgres.Open(pgConfig.Dsn()), gormConfig)
	if err != nil {
		return nil, err
	}

	underlying, err := conn.DB()
	if err != nil {
		return nil, err
	}

	pgConfig.ApplyConnectionPoolSettings(underlying)

	return conn, nil
}

type ProdDbConfig struct {
	Host                   string
	User                   string
	Password               string
	DbName                 string
	MaxIdleConns           int
	ConnMaxIdleTimeSeconds time.Duration
	MaxOpenConns           int
	ConnMaxLifetimeSeconds time.Duration
	SslMode                bool
}

func (c ProdDbConfig) Dsn() string {
	var sslMode string
	if c.SslMode {
		sslMode = "enable"
	} else {
		sslMode = "disable"
	}
	return fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.User, c.Password, c.DbName, sslMode,
	)
}

func (c ProdDbConfig) ApplyConnectionPoolSettings(underlying *sql.DB) {
	underlying.SetMaxIdleConns(c.MaxIdleConns)
	underlying.SetConnMaxIdleTime(c.ConnMaxIdleTimeSeconds)
	underlying.SetMaxOpenConns(c.MaxOpenConns)
	underlying.SetConnMaxLifetime(c.ConnMaxLifetimeSeconds)
}
