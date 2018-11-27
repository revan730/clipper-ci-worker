package types

type StatusMessage struct {
	State       string `json:"state"`
	Description string `json:"description"`
}

type BuilderPayload struct {
	RepoURL  string
	RepoName string
	Branch   string
	GCRHost  string
	GCRTag   string
	Username string
}
