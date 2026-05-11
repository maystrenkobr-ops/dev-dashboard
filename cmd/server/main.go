package main

import (
	"context"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/maystrenkobr-ops/dev-dashboard/internal/auth"
	"github.com/maystrenkobr-ops/dev-dashboard/internal/tasks"
	"github.com/maystrenkobr-ops/dev-dashboard/internal/workspaces"
)

func main() {
	ctx := context.Background()
	databaseURL := os.Getenv("DATABASE_URL")

	taskStore, err := tasks.NewStorage(ctx, databaseURL, "data/tasks.json")
	if err != nil {
		panic(err)
	}
	defer taskStore.Close()

	authStore, err := auth.NewStorage(ctx, databaseURL, os.Getenv("ADMIN_USERNAME"), os.Getenv("ADMIN_PASSWORD"))
	if err != nil {
		panic(err)
	}
	defer authStore.Close()

	workspaceStore, err := workspaces.NewStorage(ctx, databaseURL)
	if err != nil {
		panic(err)
	}
	defer workspaceStore.Close()

	sessionManager := auth.NewSessionManager(os.Getenv("SESSION_SECRET"))

	router := gin.Default()

	router.StaticFile("/static/styles.css", "web/styles.css")
	router.StaticFile("/static/app.js", "web/app.js")

	auth.RegisterRoutes(router, authStore, sessionManager)

	router.GET("/", sessionManager.RequireAuth(authStore), func(c *gin.Context) {
		c.File("web/index.html")
	})

	router.HEAD("/", func(c *gin.Context) {
		c.Status(200)
	})

	taskRoutes := router.Group("/tasks")
	taskRoutes.Use(sessionManager.RequireAuth(authStore))
	tasks.RegisterRoutes(taskRoutes, taskStore)

	apiRoutes := router.Group("/api")
	apiRoutes.Use(sessionManager.RequireAuth(authStore))
	workspaces.RegisterRoutes(apiRoutes, workspaceStore, authStore)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	router.Run(":" + port)
}
