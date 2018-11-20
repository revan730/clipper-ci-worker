package src

import (
	"os/exec"
	"strings"

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

func (w *Worker) executeBuilder(gitURL, branch, gcrHost, gcrTag string) ([]byte, error) {
	out, err := exec.Command("docker", "run", "-v", "/var/run/docker.sock:/var/run/docker.sock",
		"-v", w.config.JSONFile+":/opt/secrets/docker-login.json",
		"ci-builder", gitURL, branch, gcrHost, gcrTag).CombinedOutput()
	if err != nil {
		_, ok := err.(*exec.ExitError)
		if ok == true {
			w.logInfo("Process exited with non-zero code")
			return out, err
		}
		return out, err
	}
	return out, nil
}

func (w *Worker) executeCIJob(CIJob commonTypes.CIJob) {
	w.logInfo("Got CI job message:" + CIJob.RepoURL)
	gcrHost := strings.Split(w.config.GCRURL, "/")[0]
	repoName := strings.TrimSuffix(strings.TrimPrefix(CIJob.RepoURL, "https://github.com/"),
		".git")
	gcrTag := w.config.GCRURL + repoName
	repoURL := CIJob.RepoURL
	if CIJob.AccessToken != "" {
		repoURL = "https://" + strings.Split(strings.TrimPrefix(repoURL, "https://github.com/"), "/")[0] + ":" + CIJob.AccessToken + "@" + strings.TrimPrefix(CIJob.RepoURL, "https://")
	}
	out, err := w.executeBuilder(repoURL, CIJob.Branch, gcrHost, gcrTag)
	w.logInfo("Stdout:" + string(out))
	if err != nil {
		w.logError("Build failed", err)
		// TODO: Write log to db with failed status
	}
	// TODO: Write log to db and create CD job
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
