package api

import (
	"github.com/gin-gonic/gin"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/model"
	"net/http"
)

func (s *Server) GetUploadPath(c *gin.Context) {
	ctx := c.Request.Context()

	partition := c.Query("partition")
	extension := c.Query("extension")

	if partition == "" || extension == "" {
		c.JSON(http.StatusBadRequest, responseMsg("partition or extension is required"))
		return
	}
	name, path, err := model.GetResourceUploaderModel().GetResourceUploadPath(ctx, partition, extension)
	if err != nil {
		c.JSON(http.StatusInternalServerError, responseMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"path":       path,
		"resource_id": name,
	})
}


func (s *Server) GetPath(c *gin.Context) {
	ctx := c.Request.Context()
	resourceId := c.Param("resource_id")

	if resourceId == "" {
		c.JSON(http.StatusBadRequest, responseMsg("resourceId is required"))
		return
	}
	path, err := model.GetResourceUploaderModel().GetResourcePath(ctx, resourceId)
	if err == model.ErrInvalidResourceId {
		c.JSON(http.StatusBadRequest, responseMsg(err.Error()))
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, responseMsg(err.Error()))
		return
	}
	c.Redirect(http.StatusFound, path)
}
