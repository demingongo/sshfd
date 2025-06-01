/*
Copyright © 2025 demingongo
*/
package cmd

import (
	"github.com/demingongo/sshfd/apps/upload"
	"github.com/demingongo/sshfd/globals"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// uploadCmd represents the upload command
var uploadCmd = &cobra.Command{
	Use:   "upload",
	Short: "Upload a local file to a remote host",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		globals.LoadGlobals()
		upload.Run()
	},
}

func init() {
	rootCmd.AddCommand(uploadCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// uploadCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// uploadCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	uploadCmd.PersistentFlags().StringP("source", "s", "", "source (local file)")
	uploadCmd.PersistentFlags().StringP("target", "t", "", "target (remote path)")

	uploadCmd.MarkPersistentFlagRequired("source")
	uploadCmd.MarkPersistentFlagRequired("target")

	uploadCmd.MarkPersistentFlagFilename("source")

	viper.BindPFlag("source", uploadCmd.PersistentFlags().Lookup("source"))
	viper.BindPFlag("target", uploadCmd.PersistentFlags().Lookup("target"))
}
