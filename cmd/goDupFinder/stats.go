package main

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/arthurgray2k/goDupFinder/internal/progress"
	"github.com/arthurgray2k/goDupFinder/pkg/dupfinder"
)

var statsCmd = &cobra.Command{
	Use:   "stats [directories...]",
	Short: "Scan and display detailed duplicate statistics",
	Long:  `Scan directories for duplicates and print a detailed statistics report with live progress.`,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		opts := getOptions()

		stats := progress.NewStats()
		tracker := progress.NewTracker(stats, os.Stderr, true)
		hook := progress.NewStatsHook(stats, tracker)
		opts.Hook = hook

		tracker.Start()
		duplicates, err := dupfinder.New(opts).Scan(context.Background(), args)
		tracker.Stop()

		if err != nil {
			fmt.Fprintf(os.Stderr, "Scan error: %v\n", err)
			os.Exit(1)
		}

		// Tally duplicate stats from results.
		for _, g := range duplicates {
			stats.DuplicateGroups.Add(1)
			extras := int64(len(g.Files) - 1)
			stats.DuplicateFiles.Add(extras)
			// We don't have per-file sizes in DuplicateGroup yet; wasted bytes
			// will be shown once size tracking is added to DuplicateGroup.
		}

		stats.Summary(os.Stdout)
	},
}

func init() {
	rootCmd.AddCommand(statsCmd)
}
