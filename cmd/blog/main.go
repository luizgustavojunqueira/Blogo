package main

import (
	"context"
	"log"
	"os"

	"github.com/luizgustavojunqueira/Blog/internal/handlers"
	"github.com/luizgustavojunqueira/Blog/internal/repository"

	"github.com/luizgustavojunqueira/Blog/internal/core"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/joho/godotenv"

	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
)

func main() {
	log.Println("Starting server...")

	if os.Getenv("RAILWAY_ENVIRONMENT") == "" {
		if err := godotenv.Load(); err != nil {
			log.Println("No .env file found, using system environment variables")
		}
	}

	if os.Getenv("DB_URL") == "" || os.Getenv("SERVER_PORT") == "" || os.Getenv("USERNAME") == "" || os.Getenv("PASSWORD") == "" {
		log.Panic("Missing environment variables")
	}

	pool, err := pgxpool.New(context.Background(), os.Getenv("DB_URL"))
	if err != nil {
		log.Panic(err)
	}
	defer pool.Close()

	db := stdlib.OpenDBFromPool(pool)

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Panic(err)
	}

	m, errMigrate := migrate.NewWithDatabaseInstance(
		"file://internal/repository/migrations",
		"postgres", driver)

	if errMigrate != nil {
		log.Panic(errMigrate)
	}

	queries := repository.New(pool)

	ph := handlers.NewPostHandler(queries)
	ah := handlers.NewAuthHandler(os.Getenv("USERNAME"), os.Getenv("PASSWORD"))

	server := core.NewServer(db, m, ph, ah)
	defer server.Close()

	if err := server.MigrateUp(); err != nil {
		log.Panic(err)
	}

	serverPort := os.Getenv("SERVER_PORT")

	if err := server.Start(":" + serverPort); err != nil {
		log.Panic(err)
	}

	log.Println("Server stopped")
}
