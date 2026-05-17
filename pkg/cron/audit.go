package cron

import (
	"log"
	"time"

	"qalbum-server/pkg/dao"
	"qalbum-server/pkg/storage"
	"qalbum-server/pkg/task"
)

type AuditCron struct {
	fileDAO    *dao.FileDAO
	taskQueue  *task.Queue
	storage    storage.Interface
	interval   time.Duration
	done       chan bool
	enabled    bool
	graceHours int
}

func NewAuditCron(fileDAO *dao.FileDAO, taskQueue *task.Queue, storage storage.Interface, intervalHours int, graceHours int, enabled bool) *AuditCron {
	interval := time.Duration(intervalHours) * time.Hour
	return &AuditCron{
		fileDAO:    fileDAO,
		taskQueue:  taskQueue,
		storage:    storage,
		interval:   interval,
		done:       make(chan bool),
		enabled:    enabled,
		graceHours: graceHours,
	}
}

func (c *AuditCron) Start() {
	if !c.enabled {
		log.Printf("Audit disabled")
		return
	}

	go func() {
		ticker := time.NewTicker(c.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				c.RunAudit()
			case <-c.done:
				return
			}
		}
	}()

	log.Printf("Audit scheduled to run every %v", c.interval)
}

func (c *AuditCron) Stop() {
	close(c.done)
	log.Printf("Audit stopped")
}

func (c *AuditCron) RunAudit() {
	log.Printf("Running audit task")

	spaces, err := c.fileDAO.GetSpacesWithMissingMetadata(c.graceHours)
	if err != nil {
		log.Printf("Failed to get spaces with missing metadata: %v", err)
		return
	}

	for _, spaceID := range spaces {
		files, err := c.fileDAO.GetFilesWithMissingMetadata(spaceID)
		if err != nil {
			log.Printf("Failed to get files with missing metadata for space %d: %v", spaceID, err)
			continue
		}

		for _, file := range files {
			if file.ThumbnailPath == nil && file.FileType == "image" {
				t := task.NewThumbnailTask(c.fileDAO, c.storage, file.ID)
				if err := c.taskQueue.Enqueue(t); err != nil {
					log.Printf("Failed to enqueue thumbnail task: %v", err)
				}
			}

			if file.CoverPath == nil && file.FileType == "video" {
				t := task.NewCoverTask(c.fileDAO, c.storage, file.ID)
				if err := c.taskQueue.Enqueue(t); err != nil {
					log.Printf("Failed to enqueue cover task: %v", err)
				}
			}

			if file.PHash == nil && file.FileType == "image" {
				t := task.NewPHashTask(c.fileDAO, c.storage, file.ID)
				if err := c.taskQueue.Enqueue(t); err != nil {
					log.Printf("Failed to enqueue pHash task: %v", err)
				}
			}
		}

		log.Printf("Audited space %d: enqueued %d tasks", spaceID, len(files)*3)
	}
}
