package update

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/creativeprojects/go-selfupdate"
	"github.com/spf13/cobra"

	"github.com/atlanticbt/magecli/internal/build"
	"github.com/atlanticbt/magecli/pkg/cmdutil"
)

const repository = "atlanticbt/magecli"

func NewCmdUpdate(f *cmdutil.Factory) *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update magecli to the latest release",
		Long: `Download and install the latest magecli release from GitHub.

Compares the current version against the latest GitHub release and replaces
the running binary if a newer version is available. Use --force to reinstall
even if already on the latest version.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUpdate(f, force)
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Reinstall even if already on the latest version")

	return cmd
}

func runUpdate(f *cmdutil.Factory, force bool) error {
	ios, err := f.Streams()
	if err != nil {
		return err
	}

	spinner := f.ProgressSpinner()
	currentVersion := strings.TrimPrefix(build.Version, "v")

	updater, err := selfupdate.NewUpdater(selfupdate.Config{
		Validator: &selfupdate.ChecksumValidator{UniqueFilename: "checksums.txt"},
	})
	if err != nil {
		return fmt.Errorf("create updater: %w", err)
	}

	// Detect latest release
	spinner.Start("Checking for updates...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	release, found, err := updater.DetectLatest(ctx, selfupdate.ParseSlug(repository))
	spinner.Stop("")
	if err != nil {
		return fmt.Errorf("check for updates: %w", err)
	}
	if !found {
		return fmt.Errorf("no release found for %s/%s", runtime.GOOS, runtime.GOARCH)
	}

	if release.LessOrEqual(currentVersion) && !force {
		fmt.Fprintf(ios.Out, "Already on the latest version (%s)\n", currentVersion)
		return nil
	}

	if currentVersion == "dev" {
		fmt.Fprintf(ios.Out, "Current version: dev (built from source)\n")
	} else {
		fmt.Fprintf(ios.Out, "Current version: %s\n", currentVersion)
	}
	fmt.Fprintf(ios.Out, "Latest version:  %s\n", release.Version())

	// Download, verify checksum, and replace binary
	spinner.Start("Downloading and installing...")
	updateCtx, updateCancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer updateCancel()

	exe, err := selfupdate.ExecutablePath()
	if err != nil {
		spinner.Stop("")
		return fmt.Errorf("locate executable: %w", err)
	}

	err = updater.UpdateTo(updateCtx, release, exe)
	spinner.Stop("")
	if err != nil {
		return fmt.Errorf("update failed: %w", err)
	}

	fmt.Fprintf(ios.Out, "Updated magecli to %s\n", release.Version())
	return nil
}
