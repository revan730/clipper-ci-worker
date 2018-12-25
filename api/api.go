package api

import (
	"fmt"
	"math/rand"
	"net"
	"time"

	"google.golang.org/grpc"
	"github.com/revan730/clipper-ci-worker/db"
	"github.com/revan730/clipper-common/types"
	"github.com/revan730/clipper-ci-worker/log"
)

type Config struct {
	Port int
}

type Server struct {
	log         log.Logger
	config         Config
	databaseClient db.DatabaseClient
}

func NewServer(config Config, logger log.Logger, dbClient db.DatabaseClient) *Server {
	server := &Server{
		config:         config,
		log:         logger,
		databaseClient: dbClient,
	}
	return server
}

// Run starts api server
func (s *Server) Run() {
	defer s.databaseClient.Close()
	rand.Seed(time.Now().UnixNano())
	err := s.databaseClient.CreateSchema()
	if err != nil {
		s.log.Fatal("Failed to create database schema", err)
	}
	s.log.Info(fmt.Sprintf("Starting api server at port %d", s.config.Port))
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.config.Port))
	if err != nil {
		s.log.Fatal("API server failed", err)
	}
	grpcServer := grpc.NewServer()
	types.RegisterCIAPIServer(grpcServer, s)
	if err := grpcServer.Serve(lis); err != nil {
		s.log.Fatal("failed to serve: %s", err)
	}
}