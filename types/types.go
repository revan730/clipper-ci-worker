package types

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
