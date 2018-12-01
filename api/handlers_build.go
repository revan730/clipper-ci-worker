package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	ptypes "github.com/gogo/protobuf/types"
	commonTypes "github.com/revan730/clipper-common/types"
)

func (s *Server) getBuildHandler(c *gin.Context) {
	buildIDStr := c.Param("id")
	buildID, err := strconv.Atoi(buildIDStr)
	if err != nil {
		protoErr := commonTypes.Error{
			Err: "repo id is not int",
		}
		c.JSON(http.StatusBadRequest, protoErr)
		return
	}
	build, err := s.databaseClient.FindBuildByID(int64(buildID))
	if err != nil {
		s.logError("Find build error", err)
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	if build == nil {
		protoErr := commonTypes.Error{
			Err: "repo id is not int",
		}
		c.JSON(http.StatusNotFound, protoErr)
		return
	}
	timestamp, _ := ptypes.TimestampProto(build.Date)
	protoBuild := commonTypes.Build{
		ID:            build.ID,
		GithubRepoID:  build.GithubRepoID,
		IsSuccessfull: build.IsSuccessfull,
		Date:          timestamp,
		Branch:        build.Branch,
		Stdout:        build.Stdout,
	}
	c.JSON(http.StatusOK, protoBuild)
}
