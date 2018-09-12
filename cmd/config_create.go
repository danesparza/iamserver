package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

var yamlDefault = []byte(`# Config created %s
uiservice:
  port: 3000
  tlscert: cert.pem
  tlskey: key.pem
apiservice:
  port: 3001
  tlscert: cert.pem
  tlskey: key.pem
datastore:
  system: ./db/system
  tokens: ./db/token
`)

// configcreateCmd represents the configcreate command
var configcreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Renders a default config file",
	Long:  `Outputs a config file with default values.  Write this to a file called 'authserver.yml' and customize the values.`,
	Run: func(cmd *cobra.Command, args []string) {
		t := time.Now()
		fmt.Printf(string(yamlDefault), t.Format(time.RFC3339))
	},
}

func init() {
	configCmd.AddCommand(configcreateCmd)
}
