package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// NewRouter constructs the Gin engine, registers routes and middleware.
func NewRouter(db *gorm.DB) *gin.Engine {
	router := gin.New()

	// Basic middleware: logging + panic recovery.
	router.Use(gin.Logger(), gin.Recovery())

	// Health check endpoint for uptime/load balancers.
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	// Minimal OpenAPI description served as JSON for tooling / documentation.
	router.GET("/openapi.json", func(c *gin.Context) {
		c.JSON(http.StatusOK, openAPISpec())
	})

	checkinHandler := NewCheckinHandler(db)
	authHandler := NewAuthHandler(db)

	apiV1 := router.Group("/api/v1")
	{
		checkins := apiV1.Group("/checkins")
		{
			checkins.GET("", checkinHandler.List)
			checkins.POST("", checkinHandler.Create)
			checkins.PATCH("/:id", checkinHandler.Update)
			checkins.DELETE("/:id", checkinHandler.Delete)
			checkins.GET("/forToday", checkinHandler.GetByUserAndDate) //?userId=1&date=2026-02-22(YYYY-MM-DD)
		}
	}

	auth := router.Group("/auth") 
	{
		google := auth.Group("/google")
		{
			google.GET("/login", authHandler.Login)
			google.GET("/callback", authHandler.GoogleCallback)
		}
	}

	return router
}

// openAPISpec returns a very small OpenAPI 3 document that describes the
// public HTTP API. You can import this into Swagger UI, Postman, or other
// tools to explore the API.
func openAPISpec() gin.H {
	return gin.H{
		"openapi": "3.0.0",
		"info": gin.H{
			"title":       "Wellness Tracker API",
			"description": "Minimal Gin + Gorm API for a wellness tracking frontend.",
			"version":     "1.0.0",
		},
		"paths": gin.H{
			"/health": gin.H{
				"get": gin.H{
					"summary":     "Health check",
					"description": "Returns OK if the service is up.",
					"responses": gin.H{
						"200": gin.H{
							"description": "Service is healthy",
						},
					},
				},
			},
			"/api/v1/checkins": gin.H{
				"get": gin.H{
					"summary":     "List check-ins",
					"description": "Returns a paginated list of wellness check-ins.",
					"parameters": []gin.H{
						{
							"name":        "page",
							"in":          "query",
							"description": "Page number (1-based).",
							"schema": gin.H{
								"type":    "integer",
								"default": 1,
							},
						},
						{
							"name":        "page_size",
							"in":          "query",
							"description": "Page size.",
							"schema": gin.H{
								"type":    "integer",
								"default": 20,
							},
						},
					},
					"responses": gin.H{
						"200": gin.H{
							"description": "List of check-ins.",
						},
					},
				},
				"post": gin.H{
					"summary":     "Create a check-in",
					"description": "Creates a new wellness check-in.",
					"requestBody": gin.H{
						"required": true,
						"content": gin.H{
							"application/json": gin.H{
								"schema": gin.H{
									"type": "object",
								},
							},
						},
					},
					"responses": gin.H{
						"201": gin.H{
							"description": "Created.",
						},
						"400": gin.H{
							"description": "Invalid request body.",
						},
					},
				},
			},
			"/api/v1/checkins/{id}": gin.H{
				"get": gin.H{
					"summary":     "Get a check-in",
					"description": "Returns a single wellness check-in.",
					"parameters": []gin.H{
						{
							"name":     "id",
							"in":       "path",
							"required": true,
							"schema": gin.H{
								"type": "integer",
							},
						},
					},
					"responses": gin.H{
						"200": gin.H{
							"description": "Check-in details.",
						},
						"404": gin.H{
							"description": "Check-in not found.",
						},
					},
				},
				"patch": gin.H{
					"summary":     "Update a check-in",
					"description": "Performs a partial update on a wellness check-in.",
					"parameters": []gin.H{
						{
							"name":     "id",
							"in":       "path",
							"required": true,
							"schema": gin.H{
								"type": "integer",
							},
						},
					},
					"responses": gin.H{
						"200": gin.H{
							"description": "Updated check-in.",
						},
						"404": gin.H{
							"description": "Check-in not found.",
						},
					},
				},
				"delete": gin.H{
					"summary":     "Delete a check-in",
					"description": "Deletes a wellness check-in.",
					"parameters": []gin.H{
						{
							"name":     "id",
							"in":       "path",
							"required": true,
							"schema": gin.H{
								"type": "integer",
							},
						},
					},
					"responses": gin.H{
						"204": gin.H{
							"description": "Deleted.",
						},
						"404": gin.H{
							"description": "Check-in not found.",
						},
					},
				},
			},
		},
	}
}
