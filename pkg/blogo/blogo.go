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

type Config struct {
	BlogName string
	Title    string
	Port     string
	DB       *sql.DB // A PostgreSQL connection
	Auth     *auth.Auth
	Logger   *log.Logger
	Location *time.Location
	Queries  *repository.Queries
}

type Blogo struct {
	config *Config
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
func NewBlogo(config *Config) (*Blogo, error) {
	if config.DB == nil {
		return nil, errors.New("A database connection is required")
	}

	if config.Auth == nil {
		return nil, errors.New("Auth configuration is required")
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
		config: config,
	}

	return blog, nil
}

// Start starts the blog server and listens for incoming requests.
func (blogo *Blogo) Start() error {
	var postHandler PostHandler = handlers.NewPostHandler(blogo.config.Queries, blogo.config.Location, blogo.config.Logger, blogo.config.Auth, blogo.config.BlogName, blogo.config.Title)

	var authHandler AuthHandler = handlers.NewAuthHandler(blogo.config.Auth, blogo.config.Logger, blogo.config.BlogName, blogo.config.Title)

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

	blogo.config.Logger.Printf("Starting server on port %s\n", blogo.config.Port)

	err := http.ListenAndServe(":"+blogo.config.Port, nil)
	if err != nil {
		blogo.config.Logger.Printf("Error starting server: %v\n", err)
		return err
	}

	return nil
}
