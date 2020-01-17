package cmd

import (
	"fmt"
	"os"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/zdyxry/sparrow/pkg"
)

// rootCmd represents the base command when called without any subcommands
var (
	rootCmd = &cobra.Command{
		Use:   "sparrow",
		Short: "sparrow command line",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	serveCmd = &cobra.Command{
		Use:   "serve",
		Short: "run sparrow service",
		Run:   serveCommandFunc,
	}

	cfgFile    string
	logFile    string
	configOpts pkg.Config
)

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.AddCommand(serveCmd)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.sparrow.toml)")
	rootCmd.PersistentFlags().StringVar(&logFile, "log", "", "log path")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	viper.SetConfigType("toml")

	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".sparrow" (without extension).
		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		viper.AddConfigPath("/etc/sparrow/")

		viper.SetConfigName("sparrow")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
	viper.Unmarshal(&configOpts)
}

func serveCommandFunc(cmd *cobra.Command, args []string) {
	pkg.Serve(configOpts)
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
