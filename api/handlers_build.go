package api

import (
	"context"
	"net/http"

	ptypes "github.com/golang/protobuf/ptypes"
	commonTypes "github.com/revan730/clipper-common/types"
	"google.golang.org/grpc/status"
)

func (s *Server) GetBuild(ctx context.Context, in *commonTypes.Build) (*commonTypes.Build, error) {
	build, err := s.databaseClient.FindBuildByID(in.ID)
	if err != nil {
		s.logError("Find build error", err)
		return &commonTypes.Build{}, status.New(http.StatusInternalServerError, "").Err()
	}
	if build == nil {
		return &commonTypes.Build{}, status.New(http.StatusNotFound, "repo not found").Err()
	}
	timestamp, _ := ptypes.TimestampProto(build.Date)
	protoBuild := &commonTypes.Build{
		ID:            build.ID,
		GithubRepoID:  build.GithubRepoID,
		IsSuccessfull: build.IsSuccessfull,
		Date:          timestamp,
		Branch:        build.Branch,
		Stdout:        build.Stdout,
	}
	return protoBuild, nil
}
