package cmd

import (
	"context"
	"errors"
	"github.com/frawleyskid/ipfs-bib/archive"
	"github.com/frawleyskid/ipfs-bib/config"
	"github.com/frawleyskid/ipfs-bib/logging"
	"github.com/frawleyskid/ipfs-bib/store"
	"github.com/nickng/bibtex"
	"io/ioutil"
	"os"

	"github.com/spf13/cobra"
)

var ErrMfsAndCar = errors.New("can not add sources to MFS if exporting them as a CAR")

var (
	configPath    string
	outputPath    string
	carPath       string
	pinSources    bool
	jsonOutput    bool
	useZotero     bool
	verboseOutput bool
	dryRun        bool
	mfsPath       string

	rootCmd = &cobra.Command{
		Use:     "ipfs-bib [options] <bibtex_file>",
		Short:   "A tool for hosting bibliographic references on IPFS",
		Long:    "A tool for hosting bibliographic references on IPFS.\n\nThis command accepts the path of a bibtex/biblatex file, or `-` to read from stdin.\nIf --zotero is passed, this accepts a Zotero group ID instead.",
		Args:    cobra.ExactArgs(1),
		Version: "0.1.0",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if carPath != "" && mfsPath != "" {
				return ErrMfsAndCar
			}

			return nil
		},
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			if !verboseOutput {
				logging.Verbose.SetOutput(ioutil.Discard)
			}

			var (
				cfg config.Config
				err error
			)

			if configPath == "" {
				cfg, err = config.LoadDefault()
				if err != nil {
					return err
				}
			} else {
				cfg, err = config.FromToml(configPath)
				if err != nil {
					return err
				}
			}

			var bib bibtex.BibTex

			contentsChan := archive.NewDownloadResultChan()

			if useZotero {
				go archive.FromZotero(ctx, cfg, args[0], contentsChan)
				if err != nil {
					return err
				}
			} else {
				bib, err = archive.ParseBibtex(args[0])
				if err != nil {
					return err
				}

				go archive.FromBibtex(ctx, cfg, bib, contentsChan)
				if err != nil {
					return err
				}
			}

			var (
				location archive.Location
				metadata []archive.BibMetadata
			)

			switch {
			case dryRun:
				location, metadata, err = archive.ToNowhere(ctx, cfg, contentsChan)
				if err != nil {
					return err
				}
			case carPath == "":
				location, metadata, err = archive.ToNode(ctx, cfg, pinSources, contentsChan)
				if err != nil {
					return err
				}
			default:
				location, metadata, err = archive.ToCar(ctx, cfg, carPath, contentsChan)
				if err != nil {
					return err
				}
			}

			if useZotero {
				bib = archive.MetadataToBibtex(metadata)
			}

			if mfsPath != "" {
				if err := store.AddToMfs(ctx, cfg.Ipfs.Api, location.Root, mfsPath); err != nil {
					return err
				}
			}

			if outputPath != "" {
				if cfg.Ipfs.UseGateway {
					if err := archive.UpdateBib(bib, &cfg.Ipfs.Gateway, location); err != nil {
						return err
					}
				} else {
					if err := archive.UpdateBib(bib, nil, location); err != nil {
						return err
					}
				}

				if err := archive.WriteBib(bib, outputPath); err != nil {
					return err
				}
			}

			output, err := archive.NewOutput(cfg, metadata, location)
			if err != nil {
				return err
			}

			if jsonOutput {
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
	rootCmd.Flags().StringVarP(&configPath, "config", "c", "", "The path of the config file to use. Otherwise, use the default config.")
	rootCmd.Flags().StringVarP(&outputPath, "output", "o", "", "Generate a new bibtex file at this path with the IPFS URLs added to each entry.")
	rootCmd.Flags().StringVar(&carPath, "car", "", "Rather than add the sources to an IPFS node, export them as a CAR archive.")
	rootCmd.Flags().BoolVar(&pinSources, "pin", false, "Pin the source files when adding them to the IPFS node.")
	rootCmd.Flags().BoolVar(&jsonOutput, "json", false, "Produce machine-readable JSON output.")
	rootCmd.Flags().BoolVar(&useZotero, "zotero", false, "Pull references from a public Zotero library. Pass a Zotero group ID.")
	rootCmd.Flags().BoolVarP(&verboseOutput, "verbose", "v", false, "Print verbose output.")
	rootCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Download sources, but don't add them to IPFS or export them as a CAR.")
	rootCmd.Flags().StringVar(&mfsPath, "mfs", "", "Add the sources to MFS at this path.")
}
