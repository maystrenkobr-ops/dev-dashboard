package main

import (
	"context"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/maystrenkobr-ops/dev-dashboard/internal/tasks"
)

func main() {
	store, err := tasks.NewStorage(context.Background(), os.Getenv("DATABASE_URL"), "data/tasks.json")
	if err != nil {
		panic(err)
	}
	defer store.Close()

	router := gin.Default()

	router.StaticFile("/static/styles.css", "web/styles.css")
	router.StaticFile("/static/app.js", "web/app.js")

	router.GET("/", func(c *gin.Context) {
		c.File("web/index.html")
	})

	router.HEAD("/", func(c *gin.Context) {
		c.Status(200)
	})

	tasks.RegisterRoutes(router, store)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	router.Run(":" + port)
}
