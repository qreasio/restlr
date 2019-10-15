package main

import (
	"context"
	"fmt"
	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
	"github.com/xo/dburl"
	"net/http"
	"os"
	resthttp "github.com/qreasio/restlr/http"
	"github.com/qreasio/restlr/model"
	"github.com/qreasio/restlr/page"
	"github.com/qreasio/restlr/post"
	"github.com/qreasio/restlr/shared"
	"github.com/qreasio/restlr/term"
	"github.com/qreasio/restlr/user"
	"github.com/joho/godotenv"
)

var (
	APIHost = "http://localhost:8080"
	SiteURL = "http://localhost:8080"
	UploadPath = "uploads"
	TablePrefix = "wp_"
	APIPath = "wp-json/wp"
	Version = "v2"
)

func SetAPIContext() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			apiModel := model.APIModel{
				APIHost: APIHost,
				SiteURL:     SiteURL,
				UploadPath:  UploadPath,
				TablePrefix: TablePrefix,
				APIPath:    APIPath,
				Version:    Version,
			}
			apiModel.APIBaseURL = fmt.Sprintf("%s/%s/%s", apiModel.APIHost, apiModel.APIPath, apiModel.Version)
			ctx := context.WithValue(r.Context(), model.APICONFIGKEY, apiModel)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}


func init() {
	// Log as JSON instead of the default ASCII formatter.
	log.SetFormatter(&log.JSONFormatter{})

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	log.SetOutput(os.Stdout)

	// Only log the warning severity or above.
	log.SetLevel(log.DebugLevel)

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	APIHost = os.Getenv("API_HOST") // The rest api host
	SiteURL = os.Getenv("SITE_URL") // The site host
	UploadPath = os.Getenv("UPLOAD_PATH") // File upload path relative from site host
	TablePrefix = os.Getenv("TABLE_PREFIX") // Database table prefix
	APIPath = os.Getenv("API_PATH") // Relative API Path to api host
	Version = os.Getenv("VERSION") // API Version path
}

func main() {
	DatabaseURL := os.Getenv("DATABASE_URL")
	db, err := dburl.Open(DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}

	//initialize repositories
	postRepository := post.NewRepository(db)
	termRepository := term.NewRepository(db)
	userRepository := user.NewRepository(db)
	sharedRepository := shared.NewRepository(db)

	//initialize services
	postService := post.NewService(postRepository, termRepository, sharedRepository, userRepository)
	pageService := page.NewService(postRepository, sharedRepository, userRepository)

	r := chi.NewRouter()

	//middleware
	r.Use(SetAPIContext())

	//routing
	r.Mount("/wp-json/wp/v2/posts", post.MakeHTTPHandler(postService))
	r.Mount("/wp-json/wp/v2/pages", page.MakeHTTPHandler(pageService))

	//handle 404 notfound/invalid route with custom response
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		resthttp.EncodeJSONResponse(context.Background(), w, resthttp.NewRouteNotFoundResponse())
	})

	log.Info("SERVER is running")
	err = http.ListenAndServe(":8080", r)

	if err != nil {
		log.Fatal("Error on ListenAndServe:", err)
	}

}
