package db

import (
	"net/url"

	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
	"github.com/revan730/clipper-common/types"
)

// PostgresClient provides data access layer to objects in Postgres.
// implements DatabaseClient interface
type PostgresClient struct {
	pg         *pg.DB
	adminLogin string
	adminPass  string
}

// NewPGClient creates new copy of PostgresClient
func NewPGClient(config types.PGClientConfig) *PostgresClient {
	DBClient := &PostgresClient{
		adminLogin: config.AdminLogin,
		adminPass:  config.AdminPassword,
	}
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
func (d *PostgresClient) FindAllBuilds(repoID int64, q url.Values) ([]types.Build, error) {
	var builds []types.Build

	err := d.pg.Model(&builds).
		Apply(orm.Pagination(q)).
		Where("github_repo_id = ?", repoID).
		Select()

	return builds, err
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
