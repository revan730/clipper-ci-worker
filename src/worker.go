package src

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/revan730/clipper-ci-worker/types"
	"github.com/revan730/clipper-common/db"
	commonTypes "github.com/revan730/clipper-common/types"
	"github.com/streadway/amqp"
	"go.uber.org/zap"
)

type Worker struct {
	config           *Config
	rabbitConnection *amqp.Connection
	rabbitChannel    *amqp.Channel
	CDQueue          amqp.Queue
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

// TODO: Refactor & remove old debug info
func (w *Worker) writeGithubStatus(user, accessToken, repo, sha string, success bool) error {
	client := &http.Client{}
	url := fmt.Sprintf("https://api.github.com/repos/%s/statuses/%s",
		repo, sha)
	w.logInfo("status url:" + url)
	body := &types.StatusMessage{
		Description: "Status set by Clipper CI\\CD",
	}
	if success == true {
		body.State = "success"
	} else {
		body.State = "failure"
	}
	rawBody, err := json.Marshal(body)
	if err != nil {
		return err
	}
	buf := bytes.NewBuffer(rawBody)
	req, err := http.NewRequest("POST", url, buf)
	if err != nil {
		return err
	}
	req.SetBasicAuth(user, accessToken)
	resp, err := client.Do(req)
	defer resp.Body.Close()
	if err != nil {
		return err
	}
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	w.logInfo("Github response body:" + string(respBody))
	return nil
}

// TODO: Move all rabbit code to separate package
func (w *Worker) postCDJob(repoID int64, branch, gcrTag string) error {
	msg := &commonTypes.CDJob{
		RepoID: repoID,
		Branch: branch,
		GcrTag: gcrTag,
	}
	body, err := proto.Marshal(msg)
	if err != nil {
		return err
	}
	return w.rabbitChannel.Publish(
		"", w.CDQueue.Name, false, false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(body),
		})
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
		return
	}
	err = w.writeGithubStatus(username, CIJob.AccessToken, repoName, CIJob.HeadSHA, success)
	if err != nil {
		w.logError("Write Github status failed", err)
	}
	if success == false {
		return
	}
	err = w.postCDJob(CIJob.RepoID, CIJob.Branch, gcrTag)
	if err != nil {
		w.logError("Post CD job failed", err)
	}
}

func (w *Worker) startConsuming() {
	defer w.rabbitConnection.Close()

	ch, err := w.rabbitConnection.Channel()
	if err != nil {
		w.logFatal("Failed to open channel", err)
	}
	w.rabbitChannel = ch

	q, err := ch.QueueDeclare(w.config.RabbitQueue, false, false, false,
		false, nil)
	if err != nil {
		w.logFatal("Failed to declare queue", err)
	}

	cdQueue, err := ch.QueueDeclare(w.config.CDQueue, false, false, false,
		false, nil)
	if err != nil {
		w.logFatal("Failed to declare CD jobs queue", err)
	}

	w.CDQueue = cdQueue

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
