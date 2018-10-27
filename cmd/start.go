package cmd

import (
	"fmt"
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

var (
	uiDirectory string
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

	//	First, verify the system has been bootstrapped.  If it hasn't, instruct the user to do that first
	if _, err := os.Stat(viper.GetString("datastore.system")); os.IsNotExist(err) {
		log.Fatalf("[ERROR] System has not been bootstrapped.  \n\n*** IAMserver must be bootstrapped prior to use ***\n\nTo bootstrap the system, run:\niamserver bootstrap\n")
	}

	//	Next, verify that the TLS keys and certs we expect to use actually exist.  If they don't,
	//	indicate they need to be created and give some help:
	missingTLSInfo := ""
	if _, err := os.Stat(viper.GetString("apiservice.tlscert")); os.IsNotExist(err) {
		missingTLSInfo = missingTLSInfo + fmt.Sprintf("API service certificate: %s\n", viper.GetString("apiservice.tlscert"))
	}
	if _, err := os.Stat(viper.GetString("apiservice.tlskey")); os.IsNotExist(err) {
		missingTLSInfo = missingTLSInfo + fmt.Sprintf("API service key: %s\n", viper.GetString("apiservice.tlskey"))
	}
	if _, err := os.Stat(viper.GetString("uiservice.tlscert")); os.IsNotExist(err) {
		missingTLSInfo = missingTLSInfo + fmt.Sprintf("UI service certificate: %s\n", viper.GetString("uiservice.tlscert"))
	}
	if _, err := os.Stat(viper.GetString("uiservice.tlskey")); os.IsNotExist(err) {
		missingTLSInfo = missingTLSInfo + fmt.Sprintf("UI service key: %s\n", viper.GetString("uiservice.tlskey"))
	}

	if missingTLSInfo != "" {
		log.Fatalf("[ERROR] TLS files not found.  \n\nThe following items are missing: \n%s\n*** IAMserver requires TLS keys and certs to operate ***\n\nTo generate TLS keys/certs for the local machine (for testing purposes) you can use:\nhttps://github.com/FiloSottile/mkcert\n or \nhttps://www.npmjs.com/package/tls-keygen\n", missingTLSInfo)
	}

	//	Log our TLS key information
	log.Printf("[INFO] API TLS cert: %s\n", viper.GetString("apiservice.tlscert"))
	log.Printf("[INFO] API TLS key: %s\n", viper.GetString("apiservice.tlskey"))
	log.Printf("[INFO] UI TLS cert: %s\n", viper.GetString("uiservice.tlscert"))
	log.Printf("[INFO] UI TLS key: %s\n", viper.GetString("uiservice.tlskey"))

	//	Create our 'sigs' and 'done' channels (this is for graceful shutdown)
	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)

	//	Indicate what signals we're waiting for:
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)

	//	Create a DBManager object and associate with the api.Service
	log.Printf("[INFO] Using System DB: %s", viper.GetString("datastore.system"))
	log.Printf("[INFO] Using Token DB: %s", viper.GetString("datastore.tokens"))
	db, err := data.NewManager(viper.GetString("datastore.system"), viper.GetString("datastore.tokens"))
	if err != nil {
		log.Printf("[ERROR] Error trying to open the system or token database: %s", err)
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
	if viper.GetString("uiservice.ui-dir") == "" {
		//	Use the static assets file generated with
		//	https://github.com/elazarl/go-bindata-assetfs using the application-monitor-ui from
		//	https://github.com/danesparza/application-monitor-ui.
		//
		//	To generate this file, place the 'ui'
		//	directory under the main application-monitor-ui directory and run the commands:
		//	go-bindata-assetfs -pkg cmd build/...
		//	Move bindata_assetfs.go to the application-monitor cmd directory
		//	go install ./...
		//  // Router.PathPrefix("/ui").Handler(http.StripPrefix("/ui", http.FileServer(assetFS())))
	} else {
		//	Use the supplied directory:
		log.Printf("[INFO] Using UI directory: %s\n", viper.GetString("uiservice.ui-dir"))
		UIRouter.PathPrefix("/").Handler(http.StripPrefix("/", http.FileServer(http.Dir(viper.GetString("uiservice.ui-dir")))))
	}

	//	SERVICE ROUTES
	//	-- Auth
	APIRouter.HandleFunc("/auth/token", apiService.GetTokenForCredentials).Methods("GET")   // Get a token (from credentials)
	APIRouter.HandleFunc("/auth/authorize", apiService.IsRequestAuthorized).Methods("POST") // Validate a request for a given token
	//	-- OAuth
	APIRouter.HandleFunc("/oauth/token/client", api.HelloWorld).Methods("POST")
	APIRouter.HandleFunc("/oauth/authorize", api.HelloWorld).Methods("GET")
	//	-- 2FA enrollment
	APIRouter.HandleFunc("/2fa", apiService.BeginTOTPEnrollment).Methods("POST")
	APIRouter.HandleFunc("/2fa", apiService.GetTOTPEnrollmentImage).Methods("GET")
	//	-- User
	APIRouter.HandleFunc("/system/users", apiService.AddUser).Methods("POST")                              // Add a user
	APIRouter.HandleFunc("/system/users", apiService.GetAllUsers).Methods("GET")                           // Get all users
	APIRouter.HandleFunc("/system/user/{username}", apiService.GetUser).Methods("GET")                     // Get a user
	APIRouter.HandleFunc("/system/user/{username}/policies", apiService.GetPoliciesForUser).Methods("GET") // Get policies for a user
	//	-- Group
	APIRouter.HandleFunc("/system/groups", apiService.AddGroup).Methods("POST")                                   // Add a group
	APIRouter.HandleFunc("/system/groups", apiService.GetAllGroups).Methods("GET")                                // Get all groups
	APIRouter.HandleFunc("/system/group/{groupname}", apiService.GetGroup).Methods("GET")                         // Get a group
	APIRouter.HandleFunc("/system/group/{groupname}/users/{userlist}", apiService.AddUsersToGroup).Methods("PUT") // Add users to a group
	//	-- Resource
	APIRouter.HandleFunc("/system/resources", apiService.AddResource).Methods("POST")                                            // Add a resource
	APIRouter.HandleFunc("/system/resources", apiService.GetAllResources).Methods("GET")                                         // Get all resources
	APIRouter.HandleFunc("/system/resource/{resourcename}", apiService.GetResource).Methods("GET")                               // Get a resource
	APIRouter.HandleFunc("/system/resource/{resourcename}/actions/{actionlist}", apiService.AddActionsToResource).Methods("PUT") // Add actions to a resource
	//	-- Policy
	APIRouter.HandleFunc("/system/policies", apiService.AddPolicy).Methods("POST")                                         // Add a policy
	APIRouter.HandleFunc("/system/policies", apiService.GetAllPolicies).Methods("GET")                                     // Get all policies
	APIRouter.HandleFunc("/system/policy/{policyname}", apiService.GetPolicy).Methods("GET")                               // Get a policy
	APIRouter.HandleFunc("/system/policy/{policyname}/users/{userlist}", apiService.AttachPolicyToUsers).Methods("PUT")    // Attach policy to user(s)
	APIRouter.HandleFunc("/system/policy/{policyname}/groups/{grouplist}", apiService.AttachPolicyToGroups).Methods("PUT") // Attach policy to group(s)
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

	startCmd.Flags().StringVarP(&uiDirectory, "ui-dir", "u", "", "Directory for the UI")
	viper.BindPFlag("uiservice.ui-dir", startCmd.Flags().Lookup("ui-dir"))
}
