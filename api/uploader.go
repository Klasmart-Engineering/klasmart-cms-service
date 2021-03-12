package api

import (
	"github.com/gin-gonic/gin"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/model"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/model/storage"
	"net/http"
)

type UploadPathResponse struct {
	Path string `json:"path"`
	ResourceId string `json:"resource_id"`
}
type DownloadPathResource struct {
	Path string `json:"path"`
}
// @Summary getContentResourceUploadPath
// @ID getContentResourceUploadPath
// @Description get path to upload resource
// @Accept json
// @Produce json
// @Param partition query string true "Resource partition"
// @Param extension query string true "Resource extension"
// @Tags content
// @Success 302 {string} UploadPathResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Failure 400 {object} BadRequestResponse
// @Router /contents_resources [get]
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
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	case nil:
		c.JSON(http.StatusOK, UploadPathResponse{
			Path:       path,
			ResourceId: name,
		})
	default:
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
	}
}

// @Summary getContentResourcePath
// @ID getContentResourcePath
// @Description get the path of a resource
// @Accept json
// @Produce json
// @Param resource_id path string true "Resource id"
// @Tags content
// @Success 302 {string} string Found
// @Failure 500 {object} InternalServerErrorResponse
// @Failure 400 {object} BadRequestResponse
// @Router /contents_resources/{resource_id} [get]
func (s *Server) getContentResourcePath(c *gin.Context) {
	ctx := c.Request.Context()
	resourceId := c.Param("resource_id")

	if resourceId == "" {
		c.JSON(http.StatusBadRequest, responseMsg("resourceId is required"))
		return
	}
	path, err := model.GetResourceUploaderModel().GetResourcePath(ctx, resourceId)
	switch err {
	case model.ErrInvalidResourceID:
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	case storage.ErrInvalidUploadPartition:
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	case nil:
		c.Redirect(http.StatusFound, path)
	default:
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
	}
}


// @Summary getDownloadPath
// @ID getDownloadPath
// @Description get the path of a resource url
// @Accept json
// @Produce json
// @Param resource_id path string true "Resource id"
// @Tags content
// @Success 200 {object} DownloadPathResource
// @Failure 500 {object} InternalServerErrorResponse
// @Failure 400 {object} BadRequestResponse
// @Router /contents_resources/{resource_id}/download [get]
func (s *Server) getDownloadPath(c *gin.Context) {
	ctx := c.Request.Context()
	resourceId := c.Param("resource_id")

	if resourceId == "" {
		c.JSON(http.StatusBadRequest, responseMsg("resourceId is required"))
		return
	}
	path, err := model.GetResourceUploaderModel().GetResourcePath(ctx, resourceId)
	switch err {
	case model.ErrInvalidResourceID:
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	case storage.ErrInvalidUploadPartition:
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	case nil:
		c.JSON(http.StatusOK, DownloadPathResource{
			Path: path,
		})
	default:
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
	}
}
