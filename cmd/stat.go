/*
Copyright © 2025 demingongo
*/
package cmd

import (
	"github.com/demingongo/sshfd/apps/stat"
	"github.com/demingongo/sshfd/globals"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// statCmd represents the stat command
var statCmd = &cobra.Command{
	Use:   "stat",
	Short: "Remote host statistics",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		globals.LoadGlobals()
		stat.Run()
	},
}

func init() {
	rootCmd.AddCommand(statCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// statCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// statCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	statCmd.PersistentFlags().String("host", "", "host")

	viper.BindPFlag("host", statCmd.PersistentFlags().Lookup("host"))
}
