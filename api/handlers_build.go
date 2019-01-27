package api

import (
	"context"
	"net/http"

	ptypes "github.com/golang/protobuf/ptypes"
	"github.com/revan730/clipper-ci-worker/types"
	commonTypes "github.com/revan730/clipper-common/types"
	"google.golang.org/grpc/status"
)

func buildToProto(build *types.Build) *commonTypes.Build {
	timestamp, _ := ptypes.TimestampProto(build.Date)
	return &commonTypes.Build{
		ID:            build.ID,
		GithubRepoID:  build.GithubRepoID,
		IsSuccessfull: build.IsSuccessfull,
		Date:          timestamp,
		Branch:        build.Branch,
		Stdout:        build.Stdout,
	}
}

func (s *Server) GetBuild(ctx context.Context, in *commonTypes.Build) (*commonTypes.Build, error) {
	build, err := s.databaseClient.FindBuildByID(in.ID)
	if err != nil {
		s.log.Error("Find build error", err)
		return &commonTypes.Build{}, status.New(http.StatusInternalServerError, "").Err()
	}
	if build == nil {
		return &commonTypes.Build{}, status.New(http.StatusNotFound, "build not found").Err()
	}
	return buildToProto(build), nil
}

func (s *Server) GetAllBuilds(ctx context.Context, in *commonTypes.BuildsQuery) (*commonTypes.BuildsArray, error) {
	builds, err := s.databaseClient.FindAllBuilds(in.RepoID, in.Branch, in.Page, in.Limit)
	if err != nil {
		s.log.Error("Find all builds error", err)
		return &commonTypes.BuildsArray{}, status.New(http.StatusInternalServerError, "").Err()
	}
	count, err := s.databaseClient.FindBuildsCount(in.RepoID, in.Branch)
	if err != nil {
		s.log.Error("Find builds count error", err)
		return &commonTypes.BuildsArray{}, status.New(http.StatusInternalServerError, "").Err()
	}
	protoBuilds := &commonTypes.BuildsArray{}
	for _, build := range builds {
		protoBuild := buildToProto(build)
		protoBuilds.Builds = append(protoBuilds.Builds, protoBuild)
	}
	protoBuilds.Total = count
	return protoBuilds, nil
}
