package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"qalbum-server/pkg/middleware"
	"qalbum-server/pkg/service"
)

type DuplicateHandler struct {
	duplicateService *service.DuplicateService
}

func NewDuplicateHandler(duplicateService *service.DuplicateService) *DuplicateHandler {
	return &DuplicateHandler{duplicateService: duplicateService}
}

func (h *DuplicateHandler) Detect(c *gin.Context) {
	spaceID, _ := strconv.Atoi(c.Param("space_id"))
	userID := middleware.GetUserID(c)
	threshold, _ := strconv.Atoi(c.DefaultQuery("threshold", "10"))

	groups, err := h.duplicateService.Detect(spaceID, userID, threshold)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"groups": groups,
	})
}

func RegisterDuplicateRoutes(r *gin.Engine, duplicateHandler *DuplicateHandler) {
	spaces := r.Group("/spaces/:space_id", middleware.Auth())
	{
		spaces.GET("/duplicates", duplicateHandler.Detect)
	}
}
