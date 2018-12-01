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

	"github.com/golang/protobuf/proto"
	"github.com/revan730/clipper-ci-worker/types"
	commonTypes "github.com/revan730/clipper-common/types"
	"go.uber.org/zap"
)

func (w *Worker) executeBuilder(payload types.BuilderPayload) ([]byte, error) {
	out, err := exec.Command("docker", "run", "-v", "/var/run/docker.sock:/var/run/docker.sock",
		"-v", w.config.JSONFile+":/opt/secrets/docker-login.json",
		w.config.BuilderImage, payload.RepoURL, payload.Branch,
		payload.GCRHost, payload.GCRTag).CombinedOutput()

	return out, err
}

func (w *Worker) writeBuildToDB(repoID int64, success bool, branch, stdout, gcrTag string) error {
	build := types.Build{
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
	artifact := types.BuildArtifact{
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
		Context:     "ci-build",
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
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	w.logInfo("Github response body:" + string(respBody))
	return nil
}

func (w *Worker) postCDJob(repoID int64, branch, gcrTag string) error {
	msg := &commonTypes.CDJob{
		RepoID: repoID,
		Branch: branch,
		GcrTag: gcrTag,
	}
	return w.jobsQueue.PublishCDJob(msg)
}

func (w *Worker) makeBuilderPayload(CIJob commonTypes.CIJob) types.BuilderPayload {
	payload := types.BuilderPayload{
		GCRHost: strings.Split(w.config.GCRURL, "/")[0],
		Branch:  CIJob.Branch,
	}
	payload.RepoName = strings.TrimSuffix(strings.TrimPrefix(CIJob.RepoURL, "https://github.com/"),
		".git")
	payload.GCRTag = fmt.Sprintf("%s%s:%s-%s", w.config.GCRURL, payload.RepoName,
		CIJob.Branch, CIJob.HeadSHA[:7])
	repoURL := CIJob.RepoURL
	payload.Username = strings.Split(strings.TrimPrefix(repoURL, "https://github.com/"), "/")[0]
	if CIJob.AccessToken != "" {
		repoURL = fmt.Sprintf("https://%s:%s@%s", payload.Username, CIJob.AccessToken,
			strings.TrimPrefix(CIJob.RepoURL, "https://"))
	}
	payload.RepoURL = repoURL
	return payload
}

// TODO: Remove debug logs
func (w *Worker) executeCIJob(CIJob commonTypes.CIJob) {
	w.logInfo("Got CI job message:" + CIJob.RepoURL)
	builderPayload := w.makeBuilderPayload(CIJob)
	out, err := w.executeBuilder(builderPayload)
	w.logInfo("Stdout:" + string(out))
	success := true
	if err != nil {
		w.logError("Build failed", err)
		success = false
	}
	err = w.writeBuildToDB(CIJob.RepoID, success, CIJob.Branch,
		string(out), builderPayload.GCRTag)
	if err != nil {
		w.logError("Write build log to db failed", err)
		return
	}
	err = w.writeGithubStatus(builderPayload.Username, CIJob.AccessToken,
		builderPayload.RepoName, CIJob.HeadSHA, success)
	if err != nil {
		w.logError("Write Github status failed", err)
	}
	if success == false {
		return
	}
	err = w.postCDJob(CIJob.RepoID, CIJob.Branch, builderPayload.GCRTag)
	if err != nil {
		w.logError("Post CD job failed", err)
	}
}

func (w *Worker) startConsuming() {
	defer w.jobsQueue.Close()
	blockMain := make(chan bool)

	ciMsgs, err := w.jobsQueue.MakeCIMsgChan()
	if err != nil {
		w.logFatal("Failed to create CI jobs channel", err)
	}

	go func() {
		for m := range ciMsgs {
			w.logger.Info("Received message: ", zap.ByteString("body", m))
			jobMsg := commonTypes.CIJob{}
			err := proto.Unmarshal(m, &jobMsg)
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
