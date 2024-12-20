package main

import (
	"fmt"
	"os"

	"github.com/aniketwaliyan/etl-framework/internal/utils/config"
	"github.com/spf13/cobra"
)

func runValidate(cmd *cobra.Command, args []string) {
	configPath, _ := cmd.Flags().GetString("config")
	if err := validateConfig(configPath); err != nil {
		fmt.Printf("Configuration validation failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Configuration is valid!")
}

func validateConfig(configPath string) error {
	parser := config.NewParser()
	_, err := parser.Parse(configPath)
	return err
}
