package cmd

import (
	"fmt"
	"os"

	"github.com/revan730/clipper-ci-worker/log"
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

		logger := log.NewLogger(logVerbose)

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
		"Api gRPC port")
	startCmd.Flags().StringVarP(&rabbitAddr, "rabbitmq", "",
		"amqp://guest:guest@localhost:5672", "Set rabbitmq address")
	startCmd.Flags().StringVarP(&gcrURL, "gcr", "",
		"", "Set gcr url")
	startCmd.Flags().StringVarP(&jsonPath, "json", "",
		"secrets", "Set path to json auth file")
	startCmd.Flags().StringVarP(&dbAddr, "dbAddr", "",
		"postgres:5432", "Set PostsgreSQL address")
	startCmd.Flags().StringVarP(&db, "db", "",
		"clipper", "Set PostgreSQL database to use")
	startCmd.Flags().StringVarP(&dbUser, "user", "",
		"clipper", "Set PostgreSQL user to use")
	startCmd.Flags().StringVarP(&dbPass, "pass", "",
		"clipper", "Set PostgreSQL password to use")
	startCmd.Flags().StringVarP(&builderImage, "builder", "",
		"ci-builder", "Set docker builder image name")
	startCmd.Flags().BoolVarP(&logVerbose, "verbose", "v",
		false, "Show debug level logs",
	)
}
