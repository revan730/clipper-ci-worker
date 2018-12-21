package api

import (
	"context"
	"net/http"

	commonTypes "github.com/revan730/clipper-common/types"
	"google.golang.org/grpc/status"
)

func (s *Server) GetBuildArtifact(ctx context.Context, in *commonTypes.BuildArtifact) (*commonTypes.BuildArtifact, error) {
	artifact, err := s.databaseClient.FindBuildArtifact(in.BuildID)
	if err != nil {
		s.logError("Find build artifact error", err)
		return &commonTypes.BuildArtifact{}, status.New(http.StatusInternalServerError, "").Err()
	}
	if artifact == nil {
		return &commonTypes.BuildArtifact{}, status.New(http.StatusNotFound, "build artifact not found").Err()
	}
	protoArtifact := &commonTypes.BuildArtifact{
		ID:      artifact.ID,
		BuildID: artifact.BuildID,
		Name:    artifact.Name,
	}
	return protoArtifact, nil
}

func (s *Server) GetBuildArtifactByID(ctx context.Context, in *commonTypes.BuildArtifact) (*commonTypes.BuildArtifact, error) {
	artifact, err := s.databaseClient.FindBuildArtifactByID(in.ID)
	if err != nil {
		s.logError("Find build artifact error", err)
		return &commonTypes.BuildArtifact{}, status.New(http.StatusInternalServerError, "").Err()
	}
	if artifact == nil {
		return &commonTypes.BuildArtifact{}, status.New(http.StatusNotFound, "build artifact not found").Err()
	}
	protoArtifact := &commonTypes.BuildArtifact{
		ID:      artifact.ID,
		BuildID: artifact.BuildID,
		Name:    artifact.Name,
	}
	return protoArtifact, nil
}
