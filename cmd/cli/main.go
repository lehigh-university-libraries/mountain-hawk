// cmd/cli/main.go
package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/charmbracelet/fang"
	"github.com/joho/godotenv"
	"github.com/lehigh-university-libraries/mountain-hawk/cmd"
)

func main() {
	if err := godotenv.Load(); err != nil {
		slog.Error("Error loading .env file", "err", err)
		os.Exit(1)
	}

	rootCmd := cmd.NewRootCommand()

	if err := fang.Execute(context.Background(), rootCmd); err != nil {
		os.Exit(1)
	}
}
