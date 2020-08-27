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
		c.JSON(http.StatusBadRequest, responseMsg("partition or extension required"))
		return
	}
	name, path, err := model.GetResourceUploaderModel().GetResourceUploadPath(ctx, partition, extension)
	if err != nil {
		c.JSON(http.StatusInternalServerError, responseMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"path":  path,
		"name": name,
	})
}


func (s *Server) GetPath(c *gin.Context) {
	ctx := c.Request.Context()

	partition := c.Query("partition")
	name := c.Query("name")

	if partition == "" || name == "" {
		c.JSON(http.StatusBadRequest, responseMsg("name or partition required"))
		return
	}
	path, err := model.GetResourceUploaderModel().GetResourcePath(ctx, partition, name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, responseMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"path":  path,
	})
}
