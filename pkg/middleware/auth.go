package middleware

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"qalbum-server/pkg/utils"
)

var jwtManager *utils.JWTManager
var spaceService interface {
	IsMember(spaceID, userID int) (bool, error)
	GetMemberRole(spaceID, userID int) (string, error)
}

func InitJWTManager(manager *utils.JWTManager) {
	jwtManager = manager
}

func InitSpaceService(svc interface {
	IsMember(spaceID, userID int) (bool, error)
	GetMemberRole(spaceID, userID int) (string, error)
}) {
	spaceService = svc
}

func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "missing authorization header",
			})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "invalid authorization header format",
			})
			c.Abort()
			return
		}

		userID, err := jwtManager.ParseToken(parts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "invalid token",
			})
			c.Abort()
			return
		}

		c.Set("user_id", userID)
		c.Next()
	}
}

func SpaceMemberAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		spaceIDStr := c.Param("space_id")
		spaceID, err := strconv.Atoi(spaceIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid space_id"})
			c.Abort()
			return
		}

		userID := GetUserID(c)
		isMember, err := spaceService.IsMember(spaceID, userID)
		if err != nil || !isMember {
			c.JSON(http.StatusForbidden, gin.H{"error": "not a member of this space"})
			c.Abort()
			return
		}

		c.Next()
	}
}

func SpaceAdminAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		spaceIDStr := c.Param("space_id")
		spaceID, err := strconv.Atoi(spaceIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid space_id"})
			c.Abort()
			return
		}

		userID := GetUserID(c)
		role, err := spaceService.GetMemberRole(spaceID, userID)
		if err != nil || role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "only admin can perform this action"})
			c.Abort()
			return
		}

		c.Next()
	}
}

func GetUserID(c *gin.Context) int {
	userID, _ := c.Get("user_id")
	return userID.(int)
}
