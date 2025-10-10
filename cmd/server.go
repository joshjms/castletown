/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/joshjms/castletown/config"
	"github.com/joshjms/castletown/job"
	"github.com/joshjms/castletown/sandbox"
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
		config.Port, _ = cmd.Flags().GetInt("port")
		config.OverlayFSDir, _ = cmd.Flags().GetString("overlayfs-dir")
		config.StorageDir, _ = cmd.Flags().GetString("storage-dir")
		config.ImagesDir, _ = cmd.Flags().GetString("images-dir")
		config.LibcontainerDir, _ = cmd.Flags().GetString("libcontainer-dir")
		config.MaxConcurrency, _ = cmd.Flags().GetInt("max-concurrency")
		config.RootfsDir, _ = cmd.Flags().GetString("rootfs-dir")

		RunServer()
	},
}

func RunServer() {
	f, err := os.Stat(config.OverlayFSDir)
	if os.IsNotExist(err) {
		if err := os.Mkdir(config.OverlayFSDir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create OverlayFS directory: %v\n", err)
			os.Exit(1)
		}
	} else if !f.IsDir() {
		fmt.Fprintf(os.Stderr, "OverlayFS path exists but is not a directory: %s\n", config.OverlayFSDir)
		os.Exit(1)
	}

	f, err = os.Stat(config.StorageDir)
	if os.IsNotExist(err) {
		if err := os.Mkdir(config.StorageDir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create Storage directory: %v\n", err)
			os.Exit(1)
		}
	} else if !f.IsDir() {
		fmt.Fprintf(os.Stderr, "Storage path exists but is not a directory: %s\n", config.StorageDir)
		os.Exit(1)
	}

	f, err = os.Stat(config.ImagesDir)
	if os.IsNotExist(err) {
		if err := os.Mkdir(config.ImagesDir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create Images directory: %v\n", err)
			os.Exit(1)
		}
	} else if !f.IsDir() {
		fmt.Fprintf(os.Stderr, "Images path exists but is not a directory: %s\n", config.ImagesDir)
		os.Exit(1)
	}

	f, err = os.Stat(config.LibcontainerDir)
	if os.IsNotExist(err) {
		if err := os.Mkdir(config.LibcontainerDir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create Libcontainer directory: %v\n", err)
			os.Exit(1)
		}
	} else if !f.IsDir() {
		fmt.Fprintf(os.Stderr, "Libcontainer path exists but is not a directory: %s\n", config.LibcontainerDir)
		os.Exit(1)
	}

	f, err = os.Stat(config.RootfsDir)
	if os.IsNotExist(err) {
		if err := os.Mkdir(config.RootfsDir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create Rootfs directory: %v\n", err)
			os.Exit(1)
		}
	} else if !f.IsDir() {
		fmt.Fprintf(os.Stderr, "Rootfs path exists but is not a directory: %s\n", config.RootfsDir)
		os.Exit(1)
	}

	job.NewJobPool()

	if err := sandbox.NewManager(); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating sandbox manager: %v\n", err)
		os.Exit(1)
	}

	s, err := server.NewServer()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating server: %v\n", err)
		os.Exit(1)
	}
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

	serverCmd.Flags().String("overlayfs-dir", "/tmp/castletown/overlayfs", "Directory for overlayfs directories (i.e. lower, upper, and work directories)")
	serverCmd.Flags().String("storage-dir", "/tmp/castletown/storage", "Directory for persistent file storage")
	serverCmd.Flags().String("images-dir", "/tmp/castletown/images", "Directory for container rootfs images")
	serverCmd.Flags().String("libcontainer-dir", "/tmp/castletown/libcontainer", "Directory for libcontainer containers")
	serverCmd.Flags().String("rootfs-dir", "/tmp/castletown/rootfs", "Directory for temporary root filesystems")

	serverCmd.Flags().IntP("port", "p", 8000, "Port to run the server on")
	serverCmd.Flags().Int("max-concurrency", 10, "Maximum number of concurrent sandboxes")
}
