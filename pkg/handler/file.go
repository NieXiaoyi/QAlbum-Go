package handler

import (
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"qalbum-server/pkg/middleware"
	"qalbum-server/pkg/service"
)

type FileHandler struct {
	fileService *service.FileService
}

func NewFileHandler(fileService *service.FileService) *FileHandler {
	return &FileHandler{fileService: fileService}
}

type UpdateFileRequest struct {
	Filename string `json:"filename"`
	AlbumID  int    `json:"album_id"`
}

const maxUploadSize = 200 * 1024 * 1024

func (h *FileHandler) List(c *gin.Context) {
	spaceID, _ := strconv.Atoi(c.Param("space_id"))
	albumID, _ := strconv.Atoi(c.Param("album_id"))
	userID := middleware.GetUserID(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	files, total, err := h.fileService.ListByAlbum(spaceID, albumID, userID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"total":     total,
		"page":      page,
		"page_size": pageSize,
		"items":     files,
	})
}

func (h *FileHandler) Upload(c *gin.Context) {
	spaceID, _ := strconv.Atoi(c.Param("space_id"))
	albumID, _ := strconv.Atoi(c.Param("album_id"))
	userID := middleware.GetUserID(c)

	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxUploadSize)
	if c.Request.ContentLength > maxUploadSize {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "file too large"})
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no file uploaded"})
		return
	}

	reader, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to open file"})
		return
	}
	defer reader.Close()

	mimeType := file.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	customFilename := c.PostForm("filename")
	if customFilename == "" {
		customFilename = file.Filename
	}

	meta := &service.UploadFileMeta{
		FileName: customFilename,
		MimeType: mimeType,
		Size:     file.Size,
		Reader:   reader,
	}

	uploadedFile, err := h.fileService.Upload(spaceID, albumID, userID, meta)
	if err != nil {
		if err.Error() == "quota check failed: quota exceeded" {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusCreated, uploadedFile)
}

func (h *FileHandler) Get(c *gin.Context) {
	spaceID, _ := strconv.Atoi(c.Param("space_id"))
	fileID, _ := strconv.Atoi(c.Param("file_id"))
	userID := middleware.GetUserID(c)

	file, err := h.fileService.GetByID(fileID, userID)
	if err != nil {
		if err.Error() == "not a member" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		}
		return
	}

	if file.SpaceID != spaceID {
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}

	c.JSON(http.StatusOK, file)
}

func (h *FileHandler) Update(c *gin.Context) {
	spaceID, _ := strconv.Atoi(c.Param("space_id"))
	fileID, _ := strconv.Atoi(c.Param("file_id"))
	userID := middleware.GetUserID(c)
	var req UpdateFileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	file, err := h.fileService.GetByID(fileID, userID)
	if err != nil {
		if err.Error() == "not a member" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		}
		return
	}

	if file.SpaceID != spaceID {
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}

	updatedFile, err := h.fileService.Update(fileID, userID, req.Filename, req.AlbumID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updatedFile)
}

func (h *FileHandler) Delete(c *gin.Context) {
	spaceID, _ := strconv.Atoi(c.Param("space_id"))
	fileID, _ := strconv.Atoi(c.Param("file_id"))
	userID := middleware.GetUserID(c)

	file, err := h.fileService.GetByID(fileID, userID)
	if err != nil {
		if err.Error() == "not a member" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		}
		return
	}

	if file.SpaceID != spaceID {
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}

	if err := h.fileService.SoftDelete(fileID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *FileHandler) Download(c *gin.Context) {
	spaceID, _ := strconv.Atoi(c.Param("space_id"))
	fileID, _ := strconv.Atoi(c.Param("file_id"))
	userID := middleware.GetUserID(c)

	file, err := h.fileService.GetByID(fileID, userID)
	if err != nil {
		if err.Error() == "not a member" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		}
		return
	}

	if file.SpaceID != spaceID {
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}

	reader, mimeType, err := h.fileService.Download(fileID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer reader.Close()

	c.Header("Content-Type", mimeType)
	c.Header("Content-Disposition", "attachment; filename="+file.Filename)
	c.Stream(func(w io.Writer) bool {
		_, err := io.Copy(w, reader)
		return err == nil
	})
}

func (h *FileHandler) GetThumbnail(c *gin.Context) {
	spaceID, _ := strconv.Atoi(c.Param("space_id"))
	fileID, _ := strconv.Atoi(c.Param("file_id"))
	userID := middleware.GetUserID(c)

	file, err := h.fileService.GetByID(fileID, userID)
	if err != nil {
		if err.Error() == "not a member" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		}
		return
	}

	if file.SpaceID != spaceID {
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}

	reader, err := h.fileService.GetThumbnail(fileID, userID)
	if err != nil {
		if err.Error() == "thumbnail not ready" {
			c.Status(http.StatusAccepted)
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	defer reader.Close()

	c.Header("Content-Type", "image/jpeg")
	c.Stream(func(w io.Writer) bool {
		_, err := io.Copy(w, reader)
		return err == nil
	})
}

func RegisterFileRoutes(r *gin.Engine, fileHandler *FileHandler) {
	spaces := r.Group("/spaces/:space_id", middleware.Auth())
	{
		spaces.GET("/albums/:album_id/files", fileHandler.List)
		spaces.POST("/albums/:album_id/files", fileHandler.Upload)
		spaces.GET("/files/:file_id", fileHandler.Get)
		spaces.PUT("/files/:file_id", fileHandler.Update)
		spaces.DELETE("/files/:file_id", fileHandler.Delete)
		spaces.GET("/files/:file_id/download", fileHandler.Download)
		spaces.GET("/files/:file_id/thumbnail", fileHandler.GetThumbnail)
	}
}
