/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/joshjms/castletown/server"
	"github.com/spf13/cobra"
)

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Starts the Castletown server",
	Long: `When the light is running low and the shadows start to grow
And the places that you know seem like fantasy
There's a light inside your soul
That's still shining in the cold with the truth
The promise in our hearts
Don't forget, I'm with you in the dark`,
	Run: func(cmd *cobra.Command, args []string) {
		port, _ := cmd.Flags().GetInt("port")
		overlayfsDir, _ := cmd.Flags().GetString("overlayfs_dir")
		filesDir, _ := cmd.Flags().GetString("files_dir")
		imagesDir, _ := cmd.Flags().GetString("images_dir")
		libcontainerDir, _ := cmd.Flags().GetString("libcontainer_dir")

		if _, err := os.Stat(libcontainerDir); os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "Libcontainer directory does not exist, creating: %s\n", libcontainerDir)
			if err := os.MkdirAll(libcontainerDir, 0755); err != nil {
				fmt.Fprintf(os.Stderr, "Error creating libcontainer directory: %v\n", err)
				os.Exit(1)
			}
		}

		if _, err := os.Stat(imagesDir); os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "Images directory does not exist, creating: %s\n", imagesDir)
			if err := os.MkdirAll(imagesDir, 0755); err != nil {
				fmt.Fprintf(os.Stderr, "Error creating images directory: %v\n", err)
				os.Exit(1)
			}
		}

		serverHandler(port, overlayfsDir, filesDir, imagesDir, libcontainerDir)
	},
}

func serverHandler(port int, overlayfsDir, filesDir, imagesDir, libcontainerDir string) {
	s, err := server.NewServer(
		port,
		overlayfsDir,
		filesDir,
		imagesDir,
		libcontainerDir,
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating server: %v\n", err)
		os.Exit(1)
	}
	s.Init()
	s.Start()
}

func init() {
	rootCmd.AddCommand(serverCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// serverCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// serverCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	serverCmd.Flags().IntP("port", "p", 8000, "Port to run the server on")
	serverCmd.Flags().String("overlayfs_dir", "/tmp/ct/overlayfs", "Directory for overlayfs directories (i.e. lower, upper, and work directories)")
	serverCmd.Flags().String("files_dir", "/tmp/ct/files", "Directory for persistent file storage")
	serverCmd.Flags().String("images_dir", "/tmp/ct/images", "Directory for container rootfs images")
	serverCmd.Flags().String("libcontainer_dir", "/tmp/ct/libcontainer", "Directory for libcontainer containers")
}
