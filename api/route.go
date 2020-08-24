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

	v1.PUT("/schedules/:id", s.updateSchedule)
	v1.DELETE("/schedules/:id", s.deleteSchedule)
	v1.POST("/schedules", s.addSchedule)
	v1.GET("/schedules/:id", s.getScheduleByID)
	v1.GET("/schedules", s.querySchedule)
	v1.GET("/schedules_home", s.queryHomeSchedule)
}
