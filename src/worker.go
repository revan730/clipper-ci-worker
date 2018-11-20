package src

import (
	"os/exec"
	"strings"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/revan730/clipper-common/db"
	commonTypes "github.com/revan730/clipper-common/types"
	"github.com/streadway/amqp"
	"go.uber.org/zap"
)

type Worker struct {
	config           *Config
	rabbitConnection *amqp.Connection
	databaseClient   *db.DatabaseClient
	logger           *zap.Logger
}

// NewWorker creates new copy of worker with provided
// config and rabbitmq client
func NewWorker(config *Config, rabbitConnection *amqp.Connection, logger *zap.Logger) *Worker {
	worker := &Worker{
		config:           config,
		rabbitConnection: rabbitConnection,
		logger:           logger,
	}
	dbConfig := commonTypes.DBClientConfig{
		DBUser:     config.DBUser,
		DBAddr:     config.DBAddr,
		DBPassword: config.DBPassword,
		DB:         config.DB,
	}
	dbClient := db.NewDBClient(dbConfig)
	worker.databaseClient = dbClient
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

// TODO: Move builder image name to config
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

func (w *Worker) writeBuildToDB(repoID int64, success bool, branch, stdout, gcrTag string) error {
	build := commonTypes.Build{
		GithubRepoID:  repoID,
		IsSuccessfull: success,
		Date:          time.Now(),
		Branch:        branch,
		Stdout:        stdout,
	}
	err := w.databaseClient.CreateBuild(&build)
	if err != nil {
		return err
	}
	artifact := commonTypes.BuildArtifact{
		BuildID: build.ID,
		Name:    gcrTag,
	}
	err = w.databaseClient.CreateBuildArtifact(&artifact)
	return err
}

// TODO: Remove debug logs
// TODO: Refactor strings with sprintf function
func (w *Worker) executeCIJob(CIJob commonTypes.CIJob) {
	w.logInfo("Got CI job message:" + CIJob.RepoURL)
	gcrHost := strings.Split(w.config.GCRURL, "/")[0]
	repoName := strings.TrimSuffix(strings.TrimPrefix(CIJob.RepoURL, "https://github.com/"),
		".git")
	gcrTag := w.config.GCRURL + repoName + ":" + CIJob.Branch + "-" + CIJob.HeadSHA[:7]
	repoURL := CIJob.RepoURL
	username := strings.Split(strings.TrimPrefix(repoURL, "https://github.com/"), "/")[0]
	if CIJob.AccessToken != "" {
		repoURL = "https://" + username + ":" + CIJob.AccessToken + "@" + strings.TrimPrefix(CIJob.RepoURL, "https://")
	}
	out, err := w.executeBuilder(repoURL, CIJob.Branch, gcrHost, gcrTag)
	w.logInfo("Stdout:" + string(out))
	success := true
	if err != nil {
		w.logError("Build failed", err)
		success = false
	}
	err = w.writeBuildToDB(CIJob.RepoID, success, CIJob.Branch, string(out), gcrTag)
	if err != nil {
		w.logError("Write build log to db failed", err)
	}
	// TODO: Call github status api and create CD job if needed
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
