package types

import "time"

// Build represents CI job process with GCR container push
type Build struct {
	ID            int64     `json:"buildID"`
	GithubRepoID  int64     `json:"-"`
	IsSuccessfull bool      `json:"success"`
	Date          time.Time `json:"date"`
	Branch        string    `json:"branch"`
	Stdout        string    `json:"stdout"`
}

// BuildArtifact represents docker container pushed to GCR after CI process
type BuildArtifact struct {
	ID      int64 `json:"artifactID"`
	BuildID int64 `json:"-"`
	// Name is a complete container name, with version
	Name string `json:"name"`
}

// StatusMessage represents response from Github Status API
type StatusMessage struct {
	State       string `json:"state"`
	Description string `json:"description"`
	Context     string `json:"context"`
}

// BuilderPayload is used to provide environment args to docker builder
type BuilderPayload struct {
	RepoURL  string
	RepoName string
	Branch   string
	GCRHost  string
	GCRTag   string
	Username string
}

type PGClientConfig struct {
	DBAddr     string
	DB         string
	DBUser     string
	DBPassword string
}
