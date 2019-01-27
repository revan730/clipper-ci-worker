package db

import (
	"github.com/revan730/clipper-ci-worker/types"
)

// DatabaseClient provides interface for data access layer operations
type DatabaseClient interface {
	Close()
	CreateSchema() error
	CreateBuild(b *types.Build) error
	FindAllBuilds(repoID int64, branch string, page, limit int64) ([]*types.Build, error)
	FindBuildsCount(repoID int64, branch string) (int64, error)
	CreateBuildArtifact(b *types.BuildArtifact) error
	FindBuildArtifact(buildID int64) (*types.BuildArtifact, error)
	FindBuildArtifactByID(ID int64) (*types.BuildArtifact, error)
	FindBuildByID(buildID int64) (*types.Build, error)
}
