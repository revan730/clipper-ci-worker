package src

import (
	"os/exec"

	"github.com/gogo/protobuf/proto"
	commonTypes "github.com/revan730/clipper-common/types"
	"github.com/streadway/amqp"
	"go.uber.org/zap"
)

type Worker struct {
	config           *Config
	rabbitConnection *amqp.Connection
	logger           *zap.Logger
}

// NewWorker creates new copy of worker with provided
// config and rabbitmq client
func NewWorker(config *Config, rabbitConnection *amqp.Connection, logger *zap.Logger) *Worker {
	return &Worker{
		config:           config,
		rabbitConnection: rabbitConnection,
		logger:           logger,
	}
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

func (w *Worker) executeDockerBuilder() error {
	out, err := exec.Command("docker", "run", "hello-world").CombinedOutput()
	if err != nil {
		_, ok := err.(*exec.ExitError)
		if ok == true {
			w.logInfo("Process exited with non-zero code")
			return nil
		}
		return err
	}
	w.logInfo("Stdout:" + string(out))
	return nil
}

func (w *Worker) executeCIJob(CIJob commonTypes.CIJob) {
	// TODO: Implement
	w.logInfo("Got CI job message:" + CIJob.RepoURL)
	w.executeDockerBuilder()
}

func (w *Worker) startConsuming() {
	defer w.rabbitConnection.Close()
	ch, err := w.rabbitConnection.Channel()

	if err != nil {
		w.logFatal("Failed to open channel", err)
	}

	q, err := ch.QueueDeclare(w.config.RabbitQueue, false, false, false,
		false, nil)

	if err != nil {
		w.logFatal("Failed to declare queue", err)
	}

	msgs, err := ch.Consume(q.Name, "", true, false, false, false, nil)

	blockMain := make(chan bool)

	go func() {
		for m := range msgs {
			w.logger.Info("Received message: ", zap.ByteString("body", m.Body))
			jobMsg := commonTypes.CIJob{}
			err := proto.Unmarshal(m.Body, &jobMsg)
			if err != nil {
				w.logError("Failed to unmarshal job message", err)
				continue
			}
			go w.executeCIJob(jobMsg)
		}
	}()

	w.logInfo("Worker started")
	<-blockMain
}

func (w *Worker) Run() {
	w.startConsuming()
}
