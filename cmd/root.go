package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/harvey-earth/pocdeploy/internal"
)

var cfgFile string

// Root returns the root command to create manpages
func Root() *cobra.Command {
	return rootCmd
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "pocdeploy",
	Short: "A POC deployment tool for Kubernetes",
	Long: `A POC deployment tool for Kind.
This tool deploys a frontend image with a Postgres backend using CloudNative PG.`,
	Version: "0.0.1",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Config File
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/pocdeploy.yaml)")
	// Cluster Type
	rootCmd.PersistentFlags().StringP("type", "t", "kind", "Type of cluster(kind)")
	viper.BindPFlag("type", rootCmd.PersistentFlags().Lookup("type"))

	// Verbose Levels
	rootCmd.PersistentFlags().BoolP("debug", "d", false, "debug output")
	viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose output")
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	rootCmd.PersistentFlags().BoolP("quiet", "q", false, "no output")
	viper.BindPFlag("quiet", rootCmd.PersistentFlags().Lookup("quiet"))
	rootCmd.MarkFlagsMutuallyExclusive("debug", "verbose", "quiet")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name "pocdeploy" (without extension).
		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		viper.SetConfigType("yaml")
		viper.SetConfigName("pocdeploy")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		if q := viper.GetBool("quiet"); !q {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
		}
	}

	// Initialize Logger
	if err := internal.InitLogger(); err != nil {
		panic(err)
	}
}
