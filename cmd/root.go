package cmd

import (
	"context"
	"github.com/frawleyskid/ipfs-bib/archive"
	"github.com/frawleyskid/ipfs-bib/config"
	"github.com/frawleyskid/ipfs-bib/logging"
	"github.com/frawleyskid/ipfs-bib/store"
	"io/ioutil"
	"os"

	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:                   "ipfs-bib [options] <bibtex_file>",
		Short:                 "A tool for hosting bibliographic references on IPFS",
		Long:                  "A tool for hosting bibliographic references on IPFS.\n\nThis command accepts the path of a bibtex/biblatex file, or `-` to read from stdin.\nIf --zotero is passed, this accepts a Zotero group ID instead.",
		Args:                  cobra.ExactArgs(1),
		Version:               "0.1.0",
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			cfg, err := config.Load(cmd.Flags())
			if err != nil {
				return err
			}

			if err := cfg.Flags.Validate(); err != nil {
				return err
			}

			if !cfg.Flags.Verbose {
				logging.Verbose.SetOutput(ioutil.Discard)
			}

			bibChan, contentsChan := archive.Load(ctx, cfg, args[0])

			sourceStore, err := store.SourceStoreFromConfig(ctx, cfg)
			if err != nil {
				return err
			}

			location, metadata, err := archive.Store(ctx, cfg, contentsChan, sourceStore)
			if err != nil {
				return err
			}

			if cfg.Flags.MaybeOutputPath() != nil {
				bibResult := <-bibChan
				if bibResult.Error != nil {
					return bibResult.Error
				}

				if err := archive.UpdateBib(bibResult.Bib, cfg.File.Ipfs.MaybeGateway(), location); err != nil {
					return err
				}

				if err := archive.WriteBib(bibResult.Bib, cfg.Flags.OutputPath); err != nil {
					return err
				}
			}

			output, err := archive.NewOutput(cfg, metadata, location)
			if err != nil {
				return err
			}

			if cfg.Flags.JsonOutput {
				output.JsonPrint()
			} else {
				output.PrettyPrint()
			}

			return nil
		},
	}
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.SetVersionTemplate("ipfs-bib {{ .Version }}\n")
	rootCmd.Flags().StringP("config", "c", "", "The path of the config file to use. Otherwise, use the default config.")
	rootCmd.Flags().StringP("output", "o", "", "Generate a new bibtex file at this path with the IPFS URLs added to each entry.")
	rootCmd.Flags().String("car", "", "Rather than add the sources to an IPFS node, export them as a CAR archive at this path.")
	rootCmd.Flags().Bool("pin", false, "Pin the source files when adding them to the IPFS node.")
	rootCmd.Flags().String("pin-remote", "", "Pin the source files using each of the configured IPFS pinning services. Pass a name for the pin.")
	rootCmd.Flags().Bool("json", false, "Produce machine-readable JSON output.")
	rootCmd.Flags().Bool("zotero", false, "Pull references from a public Zotero library. Pass a Zotero group ID.")
	rootCmd.Flags().BoolP("verbose", "v", false, "Print verbose output.")
	rootCmd.Flags().Bool("dry-run", false, "Download sources, but don't add them to IPFS or export them as a CAR.")
	rootCmd.Flags().String("mfs", "", "Add the sources to MFS at this path.")
}
