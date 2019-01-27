package db

import (
	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
	"github.com/revan730/clipper-ci-worker/types"
)

// PostgresClient provides data access layer to objects in Postgres.
// implements DatabaseClient interface
type PostgresClient struct {
	pg *pg.DB
}

// NewPGClient creates new copy of PostgresClient
func NewPGClient(config types.PGClientConfig) *PostgresClient {
	DBClient := &PostgresClient{}
	pgdb := pg.Connect(&pg.Options{
		User:         config.DBUser,
		Addr:         config.DBAddr,
		Password:     config.DBPassword,
		Database:     config.DB,
		MinIdleConns: 2,
	})
	DBClient.pg = pgdb
	return DBClient
}

// Close gracefully closes db connection
func (d *PostgresClient) Close() {
	d.pg.Close()
}

// CreateSchema creates database tables if they not exist
func (d *PostgresClient) CreateSchema() error {
	for _, model := range []interface{}{
		(*types.Build)(nil),
		(*types.BuildArtifact)(nil)} {
		err := d.pg.CreateTable(model, &orm.CreateTableOptions{
			IfNotExists:   true,
			FKConstraints: true,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// CreateBuild creates repo build record from provided struct
func (d *PostgresClient) CreateBuild(b *types.Build) error {
	return d.pg.Insert(b)
}

// FindAllBuilds returns all builds for provided repo id
// with pagination support (by passing query params of request)
// TODO: don't select stdouts here?
func (d *PostgresClient) FindAllBuilds(repoID int64, branch string, page, limit int64) ([]*types.Build, error) {
	var builds []*types.Build
	offset := int((page - 1) * limit)

	err := d.pg.Model(&builds).
		Where("github_repo_id = ?", repoID).
		Where("branch = ?", branch).
		Limit(int(limit)).
		Offset(offset).
		Select()

	return builds, err
}

func (d *PostgresClient) FindBuildsCount(repoID int64, branch string) (int64, error) {
	count, err := d.pg.Model(&types.Build{}).
		Where("github_repo_id = ?", repoID).
		Where("branch = ?", branch).
		Count()
	return int64(count), err
}

// FindBuildByID finds build with provided id
func (d *PostgresClient) FindBuildByID(buildID int64) (*types.Build, error) {
	build := &types.Build{
		ID: buildID,
	}

	err := d.pg.Select(build)
	if err != nil {
		if err == pg.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return build, nil
}

// CreateBuildArtifact creates build artifact record from provided struct
func (d *PostgresClient) CreateBuildArtifact(b *types.BuildArtifact) error {
	return d.pg.Insert(b)
}

// FindBuildArtifact returns build artifact for provided build id
func (d *PostgresClient) FindBuildArtifact(buildID int64) (*types.BuildArtifact, error) {
	buildArtifact := &types.BuildArtifact{}

	err := d.pg.Model(buildArtifact).
		Where("build_id = ?", buildID).
		Select()

	return buildArtifact, err
}

// FindBuildArtifactByID returns build artifact with provided id
func (d *PostgresClient) FindBuildArtifactByID(ID int64) (*types.BuildArtifact, error) {
	buildArtifact := &types.BuildArtifact{
		ID: ID,
	}

	err := d.pg.Select(buildArtifact)
	if err != nil {
		if err == pg.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return buildArtifact, nil
}
