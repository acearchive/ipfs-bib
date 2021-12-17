package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var (
	configPath string
	outputPath string
	exportCar  bool
	pinSources bool
	jsonOutput bool

	rootCmd = &cobra.Command{
		Use:                   "ipfs-bib [options] <bibtex_file>",
		Short:                 "A tool for hosting bibliographic references on IPFS",
		Args:                  cobra.ExactArgs(1),
		Version:               "0.1.0",
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) error {
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
	rootCmd.Flags().BoolVar(&exportCar, "car", false, "Rather than add the sources to an IPFS node, export them as a CAR.")
	rootCmd.Flags().BoolVar(&pinSources, "pin", false, "Pin the source files when adding them to the IPFS node.")
	rootCmd.Flags().BoolVar(&jsonOutput, "json", false, "Produce machine-readable JSON output.")
}
