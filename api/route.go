package api

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

var (
	errRouteNotFound = errors.New("route not found")
)

func (s Server) registeRoute() {
	s.engine.NoRoute(func(c *gin.Context) {
		c.AbortWithError(http.StatusNotFound, errRouteNotFound)
	})

	s.engine.GET("/v1/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	v1 := s.engine.Group("/v1/assets")
	{
		v1.GET("/", s.searchAssets)
		v1.POST("/", s.createAsset)
		v1.GET("/:id", s.getAssetById)
		v1.PUT("/:id", s.updateAsset)
		v1.DELETE("/:id", s.deleteAsset)
		v1.GET("/:ext/upload", s.getAssetUploadPath)

	}
	category := s.engine.Group("/v1/categories")
	{
		category.GET("/", s.searchCategories)
		category.GET("/:id", s.getCategoryByID)
		category.POST("/", s.createCategory)
		category.PUT("/:id", s.updateCategory)
		category.DELETE("/:id", s.deleteCategory)
	}
}
