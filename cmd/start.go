package cmd

import (
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/danesparza/iamserver/api"
	"github.com/danesparza/iamserver/data"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the API and UI services",
	Long:  `Start the API and UI services`,
	Run:   start,
}

// @title Identity and Access Management API
// @version 1.0
// @description OAuth 2 based token issue and validation server, with built in management UI
// @termsOfService https://github.com/danesparza/iamserver

// @contact.name API Support
// @contact.url https://github.com/danesparza/iamserver

// @license.name MIT License
// @license.url https://github.com/danesparza/iamserver/blob/master/LICENSE

// @host localhost:3001
// @BasePath /

// @securitydefinitions.oauth2.application OAuth2Application
// @tokenUrl https://localhost:3001/oauth/token/client
// @scope.sys_delegate Grants write access for a specific resource
// @scope.sys_admin Grants read and write access to administrative information

func start(cmd *cobra.Command, args []string) {

	//	Create a DBManager object and associate with the api.Service
	db, err := data.NewManager(viper.GetString("datastore.system"), viper.GetString("datastore.tokens"))
	if err != nil {
		log.Printf("[ERROR] Error trying to open the system database: %s", err)
		return
	}
	defer db.Close()
	// apiService := api.Service{DB: db}

	//	Create a router and setup our REST endpoints...
	SystemRouter := mux.NewRouter()
	OAuthRouter := mux.NewRouter()

	//	Setup our UI routes
	SystemRouter.HandleFunc("/", api.ShowUI)

	//	Setup our Service routes
	OAuthRouter.HandleFunc("/oauth/token/client", api.HelloWorld).Methods("POST")
	OAuthRouter.HandleFunc("/oauth/authorize", api.HelloWorld).Methods("GET")

	//	Setup the CORS options:
	log.Printf("[INFO] Allowed CORS origins: %s\n", viper.GetString("apiservice.allowed-origins"))

	corsHandler := cors.New(cors.Options{
		AllowedOrigins:   strings.Split(viper.GetString("apiservice.allowed-origins"), ","),
		AllowCredentials: true,
	}).Handler(OAuthRouter)

	//	Format the bound interface:
	formattedAPIInterface := viper.GetString("apiservice.bind")
	if formattedAPIInterface == "" {
		formattedAPIInterface = "127.0.0.1"
	}

	formattedUIInterface := viper.GetString("uiservice.bind")
	if formattedUIInterface == "" {
		formattedUIInterface = "127.0.0.1"
	}

	//	Log our TLS key information
	log.Printf("[INFO] API TLS cert: %s\n", viper.GetString("apiservice.tlscert"))
	log.Printf("[INFO] API TLS key: %s\n", viper.GetString("apiservice.tlskey"))
	log.Printf("[INFO] UI TLS cert: %s\n", viper.GetString("uiservice.tlscert"))
	log.Printf("[INFO] UI TLS key: %s\n", viper.GetString("uiservice.tlskey"))

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Printf("[INFO] Starting API service: https://%s:%s\n", formattedAPIInterface, viper.GetString("apiservice.port"))
		log.Printf("[ERROR] %v\n", http.ListenAndServeTLS(viper.GetString("apiservice.bind")+":"+viper.GetString("apiservice.port"), viper.GetString("apiservice.tlscert"), viper.GetString("apiservice.tlskey"), corsHandler))
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Printf("[INFO] Starting UI service: https://%s:%s\n", formattedUIInterface, viper.GetString("uiservice.port"))
		log.Printf("[ERROR] %v\n", http.ListenAndServeTLS(viper.GetString("uiservice.bind")+":"+viper.GetString("uiservice.port"), viper.GetString("uiservice.tlscert"), viper.GetString("uiservice.tlskey"), SystemRouter))
	}()

	//	Wait for everything to stop...
	wg.Wait()

}

func init() {
	rootCmd.AddCommand(startCmd)
}
