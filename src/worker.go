package src

import (
	"github.com/revan730/clipper-ci-worker/api"
	"github.com/revan730/clipper-ci-worker/db"
	"github.com/revan730/clipper-ci-worker/queue"
	"github.com/revan730/clipper-ci-worker/types"
	"github.com/revan730/clipper-ci-worker/log"
)

// Worker holds CI worker logic
type Worker struct {
	config         *Config
	jobsQueue      queue.Queue
	databaseClient db.DatabaseClient
	apiServer *api.Server
	log         log.Logger
}

// NewWorker creates new copy of worker with provided
// config and rabbitmq client
func NewWorker(config *Config, logger log.Logger) *Worker {
	worker := &Worker{
		config: config,
		log: logger,
	}
	dbConfig := types.PGClientConfig{
		DBUser:     config.DBUser,
		DBAddr:     config.DBAddr,
		DBPassword: config.DBPassword,
		DB:         config.DB,
	}
	dbClient := db.NewPGClient(dbConfig)
	worker.jobsQueue = queue.NewRMQQueue(config.RabbitAddress)
	worker.databaseClient = dbClient
	apiConfig := api.Config{
		Port: config.Port,
	}
	apiServer := api.NewServer(apiConfig, logger, dbClient)
	worker.apiServer = apiServer
	return worker
}

// Run starts CI worker
func (w *Worker) Run() {
	w.apiServer.Run()
	w.startConsuming()
}
