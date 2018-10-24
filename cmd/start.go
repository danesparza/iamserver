package cmd

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

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

	//	Create our 'sigs' and 'done' channels
	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)

	//	Indicate what signals we're waiting for:
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)

	//	Create a DBManager object and associate with the api.Service
	db, err := data.NewManager(viper.GetString("datastore.system"), viper.GetString("datastore.tokens"))
	if err != nil {
		log.Printf("[ERROR] Error trying to open the system database: %s", err)
		return
	}
	defer db.Close()
	apiService := api.Service{DB: db}

	//	Log the token TTL:
	tokenttlstring := viper.GetString("apiservice.tokenttl")
	_, err = strconv.Atoi(tokenttlstring)
	if err != nil {
		log.Fatalf("[ERROR] The apiservice.tokenttl config is invalid: %s", err)
	}
	log.Printf("[INFO] Token TTL: %s minutes", tokenttlstring)

	//	Create a router and setup our REST endpoints...
	UIRouter := mux.NewRouter()
	APIRouter := mux.NewRouter()

	//	UI ROUTES
	UIRouter.HandleFunc("/", api.ShowUI)

	//	SERVICE ROUTES
	//	-- Auth
	APIRouter.HandleFunc("/auth/token", apiService.GetTokenForCredentials).Methods("GET")   // Get a token (from credentials)
	APIRouter.HandleFunc("/auth/authorize", apiService.IsRequestAuthorized).Methods("POST") // Validate a request for a given token
	//	-- OAuth
	APIRouter.HandleFunc("/oauth/token/client", api.HelloWorld).Methods("POST")
	APIRouter.HandleFunc("/oauth/authorize", api.HelloWorld).Methods("GET")
	//	-- User
	APIRouter.HandleFunc("/system/users", apiService.AddUser).Methods("POST")          // Add a user
	APIRouter.HandleFunc("/system/users", apiService.GetAllUsers).Methods("GET")       // Get all users
	APIRouter.HandleFunc("/system/user/{username}", apiService.GetUser).Methods("GET") // Get a user
	//	-- Group
	APIRouter.HandleFunc("/system/groups", apiService.AddGroup).Methods("POST")                                      // Add a group
	APIRouter.HandleFunc("/system/groups", apiService.GetAllGroups).Methods("GET")                                   // Get all groups
	APIRouter.HandleFunc("/system/group/{groupname}", apiService.GetGroup).Methods("GET")                            // Get a group
	APIRouter.HandleFunc("/system/group/{groupname}/addusers/{userlist}", apiService.AddUsersToGroup).Methods("PUT") // Add users to a group
	//	-- Resource
	APIRouter.HandleFunc("/system/resources", apiService.AddResource).Methods("POST")                                               // Add a resource
	APIRouter.HandleFunc("/system/resources", apiService.GetAllResources).Methods("GET")                                            // Get all resources
	APIRouter.HandleFunc("/system/resource/{resourcename}", apiService.GetResource).Methods("GET")                                  // Get a resource
	APIRouter.HandleFunc("/system/resource/{resourcename}/addactions/{actionlist}", apiService.AddActionsToResource).Methods("PUT") // Add actions to a resource
	//	-- Policy
	APIRouter.HandleFunc("/system/policies", apiService.AddPolicy).Methods("POST")           // Add a policy
	APIRouter.HandleFunc("/system/policy/{policyname}", apiService.GetPolicy).Methods("GET") // Get a policy
	//	-- Role
	APIRouter.HandleFunc("/system/roles", apiService.AddRole).Methods("POST")                                             // Add a role
	APIRouter.HandleFunc("/system/roles", apiService.GetAllRoles).Methods("GET")                                          // Get all roles
	APIRouter.HandleFunc("/system/role/{rolename}", apiService.GetRole).Methods("GET")                                    // Get a role
	APIRouter.HandleFunc("/system/role/{rolename}/policies/{policylist}", apiService.AttachPoliciesToRole).Methods("PUT") // Attach role to policy(s)
	APIRouter.HandleFunc("/system/role/{rolename}/groups/{grouplist}", apiService.AttachRoleToGroups).Methods("PUT")      // Attach role to group(s)
	APIRouter.HandleFunc("/system/role/{rolename}/users/{userlist}", apiService.AttachRoleToUsers).Methods("PUT")         // Attach role to user(s)

	//	Setup the CORS options:
	log.Printf("[INFO] Allowed CORS origins: %s\n", viper.GetString("apiservice.allowed-origins"))

	corsHandler := cors.New(cors.Options{
		AllowedOrigins:   strings.Split(viper.GetString("apiservice.allowed-origins"), ","),
		AllowCredentials: true,
	}).Handler(APIRouter)

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

	//	Start our shutdown listener (for graceful shutdowns)
	go func() {
		//	If we get a signal...
		_ = <-sigs

		//	Indicate we're done...
		done <- true
	}()

	//	Start the API and UI services
	go func() {
		log.Printf("[INFO] Starting API service: https://%s:%s\n", formattedAPIInterface, viper.GetString("apiservice.port"))
		log.Printf("[ERROR] %v\n", http.ListenAndServeTLS(viper.GetString("apiservice.bind")+":"+viper.GetString("apiservice.port"), viper.GetString("apiservice.tlscert"), viper.GetString("apiservice.tlskey"), corsHandler))
	}()
	go func() {
		log.Printf("[INFO] Starting UI service: https://%s:%s\n", formattedUIInterface, viper.GetString("uiservice.port"))
		log.Printf("[ERROR] %v\n", http.ListenAndServeTLS(viper.GetString("uiservice.bind")+":"+viper.GetString("uiservice.port"), viper.GetString("uiservice.tlscert"), viper.GetString("uiservice.tlskey"), UIRouter))
	}()

	//	Wait for our signal and shutdown gracefully
	<-done
	log.Printf("[INFO] Shutting down ...")
}

func init() {
	rootCmd.AddCommand(startCmd)
}
