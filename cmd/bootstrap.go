package cmd

import (
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/danesparza/iamserver/data"
)

// bootstrap represents the boostrap command
var bootstrapCmd = &cobra.Command{
	Use:   "bootstrap",
	Short: "Bootstraps the system",
	Long: `Bootstrap the system by creating the necessary database tables, 
indices, admin user, and credentials.  

Running this more than once may result in errors`,
	Run: func(cmd *cobra.Command, args []string) {
		//	Make sure the system and token database paths exist:
		err := os.MkdirAll(viper.GetString("datastore.system"), 0644)
		if err != nil {
			log.Fatalf("[ERROR] Error trying to prep the system database path: %s", err)
			return
		}

		//	Spin up a Manager
		db, err := data.NewManager(viper.GetString("datastore.system"), viper.GetString("datastore.tokens"))
		if err != nil {
			log.Printf("[ERROR] Error trying to open the system database: %s", err)
			return
		}
		defer db.Close()

		//	Call bootstrap
		user, secret, err := db.SystemBootstrap()

		//	Report any errors
		if err != nil {
			log.Printf("[ERROR] Error trying to bootstrap: %s", err)
			return
		}

		//	Spit out the admin credentials:
		log.Printf(`[INFO] System bootstrapped

######################################
Admin login: %s
Admin password: %s
######################################

PLEASE NOTE this information will ONLY be displayed now.
Passwords are encrypted in the database and are not recoverable. 

Please make a note of the admin passsword.

`, user.Name, secret)
	},
}

func init() {
	rootCmd.AddCommand(bootstrapCmd)
}
