package src

import (
	"github.com/revan730/clipper-ci-worker/api"
	"github.com/revan730/clipper-ci-worker/db"
	"github.com/revan730/clipper-ci-worker/queue"
	"github.com/revan730/clipper-ci-worker/types"
	"go.uber.org/zap"
)

// Worker holds CI worker logic
type Worker struct {
	config         *Config
	jobsQueue      *queue.Queue
	databaseClient db.DatabaseClient
	apiServer *api.Server
	logger         *zap.Logger
}

// NewWorker creates new copy of worker with provided
// config and rabbitmq client
func NewWorker(config *Config, logger *zap.Logger) *Worker {
	worker := &Worker{
		config: config,
		logger: logger,
	}
	dbConfig := types.PGClientConfig{
		DBUser:     config.DBUser,
		DBAddr:     config.DBAddr,
		DBPassword: config.DBPassword,
		DB:         config.DB,
	}
	dbClient := db.NewPGClient(dbConfig)
	worker.jobsQueue = queue.NewQueue(config.RabbitAddress)
	worker.databaseClient = dbClient
	apiConfig := api.Config{
		Port: config.Port,
	}
	apiServer := api.NewServer(apiConfig, logger, dbClient)
	worker.apiServer = apiServer
	return worker
}

func (w *Worker) logFatal(msg string, err error) {
	defer w.logger.Sync()
	w.logger.Fatal(msg, zap.Error(err))
}

func (w *Worker) logError(msg string, err error) {
	defer w.logger.Sync()
	w.logger.Error(msg, zap.String("packageLevel", "core"), zap.Error(err))
}

func (w *Worker) logInfo(msg string) {
	defer w.logger.Sync()
	w.logger.Info("INFO", zap.String("msg", msg), zap.String("packageLevel", "core"))
}

// Run starts CI worker
func (w *Worker) Run() {
	w.apiServer.Run()
	w.startConsuming()
}
