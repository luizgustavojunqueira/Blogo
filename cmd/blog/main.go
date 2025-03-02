package main

import (
	"database/sql"
	"log"
	"os"

	"github.com/luizgustavojunqueira/Blog/internal/handlers"
	"github.com/luizgustavojunqueira/Blog/internal/repository"

	"github.com/luizgustavojunqueira/Blog/internal/core"

	_ "github.com/mattn/go-sqlite3"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
)

func main() {
	log.Println("Starting server...")
	if os.Getenv("RAILWAY_ENVIRONMENT") == "" {
		if err := godotenv.Load(); err != nil {
			log.Println("No .env file found, using system environment variables")
		}
	}

	if os.Getenv("DB_PATH") == "" || os.Getenv("SERVER_PORT") == "" || os.Getenv("USERNAME") == "" || os.Getenv("PASSWORD") == "" {
		log.Panic("Missing environment variables")
	}

	db, err := sql.Open("sqlite3", os.Getenv("DB_PATH"))
	if err != nil {
		log.Panic(err)
	}

	driver, errDriver := sqlite3.WithInstance(db, &sqlite3.Config{})

	if errDriver != nil {
		log.Panic(errDriver)
	}

	m, errMigrate := migrate.NewWithDatabaseInstance(
		"file://internal/repository/migrations",
		"sqlite3", driver)

	if errMigrate != nil {
		log.Panic(errMigrate)
	}

	queries := repository.New(db)

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
