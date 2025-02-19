package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/urth-inc/s3-cp-tagging/internal/config"
	"github.com/urth-inc/s3-cp-tagging/internal/uploader"
)

var (
	// Version information (set at build time)
	Version   = "dev"
	Revision  = "unknown"
	BuildDate = "unknown"
)

// Variable to allow mocking os.Exit in tests
var osExit = os.Exit

func newRootCmd() *cobra.Command {
	cfg := &config.Config{}

	rootCmd := &cobra.Command{
		Use:     "s3-cp-tagging SOURCE s3://BUCKET",
		Short:   "Copy a local file or directory to S3 with tags",
		Long:    `Copy a local file or directory to S3 with tags.
Similar to 's3 cp' command but with tagging support.`,
		Version: Version,
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg.SourcePath = args[0]
			cfg.Bucket = strings.TrimPrefix(args[1], "s3://")

			if err := cfg.Validate(); err != nil {
				return err
			}

			u, err := uploader.New(cfg)
			if err != nil {
				return err
			}

			fmt.Printf("upload: %s to s3://%s\n", cfg.SourcePath, cfg.Bucket)
			if err := u.UploadDirectory(); err != nil {
				return err
			}
			return nil
		},
		Example: `  # Upload a directory with tags
  s3-cp-tagging ./dist s3://my-bucket --tagging "TagSet=[{Key=version,Value=v1.0.0},{Key=commit,Value=abc123}]"

  # Use a specific AWS profile
  s3-cp-tagging ./dist s3://my-bucket --profile myprofile --tagging "TagSet=[{Key=version,Value=v1.0.0}]"`,
	}

	rootCmd.SetVersionTemplate(`Version: {{.Version}}
Revision: ` + Revision + `
BuildDate: ` + BuildDate + "\n")

	rootCmd.Flags().StringVar(&cfg.Profile, "profile", "", "Use a specific profile from your credential file")
	rootCmd.Flags().StringVar(&cfg.Tagging, "tagging", "", `The tag-set for the object. The tag-set must be formatted as: "TagSet=[{Key=key1,Value=value1},{Key=key2,Value=value2}]"`)

	return rootCmd
}

func main() {
	if err := newRootCmd().Execute(); err != nil {
		osExit(1)
	}
}