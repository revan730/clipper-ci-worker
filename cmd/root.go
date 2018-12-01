package cmd

import (
	"fmt"
	"os"

	"github.com/revan730/clipper-ci-worker/src"
	"github.com/spf13/cobra"
)

var (
	serverPort   int
	rabbitAddr   string
	gcrURL       string
	jsonPath     string
	dbAddr       string
	db           string
	dbUser       string
	dbPass       string
	builderImage string
	logVerbose   bool
)

var rootCmd = &cobra.Command{
	Use:   "clipper-ci",
	Short: "CI worker microservice of Clipper CI\\CD",
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start worker",
	Run: func(cmd *cobra.Command, args []string) {
		config := &src.Config{
			Port:          serverPort,
			RabbitAddress: rabbitAddr,
			GCRURL:        gcrURL,
			JSONFile:      jsonPath,
			DBAddr:        dbAddr,
			DB:            db,
			DBUser:        dbUser,
			DBPassword:    dbPass,
			Verbose:       logVerbose,
			BuilderImage:  builderImage,
		}

		logger := src.NewLogger(logVerbose)

		worker := src.NewWorker(config, logger)
		worker.Run()
	},
}

// Execute runs application with provided cli params
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	// TODO: Remove short flags
	rootCmd.AddCommand(startCmd)
	startCmd.Flags().IntVarP(&serverPort, "port", "p", 8080,
		"Application TCP port")
	startCmd.Flags().StringVarP(&rabbitAddr, "rabbitmq", "r",
		"amqp://guest:guest@localhost:5672", "Set redis address")
	startCmd.Flags().StringVarP(&gcrURL, "gcr", "g",
		"", "Set gcr url")
	startCmd.Flags().StringVarP(&jsonPath, "json", "j",
		"secrets", "Set path to json auth file")
	startCmd.Flags().StringVarP(&dbAddr, "postgresAddr", "a",
		"postgres:5432", "Set PostsgreSQL address")
	startCmd.Flags().StringVarP(&db, "db", "d",
		"clipper", "Set PostgreSQL database to use")
	startCmd.Flags().StringVarP(&dbUser, "user", "u",
		"clipper", "Set PostgreSQL user to use")
	startCmd.Flags().StringVarP(&dbPass, "pass", "c",
		"clipper", "Set PostgreSQL password to use")
	startCmd.Flags().StringVarP(&builderImage, "builder", "b",
		"ci-builder", "Set docker builder image name")
	startCmd.Flags().BoolVarP(&logVerbose, "verbose", "v",
		false, "Show debug level logs",
	)
}
