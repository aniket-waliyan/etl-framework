package main

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "etl-cli",
		Short: "ETL Pipeline Framework CLI",
		Long:  "Command line tool for generating and managing ETL pipelines",
	}

	var generateCmd = &cobra.Command{
		Use:   "generate",
		Short: "Generate a new ETL pipeline",
		Run:   runGenerate,
	}

	generateCmd.Flags().String("name", "", "Name of the pipeline to generate")
	generateCmd.MarkFlagRequired("name")

	rootCmd.AddCommand(generateCmd)

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func runGenerate(cmd *cobra.Command, args []string) {
	name, _ := cmd.Flags().GetString("name")
	generator := NewGenerator(name)
	if err := generator.Generate(); err != nil {
		fmt.Printf("Failed to generate pipeline: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Successfully generated pipeline: %s\n", name)
}
