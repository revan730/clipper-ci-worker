package types

// StatusMessage represents response from Github Status API
type StatusMessage struct {
	State       string `json:"state"`
	Description string `json:"description"`
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
