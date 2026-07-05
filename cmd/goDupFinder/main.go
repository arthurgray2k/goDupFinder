package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/arthurgray2k/goDupFinder/pkg/dupfinder"
)

var (
	globalWorkers  int
	globalAlgo     string
	globalMaxDepth int
	globalMinSize  int64
)

var rootCmd = &cobra.Command{
	Use:   "goDupFinder",
	Short: "goDupFinder is a fast and scalable duplicate file finder",
	Long:  `A fast, scalable, and memory-efficient duplicate file finder written in Go.`,
}

func init() {
	rootCmd.PersistentFlags().IntVar(&globalWorkers, "workers", 8, "Number of concurrent hash workers")
	rootCmd.PersistentFlags().StringVar(&globalAlgo, "algo", "sha256", "Hash algorithm: sha256, sha1, md5, blake2")
	rootCmd.PersistentFlags().IntVar(&globalMaxDepth, "max-depth", 0, "Maximum directory depth to scan (0 = unlimited)")
	rootCmd.PersistentFlags().Int64Var(&globalMinSize, "min-size", 1, "Minimum file size to consider in bytes")
}

func getOptions() dupfinder.Options {
	opts := dupfinder.DefaultOptions()
	opts.Workers = globalWorkers
	opts.MaxDepth = globalMaxDepth
	opts.MinSize = globalMinSize
	switch strings.ToLower(globalAlgo) {
	case "md5":
		opts.Algorithm = dupfinder.MD5
	case "sha1":
		opts.Algorithm = dupfinder.SHA1
	case "blake2":
		opts.Algorithm = dupfinder.BLAKE2
	default:
		opts.Algorithm = dupfinder.SHA256
	}
	return opts
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

