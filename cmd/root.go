package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/frawleyskid/ipfs-bib/archive"
	"github.com/frawleyskid/ipfs-bib/config"
	"os"

	"github.com/spf13/cobra"
)

var (
	configPath string
	outputPath string
	carPath    string
	pinSources bool
	jsonOutput bool

	rootCmd = &cobra.Command{
		Use:                   "ipfs-bib [options] <bibtex_file>",
		Short:                 "A tool for hosting bibliographic references on IPFS",
		Args:                  cobra.ExactArgs(1),
		Version:               "0.1.0",
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			var (
				cfg *config.Config
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

			bib, err := archive.ParseBibtex(args[0])
			if err != nil {
				return err
			}

			contents, err := archive.Download(ctx, cfg, bib)
			if err != nil {
				return err
			}

			var location *archive.Location

			if carPath == "" {
				location, err = archive.ToNode(ctx, cfg, pinSources, contents)
				if err != nil {
					return err
				}
			} else {
				location, err = archive.ToCar(ctx, cfg, carPath, contents)
				if err != nil {
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

			if jsonOutput {
				output, err := location.ToOutput(cfg)
				if err != nil {
					return err
				}

				marshalledOutput, err := json.Marshal(output)
				if err != nil {
					return err
				}

				fmt.Println(string(marshalledOutput))
			} else {
				fmt.Println(location.Root.String())
			}

			return nil
		},
	}
)

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
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
}
