package main

import (
	"context"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/maystrenkobr-ops/dev-dashboard/internal/auth"
	"github.com/maystrenkobr-ops/dev-dashboard/internal/tasks"
	"github.com/maystrenkobr-ops/dev-dashboard/internal/workspaces"
)

func main() {
	ctx := context.Background()
	databaseURL := os.Getenv("DATABASE_URL")

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

	adminUserID := 0
	defaultWorkspaceID := 0
	adminUsername := strings.ToLower(strings.TrimSpace(os.Getenv("ADMIN_USERNAME")))

	if adminUsername != "" {
		adminUser, _, err := authStore.GetUserByUsername(ctx, adminUsername)
		if err == nil {
			adminUserID = adminUser.ID

			defaultWorkspace, err := workspaceStore.EnsureDefaultWorkspace(ctx, adminUserID)
			if err != nil {
				panic(err)
			}

			defaultWorkspaceID = defaultWorkspace.ID
		}
	}

	taskStore, err := tasks.NewStorage(ctx, databaseURL, "data/tasks.json")
	if err != nil {
		panic(err)
	}
	defer taskStore.Close()

	if defaultWorkspaceID > 0 {
		if err := taskStore.EnsureWorkspaceSupport(ctx, defaultWorkspaceID, adminUserID); err != nil {
			panic(err)
		}
	}

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
	tasks.RegisterRoutes(taskRoutes, taskStore, workspaceStore)

	apiRoutes := router.Group("/api")
	apiRoutes.Use(sessionManager.RequireAuth(authStore))
	workspaces.RegisterRoutes(apiRoutes, workspaceStore, authStore)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	router.Run(":" + port)
}
