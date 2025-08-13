package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

	const version = "1.0.0"

	type config struct {
		port int
		env  string
		db struct {
			dsn string
		}
	}

	type application struct {
		config config
		logger *log.Logger
	}

	func main() {
		var cfg config

		// command line flags
		flag.IntVar(&cfg.port, "port", 4000, "API server port")
		flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")
		flag.StringVar(&cfg.db.dsn, "db-dsn", "postgres://popcorndb_api:paSSWORD@localhost/popcorndb_api", "PostgreSQL DSN")
		flag.Parse()

		logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

		db, err := openDB(cfg)
		if err != nil {
			logger.Fatal(err)
		}
		defer db.Close()
		logger.Printf("db connection pool established")

		app := &application{
			config: cfg,
			logger: logger,
		}
		
		srv := &http.Server{
			Addr:         fmt.Sprintf(":%d", cfg.port),
			Handler:      app.routes(),
			IdleTimeout:  time.Minute,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 30 * time.Second,
		}

		logger.Printf("starting %s server on %s", cfg.env, srv.Addr)
		err = srv.ListenAndServe()
		logger.Fatal(err)
	}

	func openDB(cfg config) (*pgxpool.Pool, error) {
		ctx, cancel := context.WithTimeout(context.Background(), 5 *time.Second)
		defer cancel()

		pool, err := pgxpool.New(ctx, cfg.db.dsn)
		if err != nil {
			return nil, err
		}

		if err := pool.Ping(ctx); err != nil {
			pool.Close()
			return nil, err
		}

		return pool, nil
	}
 