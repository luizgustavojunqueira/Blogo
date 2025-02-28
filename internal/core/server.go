package core

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/golang-migrate/migrate/v4"
)

type PostHandler interface {
	GetPosts(w http.ResponseWriter, r *http.Request)
	CreatePost(w http.ResponseWriter, r *http.Request)
	ParseMarkdown(w http.ResponseWriter, r *http.Request)
	Editor(w http.ResponseWriter, r *http.Request)
	ViewPost(w http.ResponseWriter, r *http.Request)
}

type Server struct {
	DB          *sql.DB
	Migrate     *migrate.Migrate
	PostHandler PostHandler
}

func NewServer(db *sql.DB, m *migrate.Migrate, ph PostHandler) *Server {
	return &Server{
		DB:          db,
		Migrate:     m,
		PostHandler: ph,
	}
}

func (s *Server) Start(port string) error {
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("internal/static"))))

	http.HandleFunc("/", s.PostHandler.GetPosts)
	http.HandleFunc("/editor", s.PostHandler.Editor)
	http.HandleFunc("/post/new", s.PostHandler.CreatePost)
	http.HandleFunc("/post/parse", s.PostHandler.ParseMarkdown)
	http.HandleFunc("/post/{slug}", s.PostHandler.ViewPost)

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
