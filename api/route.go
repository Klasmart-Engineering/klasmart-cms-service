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
		content.GET("/contents/:content_id", MustLogin, s.getContent)
		content.PUT("/contents/:content_id", MustLogin, s.updateContent)
		content.PUT("/contents/:content_id/lock", MustLogin, s.lockContent)
		content.PUT("/contents/:content_id/publish", MustLogin, s.publishContent)
		content.PUT("/contents_review/:content_id/approve", MustLogin, s.approve)
		content.PUT("/contents_review/:content_id/reject", MustLogin, s.reject)
		content.DELETE("/contents/:content_id", MustLogin, s.deleteContent)
		content.GET("/contents", MustLogin, s.queryContent)
		content.GET("/contents_dynamo", MustLogin, s.queryDynamoContent)
		content.GET("/contents_private", MustLogin, s.queryPrivateContent)
		content.GET("/contents_pending", MustLogin, s.queryPendingContent)

		content.GET("/contents_statistics", MustLogin, s.contentDataCount)

		content.PUT("/contents_bulk/publish", MustLogin, s.publishContentBulk)
		content.DELETE("/contents_bulk", MustLogin, s.deleteContentBulk)

		content.GET("/contents_resources", MustLogin, s.getUploadPath)
		content.GET("/contents_resources/:resource_id", MustLogin, s.getPath)
	}
	schedules := s.engine.Group("/v1")
	{
		schedules.PUT("/schedules/:id", MustLogin, s.updateSchedule)
		schedules.DELETE("/schedules/:id", MustLogin, s.deleteSchedule)
		schedules.POST("/schedules", MustLogin, s.addSchedule)
		schedules.GET("/schedules/:id", MustLogin, s.getScheduleByID)
		schedules.GET("/schedules", MustLogin, s.querySchedule)
		schedules.GET("/schedules_time_view", MustLogin, s.getScheduleTimeView)
		//schedules.GET("/schedule_attachment_upload/:ext", MustLogin, s.getAttachmentUploadPath)
	}
}

