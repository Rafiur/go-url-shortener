package postgres

import (
	"database/sql"
	"fmt"
	"github.com/Rafiur/go-url-shortener/internal/config"
	"log"
	"net/url"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/extra/bundebug"
	"github.com/uptrace/bun/extra/bunotel"
)

var DB *bun.DB

func NewDB(conf *config.Config) *bun.DB {
	var err error
	dsn := fmt.Sprintf("postgres://%s:%s@%s:5432/%s?sslmode=disable&search_path=%s",
		conf.DbUser, url.QueryEscape(conf.DbPass), conf.DBHost, conf.DbName, conf.DbSchema)
	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))

	err = sqldb.Ping()
	if err != nil {
		log.Fatal(err)
	}
	//sqldb.SetMaxOpenConns(conf.SetMaxOpenConns)
	db := bun.NewDB(sqldb, pgdialect.New())

	//if DEBUG=1, enable sql query printing
	if conf.Debug {
		db.AddQueryHook(bundebug.NewQueryHook(bundebug.WithVerbose(true)))
	}

	db.AddQueryHook(bunotel.NewQueryHook())
	return db
}
