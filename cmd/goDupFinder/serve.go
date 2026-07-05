package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/arthurgray2k/goDupFinder/internal/web"
)

var port int

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the web dashboard UI",
	Long:  `Starts a local web server with a beautiful dashboard to visually configure and run duplicate scans.`,
	Run: func(cmd *cobra.Command, args []string) {
		addr := fmt.Sprintf("127.0.0.1:%d", port)
		server := web.NewServer(addr)
		if err := server.Start(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to start web server: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	serveCmd.Flags().IntVarP(&port, "port", "p", 8080, "Port to run the web server on")
	rootCmd.AddCommand(serveCmd)
}
