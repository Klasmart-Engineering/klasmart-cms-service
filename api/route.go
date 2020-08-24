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
	v1 := s.engine.Group("/v1")

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

	tag := s.engine.Group("/v1/tag")
	{
		tag.GET("/", s.queryTag)
		tag.GET("/:id", s.getTagByID)
		tag.POST("/", s.addTag)
		tag.PUT("/:id", s.updateTag)
		tag.DELETE("/:id", s.delTag)
	}
	content := s.engine.Group("/v1")
	{
		content.POST("/contents", MustLogin, s.createContent)
		content.GET("/contents/:content_id", MustLogin, s.GetContent)
		content.PUT("/contents/:content_id", MustLogin, s.updateContent)
		content.PUT("/contents/:content_id/lock", MustLogin, s.lockContent)
		content.PUT("/contents/:content_id/publish", MustLogin, s.publishContent)
		content.PUT("/contents_review/:content_id/approve", MustLogin, s.approve)
		content.PUT("/contents_review/:content_id/reject", MustLogin, s.reject)
		content.DELETE("/contents/:content_id", MustLogin, s.deleteContent)
		content.GET("/contents", MustLogin, s.QueryContent)
		content.GET("/contents_dynamo", MustLogin, s.QueryDynamoContent)
		content.GET("/contents_private", MustLogin, s.QueryPrivateContent)
		content.GET("/contents_pending", MustLogin, s.QueryPendingContent)

	}

	v1.PUT("/schedules/:id", s.updateSchedule)
	v1.DELETE("/schedules/:id", s.deleteSchedule)
	v1.POST("/schedules", s.addSchedule)
	v1.GET("/schedules/:id", s.getScheduleByID)
	v1.GET("/schedules", s.querySchedule)
	v1.GET("/schedules_home", s.queryHomeSchedule)
}
