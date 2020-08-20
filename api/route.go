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
		content.POST("/contents", s.createContent)
		content.PUT("/contents/:content_id/publish", s.publishContent)
		content.PUT("/contents_review/:content_id/approve", s.approve)
		content.PUT("/contents_review/:content_id/reject", s.reject)
		content.GET("/contents/:content_id", s.GetContent)
		content.PUT("/contents/:content_id", s.updateContent)
		content.DELETE("/contents/:content_id", s.deleteContent)
		content.GET("/contents", s.QueryContent)
		content.GET("/contents_private", s.QueryPrivateContent)
		content.GET("/contents_pending", s.QueryPendingContent)
	}
}
