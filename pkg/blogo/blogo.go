package blogo

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/luizgustavojunqueira/Blogo/internal/auth"
	"github.com/luizgustavojunqueira/Blogo/internal/handlers"
	"github.com/luizgustavojunqueira/Blogo/internal/repository"
)

type User struct {
	Username string
	Password string
}

type BlogoConfig struct {
	BlogName   string
	Title      string
	Port       string
	DB         *sql.DB // A PostgreSQL connection
	AuthConfig *auth.AuthConfig
	Logger     *log.Logger
	Location   *time.Location
	Queries    *repository.Queries
}

type Blogo struct {
	blogName string
	title    string
	port     string
	db       *sql.DB
	auth     *auth.Auth
	logger   *log.Logger
	location *time.Location
	queries  *repository.Queries
}

type PostHandler interface {
	GetPosts(w http.ResponseWriter, r *http.Request)
	CreatePost(w http.ResponseWriter, r *http.Request)
	ParseMarkdown(w http.ResponseWriter, r *http.Request)
	Editor(w http.ResponseWriter, r *http.Request)
	ViewPost(w http.ResponseWriter, r *http.Request)
	DeletePost(w http.ResponseWriter, r *http.Request)
	EditPost(w http.ResponseWriter, r *http.Request)
}

type AuthHandler interface {
	Login(w http.ResponseWriter, r *http.Request)
	Logout(w http.ResponseWriter, r *http.Request)
}

// NewBlogo creates and initializes a new instance of Blogo from the provided configuration.
// It validates the essential dependencies and returns an error if any are missing.
// Note: The database connection must be a PostgreSQL connection.
func NewBlogo(config *BlogoConfig) (*Blogo, error) {
	if config.DB == nil {
		return nil, errors.New("A database connection is required")
	}

	if config.AuthConfig == nil {
		return nil, errors.New("Auth configuration is required")
	}

	auth, err := auth.NewAuth(*config.AuthConfig)
	if err != nil {
		return nil, err
	}

	if config.BlogName == "" {
		return nil, errors.New("A Blog name is required")
	}

	if config.Port == "" {
		return nil, errors.New("Server port is required")
	}

	if config.Logger == nil {
		fmt.Println("Logger not provided. Using default logger")
		config.Logger = log.New(os.Stdout, "", log.LstdFlags)
	}

	if config.Location == nil {
		fmt.Println("Location not provided. Using default location")
		config.Location = time.Local
	}

	if config.Title == "" {
		fmt.Println("Title not provided. Using default title")
		config.Title = "Blogo"
	}

	if config.Queries == nil {
		return nil, errors.New("Queries not provided")
	}

	blog := &Blogo{
		blogName: config.BlogName,
		title:    config.Title,
		port:     config.Port,
		db:       config.DB,
		auth:     auth,
		logger:   config.Logger,
		location: config.Location,
		queries:  config.Queries,
	}

	return blog, nil
}

// Start starts the blog server and listens for incoming requests.
func (blogo *Blogo) Start() error {
	var postHandler PostHandler = handlers.NewPostHandler(blogo.queries, blogo.location, blogo.logger, blogo.auth, blogo.blogName, blogo.title)

	var authHandler AuthHandler = handlers.NewAuthHandler(blogo.auth, blogo.logger, blogo.blogName, blogo.title)

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("internal/static"))))

	http.HandleFunc("/", postHandler.GetPosts)
	http.HandleFunc("/editor", postHandler.Editor)
	http.HandleFunc("/editor/{slug}", postHandler.Editor)
	http.HandleFunc("/post/new", postHandler.CreatePost)
	http.HandleFunc("/post/parse", postHandler.ParseMarkdown)
	http.HandleFunc("/post/{slug}", postHandler.ViewPost)
	http.HandleFunc("/post/delete/{slug}", postHandler.DeletePost)
	http.HandleFunc("/post/edit/{slug}", postHandler.EditPost)

	http.HandleFunc("/login", authHandler.Login)
	http.HandleFunc("/logout", authHandler.Logout)

	blogo.logger.Printf("Starting server on port %s\n", blogo.port)

	err := http.ListenAndServe(":"+blogo.port, nil)
	if err != nil {
		blogo.logger.Printf("Error starting server: %v\n", err)
		return err
	}

	return nil
}
