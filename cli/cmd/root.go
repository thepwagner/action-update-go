package cmd

import (
	"fmt"
	"os"

	"github.com/mitchellh/go-homedir"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile     string
	updaterType string
	logLevel    string
)

const (
	flagConfig      = "config"
	flagUpdaterType = "updater"
	flagLogLevel    = "log"

	cfgGitHubToken = "GitHubToken"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "action-update-go",
	Short: "Action-update-go",
	Long:  `Simulates GitHub Actions environment enough to test action-update-go.`,
	PersistentPreRunE: func(*cobra.Command, []string) error {
		level, err := logrus.ParseLevel(logLevel)
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

	rootCmd.PersistentFlags().StringVar(&cfgFile, flagConfig, "", "config file (default is $HOME/.action-update-go.yaml)")

	rootCmd.PersistentFlags().StringVar(&updaterType, flagUpdaterType, "go", "updater to use")
	_ = viper.BindPFlag(flagUpdaterType, rootCmd.PersistentFlags().Lookup(flagUpdaterType))

	rootCmd.PersistentFlags().StringVar(&logLevel, flagLogLevel, logrus.InfoLevel.String(), "log level")
	_ = viper.BindPFlag(flagLogLevel, rootCmd.PersistentFlags().Lookup(flagLogLevel))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := homedir.Dir()
		if err == nil {
			viper.AddConfigPath(home)
			viper.SetConfigName(".action-update-go")
		}
	}

	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err == nil {
		logrus.WithField("cfg", viper.ConfigFileUsed()).Debug("using config file")
	}
}
