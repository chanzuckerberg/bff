package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

var cfgFile string
var defaultBranchRef string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use: "bff",
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

	// Update HEAD to upstream default branch
	combinedOut, err := exec.Command("git", "remote", "set-head", "origin", "-a").CombinedOutput()
	if err != nil {
		fmt.Println("git remote set-head origin -a output: ", string(combinedOut))
		fmt.Println("error: ", err)
		os.Exit(1)
	}

	// Get the branch ref, will be used for bump & changelog commands
	cmdOutput, err := exec.Command("git", "symbolic-ref", "refs/remotes/origin/HEAD").Output()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defaultBranchRef = strings.TrimSpace(string(cmdOutput))
	// // Here you will define your flags and configuration settings.
	// // Cobra supports persistent flags, which, if defined here,
	// // will be global for your application.
	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.bff.yaml)")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	// if cfgFile != "" {
	// 	// Use config file from the flag.
	// 	viper.SetConfigFile(cfgFile)
	// } else {
	// 	// Find home directory.
	// 	home, err := homedir.Dir()
	// 	if err != nil {
	// 		fmt.Println(err)
	// 		os.Exit(1)
	// 	}

	// 	// Search config in home directory with name ".bff" (without extension).
	// 	viper.AddConfigPath(home)
	// 	viper.SetConfigName(".bff")
	// }

	// viper.AutomaticEnv() // read in environment variables that match

	// // If a config file is found, read it in.
	// if err := viper.ReadInConfig(); err == nil {
	// 	fmt.Println("Using config file:", viper.ConfigFileUsed())
	// }
}
