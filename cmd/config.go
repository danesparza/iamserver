package cmd

import (
	"fmt"
	"io/ioutil"
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Show or create the config file",
	Long: `By default, this shows the current configuration.  
	
	To create a new config file, use 'config create'`,
	Run: func(cmd *cobra.Command, args []string) {
		//	Get the config file
		//	Read the config file
		dat, err := ioutil.ReadFile(viper.ConfigFileUsed())
		if err != nil {
			log.Fatal(err)
		}

		//	Print the config file
		fmt.Println(string(dat))
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
}
