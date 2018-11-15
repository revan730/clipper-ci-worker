package cmd

import (
	"fmt"
	"os"

	"github.com/revan730/clipper-ci-worker/src"
	"github.com/spf13/cobra"
	"github.com/streadway/amqp"
)

var (
	rabbitAddr string
	rabbitQ    string
	logVerbose bool
)

func connectToRabbit(addr string) *amqp.Connection {
	connection, err := amqp.Dial(addr)
	if err != nil {
		panic(fmt.Sprintf("Couldn't connect to rabbitmq: %s", err))
	}

	return connection
}

var RootCmd = &cobra.Command{
	Use:   "clipper-ci",
	Short: "CI worker microservice of Clipper CI\\CD",
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start worker",
	Run: func(cmd *cobra.Command, args []string) {
		config := &src.Config{
			RabbitAddress: rabbitAddr,
			RabbitQueue:   rabbitQ,
			Verbose:       logVerbose,
		}

		logger := src.NewLogger(logVerbose)

		rabbitConnection := connectToRabbit(config.RabbitAddress)
		worker := src.NewWorker(config, rabbitConnection, logger)
		worker.Run()
	},
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	RootCmd.AddCommand(startCmd)
	startCmd.Flags().StringVarP(&rabbitAddr, "rabbitmq", "r",
		"amqp://guest:guest@localhost:5672", "Set redis address")
	startCmd.Flags().StringVarP(&rabbitQ, "queue", "q",
		"ciJobs", "Set rabbitmq queue name")
	startCmd.Flags().BoolVarP(&logVerbose, "verbose", "v",
		false, "Show debug level logs",
	)
}
