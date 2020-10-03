package cmd

import (
	"fmt"
	"os"

	"github.com/mitchellh/go-homedir"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

const (
	flagGitHubToken = "GitHubToken"
	flagLogLevel    = "LogLevel"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "action-update-go",
	Short: "Action-update-go",
	Long:  `Simulates GitHub Actions environment enough to test action-update-go.`,
	PersistentPreRunE: func(*cobra.Command, []string) error {
		viper.SetDefault(flagLogLevel, logrus.InfoLevel.String())
		level, err := logrus.ParseLevel(viper.GetString(flagLogLevel))
		if err != nil {
			return err
		}
		logrus.SetLevel(level)
		logrus.WithField("level", level).Debug("parsed log level")
		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.action-update-go.yaml)")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
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

		// Search config in home directory with name ".action-update-go" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".action-update-go")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
