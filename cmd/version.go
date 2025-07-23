/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var Version string = "0.0.99"

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "castletown version",
	Long:  `castletown version`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("castletown v%s\n", Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// versionCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// versionCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
