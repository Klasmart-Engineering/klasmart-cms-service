package api

import (
	"github.com/gin-gonic/gin"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/model"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/model/storage"
	"net/http"
)

func (s *Server) getUploadPath(c *gin.Context) {
	ctx := c.Request.Context()

	partition := c.Query("partition")
	extension := c.Query("extension")

	if partition == "" || extension == "" {
		c.JSON(http.StatusBadRequest, responseMsg("partition or extension is required"))
		return
	}
	name, path, err := model.GetResourceUploaderModel().GetResourceUploadPath(ctx, partition, extension)
	switch err {
	case storage.ErrInvalidUploadPartition:
		c.JSON(http.StatusBadRequest, responseMsg(err.Error()))
	case nil:
		c.JSON(http.StatusOK, gin.H{
			"path":        path,
			"resource_id": name,
		})
	default:
		c.JSON(http.StatusInternalServerError, responseMsg(err.Error()))
	}
}

func (s *Server) getPath(c *gin.Context) {
	ctx := c.Request.Context()
	resourceId := c.Param("resource_id")

	if resourceId == "" {
		c.JSON(http.StatusBadRequest, responseMsg("resourceId is required"))
		return
	}
	path, err := model.GetResourceUploaderModel().GetResourcePath(ctx, resourceId)
	switch err {
	case model.ErrInvalidResourceId:
		c.JSON(http.StatusBadRequest, responseMsg(err.Error()))
	case storage.ErrInvalidUploadPartition:
		c.JSON(http.StatusBadRequest, responseMsg(err.Error()))
	case nil:
		c.Redirect(http.StatusFound, path)
	default:
		c.JSON(http.StatusInternalServerError, responseMsg(err.Error()))
	}
}

func (s *Server) getContentLiveToken(c *gin.Context) {
	op := GetOperator(c)
	ctx := c.Request.Context()
	contentID := c.Param("content_id")
	token, err := model.GetLiveTokenModel().MakeLivePreviewToken(ctx, op, contentID)
	if err != nil {
		log.Error(ctx, "make content live token error", log.String("contentID", contentID), log.Err(err))
		c.JSON(http.StatusInternalServerError, L(Unknown))
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"token": token,
	})
}
