package core

import (
	"Blog/internal/handlers"
	"database/sql"
	"log"
	"net/http"

	"github.com/golang-migrate/migrate/v4"
)

type Server struct {
	DB          *sql.DB
	Migrate     *migrate.Migrate
	PostHandler *handlers.PostHandler
}

func NewServer(db *sql.DB, m *migrate.Migrate, ph *handlers.PostHandler) *Server {
	return &Server{
		DB:          db,
		Migrate:     m,
		PostHandler: ph,
	}
}

func (s *Server) Start(port string) error {
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	http.HandleFunc("/", s.PostHandler.GetPosts)

	log.Printf("Server started on port %s", port)

	httpErr := http.ListenAndServe(port, nil)

	log.Panic(httpErr)
	return nil
}

func (s *Server) Close() {
	s.DB.Close()
}

func (s *Server) MigrateUp() error {
	if err := s.Migrate.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}
