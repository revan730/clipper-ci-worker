package api

import (
	"context"
	"net/http"

	"github.com/revan730/clipper-ci-worker/types"
	commonTypes "github.com/revan730/clipper-common/types"
	"google.golang.org/grpc/status"
)

func artifactToProto(a *types.BuildArtifact) *commonTypes.BuildArtifact {
	return &commonTypes.BuildArtifact{
		ID:      a.ID,
		BuildID: a.BuildID,
		Name:    a.Name,
	}
}

func (s *Server) GetBuildArtifact(ctx context.Context, in *commonTypes.BuildArtifact) (*commonTypes.BuildArtifact, error) {
	artifact, err := s.databaseClient.FindBuildArtifact(in.BuildID)
	if err != nil {
		s.log.Error("Find build artifact error", err)
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
		s.log.Error("Find build artifact error", err)
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

func (s *Server) GetAllArtifacts(ctx context.Context, in *commonTypes.BuildsQuery) (*commonTypes.ArtifactsArray, error) {
	artifacts, err := s.databaseClient.FindAllBuildArtifacts(in.RepoID, in.Branch, in.Page, in.Limit)
	if err != nil {
		s.log.Error("Find all artifacts error", err)
		return &commonTypes.ArtifactsArray{}, status.New(http.StatusInternalServerError, "").Err()
	}
	count, err := s.databaseClient.FindBuildArtifactsCount(in.RepoID, in.Branch)
	if err != nil {
		s.log.Error("Find artifacts count error", err)
		return &commonTypes.ArtifactsArray{}, status.New(http.StatusInternalServerError, "").Err()
	}
	protoArtifacts := &commonTypes.ArtifactsArray{}
	for _, artifact := range artifacts {
		protoArtifact := artifactToProto(artifact)
		protoArtifacts.Artifacts = append(protoArtifacts.Artifacts, protoArtifact)
	}
	protoArtifacts.Total = count
	return protoArtifacts, nil
}
