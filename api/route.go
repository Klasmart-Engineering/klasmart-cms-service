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

	assets := s.engine.Group("/v1/assets")
	{
		assets.GET("/", s.searchAssets)
		assets.POST("/", s.createAsset)
		assets.GET("/:id", s.getAssetByID)
		assets.PUT("/:id", s.updateAsset)
		assets.DELETE("/:id", s.deleteAsset)
	}
	resource := s.engine.Group("/v1/resources")
	{
		resource.GET("/upload/:ext", s.getAssetUploadPath)
		resource.GET("/path/:resource_name", s.getAssetResourcePath)
	}
	category := s.engine.Group("/v1/categories")
	{
		category.GET("/", MustLogin, s.searchCategories)
		category.GET("/:id", MustLogin, s.getCategoryByID)
		category.POST("/", MustLogin, s.createCategory)
		category.PUT("/:id", MustLogin, s.updateCategory)
		category.DELETE("/:id", MustLogin, s.deleteCategory)
	}
	content := s.engine.Group("/v1")
	{
		content.POST("/content", s.createContent)
		content.PUT("/content/:content_id/publish", s.publishContent)
		content.PUT("/content/:content_id/review/approve", s.approve)
		content.PUT("/content/:content_id/review/reject", s.reject)
		content.GET("/content/:content_id", s.GetContent)
		content.PUT("/content/:content_id", s.updateContent)
		content.DELETE("/content/:content_id", s.deleteContent)
		content.GET("/contents", s.QueryContent)
		content.GET("/contents/private", s.QueryPrivateContent)
		content.GET("/contents/pending", s.QueryPendingContent)
	}
}
