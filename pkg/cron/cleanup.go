package cron

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"qalbum-server/pkg/dao"
	"qalbum-server/pkg/storage"
)

type CleanupCron struct {
	fileDAO  *dao.FileDAO
	albumDAO *dao.AlbumDAO
	spaceDAO  *dao.SpaceDAO
	storage  storage.Interface
	interval  time.Duration
	done      chan bool
}

func NewCleanupCron(fileDAO *dao.FileDAO, albumDAO *dao.AlbumDAO, spaceDAO *dao.SpaceDAO, storage storage.Interface, hour int) *CleanupCron {
	return &CleanupCron{
		fileDAO:  fileDAO,
		albumDAO: albumDAO,
		spaceDAO:  spaceDAO,
		storage:  storage,
		interval:  24 * time.Hour,
		done:      make(chan bool),
	}
}

func (c *CleanupCron) Start() {
	go func() {
		ticker := time.NewTicker(c.interval)
		defer ticker.Stop()
		
		for {
			select {
			case <-ticker.C:
				c.RunCleanup()
			case <-c.done:
				return
			}
		}
	}()
	
	log.Printf("Cleanup scheduled to run every %v", c.interval)
}

func (c *CleanupCron) Stop() {
	close(c.done)
	log.Printf("Cleanup stopped")
}

func (c *CleanupCron) RunCleanup() {
	log.Printf("Running cleanup task")
	
	expiredFiles, err := c.fileDAO.GetExpiredRecycledFiles(30)
	if err != nil {
		log.Printf("Failed to get expired files: %v", err)
		return
	}
	
	for _, file := range expiredFiles {
		if err := c.storage.Delete(file.StoragePath); err != nil {
			log.Printf("Failed to delete file %d: %v", file.ID, err)
		}

		if file.ThumbnailPath != nil {
			c.storage.Delete(*file.ThumbnailPath)
		}

		if file.CoverPath != nil {
			c.storage.Delete(*file.CoverPath)
		}

		space, _ := c.spaceDAO.GetByID(file.SpaceID)
		if space != nil && space.BackupPath != "" {
			backupPath := filepath.Join(space.BackupPath, strings.TrimPrefix(file.StoragePath, "photos"))
			os.Remove(backupPath)
		}

		if err := c.fileDAO.HardDelete(file.ID); err != nil {
			log.Printf("Failed to hard delete file %d: %v", file.ID, err)
		}
	}
	
	log.Printf("Cleaned up %d expired recycled files", len(expiredFiles))
	
	expiredAlbums, err := c.albumDAO.GetExpiredDeletedAlbums(30)
	if err != nil {
		log.Printf("Failed to get expired albums: %v", err)
		return
	}
	
	for _, album := range expiredAlbums {
		if err := c.albumDAO.HardDelete(album.ID); err != nil {
			log.Printf("Failed to hard delete album %d: %v", album.ID, err)
		}
	}
	
	log.Printf("Cleaned up %d expired deleted albums", len(expiredAlbums))
}
