package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"qalbum-server/pkg/middleware"
	"qalbum-server/pkg/service"
)

type SpaceHandler struct {
	spaceService *service.SpaceService
}

func NewSpaceHandler(spaceService *service.SpaceService) *SpaceHandler {
	return &SpaceHandler{spaceService: spaceService}
}

type CreateSpaceRequest struct {
	Name       string `json:"name" binding:"required"`
	QuotaBytes int64  `json:"quota_bytes" binding:"required"`
	BackupPath string `json:"backup_path"`
}

type UpdateSpaceRequest struct {
	Name       string `json:"name"`
	QuotaBytes int64  `json:"quota_bytes"`
	BackupPath string `json:"backup_path"`
}

type CreateInviteRequest struct {
	ExpireHours int `json:"expire_hours"`
}

type JoinSpaceRequest struct {
	InviteCode string `json:"invite_code" binding:"required"`
}

func (h *SpaceHandler) List(c *gin.Context) {
	userID := middleware.GetUserID(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	spaces, err := h.spaceService.ListByUser(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"total":     len(spaces),
		"page":      page,
		"page_size": pageSize,
		"items":     spaces,
	})
}

func (h *SpaceHandler) Create(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req CreateSpaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	space, err := h.spaceService.Create(req.Name, userID, req.QuotaBytes, req.BackupPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, space)
}

func (h *SpaceHandler) Get(c *gin.Context) {
	spaceID, _ := strconv.Atoi(c.Param("space_id"))
	userID := middleware.GetUserID(c)

	if _, err := h.spaceService.GetMemberRole(spaceID, userID); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "not a member"})
		return
	}

	space, err := h.spaceService.GetByID(spaceID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "space not found"})
		return
	}

	c.JSON(http.StatusOK, space)
}

func (h *SpaceHandler) Update(c *gin.Context) {
	spaceID, _ := strconv.Atoi(c.Param("space_id"))
	userID := middleware.GetUserID(c)
	var req UpdateSpaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	space, err := h.spaceService.Update(spaceID, userID, req.Name, req.QuotaBytes, req.BackupPath)
	if err != nil {
		if err.Error() == "only admin can update space" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, space)
}

func (h *SpaceHandler) Delete(c *gin.Context) {
	spaceID, _ := strconv.Atoi(c.Param("space_id"))
	userID := middleware.GetUserID(c)

	if err := h.spaceService.Delete(spaceID, userID); err != nil {
		if err.Error() == "only admin can delete space" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *SpaceHandler) GenerateInvite(c *gin.Context) {
	spaceID, _ := strconv.Atoi(c.Param("space_id"))
	userID := middleware.GetUserID(c)
	var req CreateInviteRequest
	c.ShouldBindJSON(&req)

	if req.ExpireHours == 0 {
		req.ExpireHours = 24
	}

	invite, err := h.spaceService.GenerateInviteToken(spaceID, userID, req.ExpireHours)
	if err != nil {
		if err.Error() == "only admin can generate invite tokens" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"invite_code": invite.Token,
		"expires_at":  invite.ExpiresAt,
	})
}

func (h *SpaceHandler) Join(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req JoinSpaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	space, err := h.spaceService.JoinByToken(userID, req.InviteCode)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, space)
}

func (h *SpaceHandler) GetMembers(c *gin.Context) {
	spaceID, _ := strconv.Atoi(c.Param("space_id"))
	userID := middleware.GetUserID(c)

	members, err := h.spaceService.GetMembers(spaceID, userID)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, members)
}

func (h *SpaceHandler) RemoveMember(c *gin.Context) {
	spaceID, _ := strconv.Atoi(c.Param("space_id"))
	userID := middleware.GetUserID(c)
	targetUserID, _ := strconv.Atoi(c.Param("user_id"))

	if err := h.spaceService.RemoveMember(spaceID, userID, targetUserID); err != nil {
		if err.Error() == "only admin can remove members" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *SpaceHandler) SyncBackup(c *gin.Context) {
	spaceID, _ := strconv.Atoi(c.Param("space_id"))
	userID := middleware.GetUserID(c)

	pendingCount, err := h.spaceService.SyncBackup(spaceID, userID)
	if err != nil {
		if err.Error() == "only admin can sync backup" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"pending_count": pendingCount})
}

func RegisterSpaceRoutes(r *gin.Engine, spaceHandler *SpaceHandler) {
	spaces := r.Group("/spaces", middleware.Auth())
	{
		spaces.GET("", spaceHandler.List)
		spaces.POST("", spaceHandler.Create)
		spaces.GET("/:space_id", spaceHandler.Get)
		spaces.PUT("/:space_id", spaceHandler.Update)
		spaces.DELETE("/:space_id", spaceHandler.Delete)
		spaces.POST("/:space_id/invite", spaceHandler.GenerateInvite)
		spaces.GET("/:space_id/members", spaceHandler.GetMembers)
		spaces.DELETE("/:space_id/members/:user_id", spaceHandler.RemoveMember)
		spaces.POST("/:space_id/backup/sync", spaceHandler.SyncBackup)
	}
	r.POST("/spaces/join", middleware.Auth(), spaceHandler.Join)
}
