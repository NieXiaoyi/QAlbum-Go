package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"qalbum-server/pkg/middleware"
	"qalbum-server/pkg/service"
)

type RecycleHandler struct {
	recycleService *service.RecycleService
}

func NewRecycleHandler(recycleService *service.RecycleService) *RecycleHandler {
	return &RecycleHandler{recycleService: recycleService}
}

func (h *RecycleHandler) List(c *gin.Context) {
	spaceID, _ := strconv.Atoi(c.Param("space_id"))
	userID := middleware.GetUserID(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	files, total, err := h.recycleService.List(spaceID, userID, page, pageSize)
	if err != nil {
		if err.Error() == "admin only" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"total":     total,
		"page":      page,
		"page_size": pageSize,
		"items":     files,
	})
}

func (h *RecycleHandler) Restore(c *gin.Context) {
	spaceID, _ := strconv.Atoi(c.Param("space_id"))
	fileID, _ := strconv.Atoi(c.Param("file_id"))
	userID := middleware.GetUserID(c)

	file, err := h.recycleService.Restore(spaceID, fileID, userID)
	if err != nil {
		if err.Error() == "admin only" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, file)
}

func (h *RecycleHandler) Clear(c *gin.Context) {
	spaceID, _ := strconv.Atoi(c.Param("space_id"))
	userID := middleware.GetUserID(c)

	if err := h.recycleService.Clear(spaceID, userID); err != nil {
		if err.Error() == "admin only" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.Status(http.StatusNoContent)
}

func RegisterRecycleRoutes(r *gin.Engine, recycleHandler *RecycleHandler) {
	spaces := r.Group("/spaces/:space_id", middleware.Auth())
	{
		spaces.GET("/recycle", recycleHandler.List)
		spaces.POST("/recycle/:file_id/restore", recycleHandler.Restore)
		spaces.POST("/recycle/clear", recycleHandler.Clear)
	}
}
