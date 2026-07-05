package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/arthurgray2k/goDupFinder/internal/fileops"
	"github.com/arthurgray2k/goDupFinder/pkg/dupfinder"
)

var (
	dryRun      bool
	skipConfirm bool
)

// interactiveConfirm asks the user to confirm each file operation at the terminal.
func interactiveConfirm(op, path string) bool {
	fmt.Printf("  %s %q? [y/N] ", op, path)
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		answer := strings.TrimSpace(strings.ToLower(scanner.Text()))
		return answer == "y" || answer == "yes"
	}
	return false
}

// buildOperator returns an Operator wired to the interactive confirm function
// unless --skip-confirm was passed.
func buildOperator(opts fileops.Options) *fileops.Operator {
	confirm := fileops.DefaultConfirmFunc
	if !skipConfirm && !dryRun {
		confirm = interactiveConfirm
	}
	return fileops.New(opts, confirm)
}

// runScanForOps scans the provided directories and returns duplicate file groups.
func runScanForOps(args []string) ([]dupfinder.DuplicateGroup, error) {
	opts := getOptions()
	finder := dupfinder.New(opts)
	return finder.Scan(context.Background(), args)
}

// ──────────────────────────────────────────────────────────────
// goDupFinder delete <directories...>
// ──────────────────────────────────────────────────────────────

var deleteCmd = &cobra.Command{
	Use:   "delete [directories...]",
	Short: "Delete duplicate files (keeps first in each group)",
	Long: `Scan directories for duplicates and delete the extras.
The first file in each duplicate group is always kept.
Use --dry-run to preview changes without deleting anything.`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		groups, err := runScanForOps(args)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Scan error: %v\n", err)
			os.Exit(1)
		}

		if len(groups) == 0 {
			fmt.Println("No duplicates found.")
			return
		}

		op := buildOperator(fileops.Options{
			DryRun:      dryRun,
			SkipConfirm: skipConfirm,
			KeepFirst:   true,
		})

		if dryRun {
			fmt.Println("[DRY RUN] The following files would be deleted:")
		}

		totalDeleted := 0
		for _, g := range groups {
			results := op.Delete(g.Files)
			for _, r := range results {
				if r.Err != nil {
					fmt.Fprintf(os.Stderr, "  ERROR %s: %v\n", r.Source, r.Err)
					continue
				}
				if r.Skipped {
					fmt.Printf("  SKIP  %s\n", r.Source)
					continue
				}
				prefix := "DELETE"
				if r.DryRun {
					prefix = "WOULD DELETE"
				}
				fmt.Printf("  %-13s %s\n", prefix, r.Source)
				totalDeleted++
			}
		}

		fmt.Printf("\n%d file(s) %s.\n", totalDeleted, map[bool]string{true: "would be deleted", false: "deleted"}[dryRun])
	},
}

// ──────────────────────────────────────────────────────────────
// goDupFinder move <directories...>
// ──────────────────────────────────────────────────────────────

var moveDest string

var moveCmd = &cobra.Command{
	Use:   "move [directories...]",
	Short: "Move duplicate files to a destination directory",
	Long: `Scan directories for duplicates and move extras to --dest.
The first file in each duplicate group is always kept.
Use --dry-run to preview changes without moving anything.`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if moveDest == "" {
			fmt.Fprintln(os.Stderr, "Error: --dest is required")
			os.Exit(1)
		}

		groups, err := runScanForOps(args)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Scan error: %v\n", err)
			os.Exit(1)
		}

		if len(groups) == 0 {
			fmt.Println("No duplicates found.")
			return
		}

		op := buildOperator(fileops.Options{
			DryRun:          dryRun,
			SkipConfirm:     skipConfirm,
			KeepFirst:       true,
			MoveDestination: moveDest,
		})

		if dryRun {
			fmt.Printf("[DRY RUN] The following files would be moved to %s:\n", moveDest)
		}

		totalMoved := 0
		for _, g := range groups {
			results := op.Move(g.Files)
			for _, r := range results {
				if r.Err != nil {
					fmt.Fprintf(os.Stderr, "  ERROR %s: %v\n", r.Source, r.Err)
					continue
				}
				if r.Skipped {
					fmt.Printf("  SKIP  %s\n", r.Source)
					continue
				}
				prefix := "MOVE"
				if r.DryRun {
					prefix = "WOULD MOVE"
				}
				fmt.Printf("  %-11s %s  →  %s\n", prefix, r.Source, r.Target)
				totalMoved++
			}
		}

		fmt.Printf("\n%d file(s) %s.\n", totalMoved, map[bool]string{true: "would be moved", false: "moved"}[dryRun])
	},
}

