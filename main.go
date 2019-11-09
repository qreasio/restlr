package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi"
	"github.com/joho/godotenv"
	resthttp "github.com/qreasio/restlr/http"
	"github.com/qreasio/restlr/model"
	"github.com/qreasio/restlr/page"
	"github.com/qreasio/restlr/post"
	"github.com/qreasio/restlr/shared"
	"github.com/qreasio/restlr/term"
	"github.com/qreasio/restlr/user"
	log "github.com/sirupsen/logrus"
	"github.com/xo/dburl"
)

var (
	APIHost     = "http://localhost:8080"
	SiteURL     = "http://localhost:8080"
	UploadPath  = "uploads"
	TablePrefix = "wp_"
	APIPath     = "wp-json/wp"
	Version     = "v2"
	ServerPort  = "8080"
)

// SetAPIContext will set the APIConfig struct instance in context to store the important data that will be used in most all endpoints
// so it is easily accessible from endpoint by getting it from context
func SetAPIContext() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			apiModel := model.APIConfig{
				APIHost:     APIHost,
				SiteURL:     SiteURL,
				UploadPath:  UploadPath,
				TablePrefix: TablePrefix,
				APIPath:     APIPath,
				Version:     Version,
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

	ServerPort = os.Getenv("SERVER_PORT")   // Port of API Server
	APIHost = os.Getenv("API_HOST")         // The rest api host
	SiteURL = os.Getenv("SITE_URL")         // The site host
	UploadPath = os.Getenv("UPLOAD_PATH")   // File upload path relative from site host
	TablePrefix = os.Getenv("TABLE_PREFIX") // Database table prefix
	APIPath = os.Getenv("API_PATH")         // Relative API Path to api host
	Version = os.Getenv("VERSION")          // API Version path
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

	//set base api path base on env var
	baseAPIPath := fmt.Sprintf("%s/%s", APIPath, Version)

	//routing
	r.Mount(baseAPIPath+"/posts", post.MakeHTTPHandler(postService))
	r.Mount(baseAPIPath+"/pages", page.MakeHTTPHandler(pageService))

	//handle 404 notfound/invalid route with custom response
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		resthttp.EncodeJSONResponse(context.Background(), w, resthttp.NewRouteNotFoundResponse())
	})

	log.Printf("Restlr API starts to run at port : %s", ServerPort)
	err = http.ListenAndServe(":"+ServerPort, r)

	if err != nil {
		log.Fatal("Error on ListenAndServe:", err)
	}
}
