package db

import (
	"net/url"

	"github.com/revan730/clipper-ci-worker/types"
)

// DatabaseClient provides interface for data access layer operations
type DatabaseClient interface {
	Close()
	CreateSchema() error
	CreateBuild(b *types.Build) error
	FindAllBuilds(repoID int64, q url.Values) ([]types.Build, error)
	CreateBuildArtifact(b *types.BuildArtifact) error
	FindBuildArtifact(buildID int64) (*types.BuildArtifact, error)
	FindBuildByID(buildID int64) (*types.Build, error)
}
