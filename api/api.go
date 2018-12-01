package api

import (
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/revan730/clipper-ci-worker/db"
	"go.uber.org/zap"
)

type Config struct {
	Port int
}

type Server struct {
	logger         *zap.Logger
	config         Config
	databaseClient db.DatabaseClient
	router         *gin.Engine
}

func NewServer(config Config, logger *zap.Logger, dbClient db.DatabaseClient) *Server {
	server := &Server{
		config:         config,
		logger:         logger,
		databaseClient: dbClient,
		router:         gin.Default(),
	}
	server.Routes()
	return server
}

func (s *Server) logError(msg string, err error) {
	defer s.logger.Sync()
	s.logger.Error(msg, zap.String("packageLevel", "api"), zap.Error(err))
}

func (s *Server) logInfo(msg string) {
	defer s.logger.Sync()
	s.logger.Info("INFO", zap.String("msg", msg), zap.String("packageLevel", "api"))
}

// Routes binds api routes to handlers
func (s *Server) Routes() *Server {
	s.router.GET("/api/v1/build/:id", s.getBuildHandler)
	return s
}

// Run starts api server
func (s *Server) Run() {
	defer s.databaseClient.Close()
	rand.Seed(time.Now().UnixNano())
	err := s.databaseClient.CreateSchema()
	if err != nil {
		s.logError("Failed to create database schema", err)
		os.Exit(1)
	}
	s.logger.Info("Starting api server", zap.Int("port", s.config.Port))
	err = http.ListenAndServe(fmt.Sprintf(":%d", s.config.Port), s.router)
	if err != nil {
		s.logError("API server failed", err)
		os.Exit(1)
	}
}

func (s *Server) bindJSON(c *gin.Context, msg interface{}) bool {
	err := c.ShouldBindJSON(&msg)
	if err != nil {
		s.logError("JSON read error", err)
		c.JSON(http.StatusBadRequest, gin.H{"err": "Bad json"})
		return false
	}
	return true
}
