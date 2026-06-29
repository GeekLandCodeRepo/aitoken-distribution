package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/joho/godotenv"
	"github.com/urfave/cli/v3"

	"llm-gateway/internal/shared/config"
	"llm-gateway/internal/shared/database"
	"llm-gateway/internal/shared/migration"
)

func main() {
	var envFile string
	var skipEnv bool

	cmd := &cli.Command{
		Name:  "aitsd-migrate",
		Usage: "Run AiToken database schema migrations",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "env-file",
				Aliases:     []string{"e"},
				Value:       ".env",
				Usage:       "environment file to load before reading database config",
				Destination: &envFile,
			},
			&cli.BoolFlag{
				Name:        "skip-env",
				Usage:       "skip loading an environment file",
				Destination: &skipEnv,
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			if !skipEnv {
				if err := godotenv.Load(envFile); err != nil {
					slog.Info("no env file loaded, using system environment variables", "env_file", envFile)
				}
			}

			cfg := config.Load()
			if err := database.EnsureDatabase(cfg); err != nil {
				return fmt.Errorf("ensure database exists: %w", err)
			}
			if err := database.Init(cfg); err != nil {
				return fmt.Errorf("initialize database: %w", err)
			}
			defer database.GetDB().Close()

			if err := migration.Up(database.GetDB(), cfg); err != nil {
				return fmt.Errorf("apply migrations: %w", err)
			}
			slog.Info("database migrations applied successfully")
			return nil
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		slog.Error("migration failed", "error", err)
		os.Exit(1)
	}
}
