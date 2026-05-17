package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"qalbum-server/pkg/middleware"
	"qalbum-server/pkg/service"
)

type AlbumHandler struct {
	albumService *service.AlbumService
}

func NewAlbumHandler(albumService *service.AlbumService) *AlbumHandler {
	return &AlbumHandler{albumService: albumService}
}

type CreateAlbumRequest struct {
	Name string `json:"name" binding:"required"`
}

type UpdateAlbumRequest struct {
	Name string `json:"name" binding:"required"`
}

func (h *AlbumHandler) List(c *gin.Context) {
	spaceID, _ := strconv.Atoi(c.Param("space_id"))
	userID := middleware.GetUserID(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	albums, err := h.albumService.ListBySpace(spaceID, userID)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"total":     len(albums),
		"page":      page,
		"page_size": pageSize,
		"items":     albums,
	})
}

func (h *AlbumHandler) Create(c *gin.Context) {
	spaceID, _ := strconv.Atoi(c.Param("space_id"))
	userID := middleware.GetUserID(c)
	var req CreateAlbumRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	album, err := h.albumService.Create(spaceID, req.Name, userID)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, album)
}

func (h *AlbumHandler) Get(c *gin.Context) {
	spaceID, _ := strconv.Atoi(c.Param("space_id"))
	albumID, _ := strconv.Atoi(c.Param("album_id"))
	userID := middleware.GetUserID(c)

	album, err := h.albumService.GetByID(albumID, userID)
	if err != nil {
		if err.Error() == "not a member" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusNotFound, gin.H{"error": "album not found"})
		}
		return
	}

	if album.SpaceID != spaceID {
		c.JSON(http.StatusNotFound, gin.H{"error": "album not found"})
		return
	}

	c.JSON(http.StatusOK, album)
}

func (h *AlbumHandler) Update(c *gin.Context) {
	spaceID, _ := strconv.Atoi(c.Param("space_id"))
	albumID, _ := strconv.Atoi(c.Param("album_id"))
	userID := middleware.GetUserID(c)
	var req UpdateAlbumRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	album, err := h.albumService.Update(albumID, req.Name, userID)
	if err != nil {
		if err.Error() == "not a member" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusNotFound, gin.H{"error": "album not found"})
		}
		return
	}

	if album.SpaceID != spaceID {
		c.JSON(http.StatusNotFound, gin.H{"error": "album not found"})
		return
	}

	c.JSON(http.StatusOK, album)
}

func (h *AlbumHandler) Delete(c *gin.Context) {
	spaceID, _ := strconv.Atoi(c.Param("space_id"))
	albumID, _ := strconv.Atoi(c.Param("album_id"))
	userID := middleware.GetUserID(c)

	album, err := h.albumService.GetByID(albumID, userID)
	if err != nil {
		if err.Error() == "not a member" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusNotFound, gin.H{"error": "album not found"})
		}
		return
	}

	if album.SpaceID != spaceID {
		c.JSON(http.StatusNotFound, gin.H{"error": "album not found"})
		return
	}

	if err := h.albumService.SoftDelete(albumID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

func RegisterAlbumRoutes(r *gin.Engine, albumHandler *AlbumHandler) {
	spaces := r.Group("/spaces/:space_id/albums", middleware.Auth())
	{
		spaces.GET("", albumHandler.List)
		spaces.POST("", albumHandler.Create)
		spaces.GET("/:album_id", albumHandler.Get)
		spaces.PUT("/:album_id", albumHandler.Update)
		spaces.DELETE("/:album_id", albumHandler.Delete)
	}
}
