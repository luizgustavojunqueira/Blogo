package main

import (
	"database/sql"
	"log"
	"os"
	"time"

	"github.com/luizgustavojunqueira/Blogo/internal/auth"
	"github.com/luizgustavojunqueira/Blogo/internal/repository"
	"github.com/luizgustavojunqueira/Blogo/pkg/blogo"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/joho/godotenv"

	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	// Load environment variables from .env file if not running on Railway
	if os.Getenv("RAILWAY_ENVIRONMENT") == "" {
		if err := godotenv.Load(); err != nil {
			log.Println("No .env file found, using system environment variables")
		}
	}

	db, err := sql.Open("sqlite3", os.Getenv("DB_PATH"))
	if err != nil {
		log.Panic(err)
	}

	driver, errDriver := sqlite3.WithInstance(db, &sqlite3.Config{
		DatabaseName: "blog.db",
	})

	if errDriver != nil {
		log.Panic(errDriver)
	}

	m, errMigrate := migrate.NewWithDatabaseInstance(
		"file://internal/repository/migrations",
		"sqlite3", driver)

	if errMigrate != nil {
		log.Panic(errMigrate)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Panic("Failed to run migrations: ", err)
	}

	queries := repository.New(db)

	location, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		log.Panic(err)
	}

	blog, err := blogo.NewBlogo(&blogo.BlogoConfig{
		BlogName: "Luiz Gustavo Junqueira",
		Title:    "Luiz Gustavo",
		Port:     os.Getenv("SERVER_PORT"),
		DB:       db,
		AuthConfig: &auth.AuthConfig{
			Username:      os.Getenv("USERNAME"),
			Password:      os.Getenv("PASSWORD"),
			SecretKey:     os.Getenv("SECRET_KEY"),
			CookieName:    os.Getenv("COOKIE_NAME"),
			TokenValidity: 3600,
		},
		Logger:   log.New(os.Stdout, "", log.LstdFlags),
		Location: location,
		Queries:  queries,
	})
	if err != nil {
		log.Panic(err)
	}

	blog.Start()
}
