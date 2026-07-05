package main

import (
	"context"
	"fmt"
	"os"

	"github.com/arthurgray2k/goDupFinder/internal/exporter"
	"github.com/arthurgray2k/goDupFinder/pkg/dupfinder"
	"github.com/spf13/cobra"
)

var (
	exportFormat string
	exportOutput string
)

var exportCmd = &cobra.Command{
	Use:   "export [directories...]",
	Short: "Scan and export duplicates in a specific format",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		opts := getOptions()

		finder := dupfinder.New(opts)
		fmt.Fprintf(os.Stderr, "Scanning %d directories for export...\n", len(args))

		duplicates, err := finder.Scan(context.Background(), args)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error scanning: %v\n", err)
			os.Exit(1)
		}

		exp, err := exporter.CreateExporter(exportFormat, exportOutput)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create exporter: %v\n", err)
			os.Exit(1)
		}
		defer exp.Close()

		if err := exp.Export(duplicates); err != nil {
			fmt.Fprintf(os.Stderr, "Export failed: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	exportCmd.Flags().StringVar(&exportFormat, "format", "json", "Output format (json, ndjson, csv)")
	exportCmd.Flags().StringVar(&exportOutput, "output", "", "Output file (default is stdout)")
	rootCmd.AddCommand(exportCmd)
}
