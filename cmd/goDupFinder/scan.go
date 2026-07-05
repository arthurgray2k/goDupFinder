package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/arthurgray2k/goDupFinder/internal/progress"
	"github.com/arthurgray2k/goDupFinder/pkg/dupfinder"
)

var scanCmd = &cobra.Command{
	Use:   "scan [directories...]",
	Short: "Scan directories for duplicate files",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		opts := getOptions()

		stats := progress.NewStats()
		tracker := progress.NewTracker(stats, os.Stderr, true)
		hook := progress.NewStatsHook(stats, tracker)
		opts.Hook = hook

		tracker.Start()
		start := time.Now()

		duplicates, err := dupfinder.New(opts).Scan(context.Background(), args)

		tracker.Stop()

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		elapsed := time.Since(start)
		fmt.Printf("\nScan complete in %v\n", elapsed.Round(time.Millisecond))
		fmt.Printf("Found %d duplicate group(s)\n\n", len(duplicates))

		for i, group := range duplicates {
			fmt.Printf("Group %d  Hash: %s\n", i+1, group.Hash)
			for _, file := range group.Files {
				fmt.Printf("  %s\n", file)
			}
			fmt.Println()
		}
	},
}

func init() {
	rootCmd.AddCommand(scanCmd)
}
