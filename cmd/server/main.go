package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"qalbum-server/pkg/config"
	"qalbum-server/pkg/cron"
	"qalbum-server/pkg/dao"
	"qalbum-server/pkg/handler"
	"qalbum-server/pkg/middleware"
	"qalbum-server/pkg/service"
	"qalbum-server/pkg/storage"
	"qalbum-server/pkg/task"
	"qalbum-server/pkg/utils"
)

func main() {
	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	db, err := InitDB("data/db/qalbum.db")
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	jwtManager := utils.NewJWTManager(cfg.JWT.Secret, cfg.JWT.ExpireHour)
	middleware.InitJWTManager(jwtManager)

	userDAO := dao.NewUserDAO(db.DB)
	spaceDAO := dao.NewSpaceDAO(db.DB)
	albumDAO := dao.NewAlbumDAO(db.DB)
	fileDAO := dao.NewFileDAO(db.DB)
	inviteDAO := dao.NewInviteDAO(db.DB)

	localStorage := storage.NewLocalStorage(cfg.Storage.BasePath)

	taskQueue := task.NewQueue(cfg.Task.QueueSize)
	worker := task.NewWorker(context.Background(), cfg.Task.WorkerCount, taskQueue)
	worker.Start()

	authService := service.NewAuthService(userDAO, jwtManager, &service.WeChatConfig{
		AppID:  cfg.WeChat.AppID,
		Secret: cfg.WeChat.Secret,
	})

	spaceService := service.NewSpaceService(spaceDAO, userDAO, inviteDAO, fileDAO, taskQueue, localStorage)
	middleware.InitSpaceService(spaceService)
	albumService := service.NewAlbumService(albumDAO, spaceDAO)
	fileService := service.NewFileService(fileDAO, albumDAO, spaceDAO, localStorage, taskQueue)
	duplicateService := service.NewDuplicateService(fileDAO, spaceDAO)
	recycleService := service.NewRecycleService(fileDAO, albumDAO, spaceDAO, localStorage)

	authHandler := handler.NewAuthHandler(authService)
	spaceHandler := handler.NewSpaceHandler(spaceService)
	albumHandler := handler.NewAlbumHandler(albumService)
	fileHandler := handler.NewFileHandler(fileService)
	duplicateHandler := handler.NewDuplicateHandler(duplicateService)
	recycleHandler := handler.NewRecycleHandler(recycleService)

	cleanupCron := cron.NewCleanupCron(fileDAO, albumDAO, spaceDAO, localStorage, cfg.Cleanup.Hour)
	if cfg.Cleanup.Enabled {
		cleanupCron.Start()
	}

	auditCron := cron.NewAuditCron(fileDAO, taskQueue, localStorage, cfg.Audit.IntervalHours, cfg.Backup.UploadGraceHours, cfg.Audit.Enabled)
	auditCron.Start()

	r := gin.New()
	r.Use(middleware.Recovery())
	r.Use(middleware.Logger())
	r.Use(middleware.CORS())

	handler.RegisterAuthRoutes(r, authHandler)
	handler.RegisterSpaceRoutes(r, spaceHandler)
	handler.RegisterAlbumRoutes(r, albumHandler)
	handler.RegisterFileRoutes(r, fileHandler)
	handler.RegisterDuplicateRoutes(r, duplicateHandler)
	handler.RegisterRecycleRoutes(r, recycleHandler)

	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
	}

	go func() {
		log.Printf("Starting server on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	srv.Shutdown(ctx)
	worker.Stop()
	cleanupCron.Stop()
	auditCron.Stop()
	
	log.Println("Server stopped")
}