// ──────────────────────────────────────────────────────────────
// goDupFinder symlink <directories...>
// ──────────────────────────────────────────────────────────────

var symlinkCmd = &cobra.Command{
	Use:   "symlink [directories...]",
	Short: "Replace duplicate files with symbolic links",
	Long: `Scan directories for duplicates and replace extras with symbolic links
pointing to the first file in each group.
Use --dry-run to preview changes without modifying anything.`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		groups, err := runScanForOps(args)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Scan error: %v\n", err)
			os.Exit(1)
		}

		if len(groups) == 0 {
			fmt.Println("No duplicates found.")
			return
		}

		op := buildOperator(fileops.Options{
			DryRun:      dryRun,
			SkipConfirm: skipConfirm,
		})

		if dryRun {
			fmt.Println("[DRY RUN] The following symlinks would be created:")
		}

		total := 0
		for _, g := range groups {
			results := op.Symlink(g.Files)
			for _, r := range results {
				if r.Err != nil {
					fmt.Fprintf(os.Stderr, "  ERROR %s: %v\n", r.Source, r.Err)
					continue
				}
				if r.Skipped {
					fmt.Printf("  SKIP    %s\n", r.Source)
					continue
				}
				prefix := "SYMLINK"
				if r.DryRun {
					prefix = "WOULD SYMLINK"
				}
				fmt.Printf("  %-13s %s  →  %s\n", prefix, r.Source, r.Target)
				total++
			}
		}

		fmt.Printf("\n%d symlink(s) %s.\n", total, map[bool]string{true: "would be created", false: "created"}[dryRun])
	},
}

// ──────────────────────────────────────────────────────────────
// goDupFinder hardlink <directories...>
// ──────────────────────────────────────────────────────────────

var hardlinkCmd = &cobra.Command{
	Use:   "hardlink [directories...]",
	Short: "Replace duplicate files with hard links",
	Long: `Scan directories for duplicates and replace extras with hard links
pointing to the first file in each group.
Use --dry-run to preview changes without modifying anything.`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		groups, err := runScanForOps(args)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Scan error: %v\n", err)
			os.Exit(1)
		}

		if len(groups) == 0 {
			fmt.Println("No duplicates found.")
			return
		}

		op := buildOperator(fileops.Options{
			DryRun:      dryRun,
			SkipConfirm: skipConfirm,
		})

		if dryRun {
			fmt.Println("[DRY RUN] The following hard links would be created:")
		}

		total := 0
		for _, g := range groups {
			results := op.HardLink(g.Files)
			for _, r := range results {
				if r.Err != nil {
					fmt.Fprintf(os.Stderr, "  ERROR %s: %v\n", r.Source, r.Err)
					continue
				}
				if r.Skipped {
					fmt.Printf("  SKIP      %s\n", r.Source)
					continue
				}
				prefix := "HARDLINK"
				if r.DryRun {
					prefix = "WOULD HARDLINK"
				}
				fmt.Printf("  %-14s %s  →  %s\n", prefix, r.Source, r.Target)
				total++
			}
		}

		fmt.Printf("\n%d hard link(s) %s.\n", total, map[bool]string{true: "would be created", false: "created"}[dryRun])
	},
}

func init() {
	// Shared flags for all file operation commands
	for _, cmd := range []*cobra.Command{deleteCmd, moveCmd, symlinkCmd, hardlinkCmd} {
		cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview changes without modifying any files")
		cmd.Flags().BoolVar(&skipConfirm, "skip-confirm", false, "Skip per-file confirmation prompts")
	}
	moveCmd.Flags().StringVar(&moveDest, "dest", "", "Destination directory for moved files (required)")

	rootCmd.AddCommand(deleteCmd)
	rootCmd.AddCommand(moveCmd)
	rootCmd.AddCommand(symlinkCmd)
	rootCmd.AddCommand(hardlinkCmd)
}
