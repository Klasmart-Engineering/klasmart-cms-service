package api

import (
	"github.com/gin-gonic/gin"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/model"
	"net/http"
	"strconv"
)

func (s *Server) createAsset(c *gin.Context) {
	data := new(entity.AssetObject)
	err := c.ShouldBind(data)
	if err != nil{
		c.JSON(http.StatusBadRequest, responseMsg(err.Error()))
		return
	}

	id, err := model.GetAssetModel().CreateAsset(c.Request.Context(), *data)
	if err != nil{
		c.JSON(http.StatusInternalServerError, responseMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"id": id,
	})
}
func (s *Server) updateAsset(c *gin.Context){
	id := c.Param("id")

	data := new(entity.UpdateAssetRequest)
	err := c.ShouldBind(data)
	if err != nil{
		c.JSON(http.StatusBadRequest, responseMsg(err.Error()))
		return
	}
	data.ID = id

	err = model.GetAssetModel().UpdateAsset(c.Request.Context(), *data)
	if err != nil{
		c.JSON(http.StatusInternalServerError, responseMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, responseMsg("success"))
}
func (s *Server) deleteAsset(c *gin.Context) {
	id := c.Param("id")
	err := model.GetAssetModel().DeleteAsset(c.Request.Context(), id)
	if err != nil{
		c.JSON(http.StatusInternalServerError, responseMsg(err.Error()))
		return
	}

	c.JSON(http.StatusOK, responseMsg("success"))
}

func (s *Server) getAssetByID(c *gin.Context) {
	id := c.Param("id")
	assetInfo, err := model.GetAssetModel().GetAssetByID(c.Request.Context(), id)
	if err == da.ErrRecordNotFound{
		c.JSON(http.StatusNotFound, responseMsg(err.Error()))
		return
	}
	if err != nil{
		c.JSON(http.StatusInternalServerError, responseMsg(err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"asset": assetInfo,
	})
}
func (s *Server) searchAssets(c *gin.Context){
	data := buildAssetSearchCondition(c)
	count, assetsList, err := model.GetAssetModel().SearchAssets(c.Request.Context(), data)
	if err != nil{
		c.JSON(http.StatusInternalServerError, responseMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"total": count,
		"assets": assetsList,
	})
}

func (s *Server) getAssetUploadPath(c *gin.Context) {
	ext := c.Param("ext")

	resource, err := model.GetAssetModel().GetAssetUploadPath(c.Request.Context(), ext)
	if err != nil{
		c.JSON(http.StatusInternalServerError, responseMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"path": resource.Path,
		"name": resource.Name,
	})
}


func (s *Server) getAssetResourcePath(c *gin.Context) {
	name := c.Param("name")

	path, err := model.GetAssetModel().GetAssetResourcePath(c.Request.Context(), name)
	if err != nil{
		c.JSON(http.StatusInternalServerError, responseMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"path": path,
	})
}


func responseMsg(msg string) interface{}{
	return gin.H{
		"msg": msg,
	}
}
func buildAssetSearchCondition(c *gin.Context) *entity.SearchAssetCondition{
	sizeMin, _ := strconv.Atoi("size_min")
	sizeMax, _ := strconv.Atoi("size_max")
	PageSize, _ := strconv.Atoi("page_size")
	Page, _ := strconv.Atoi("page")

	data := &entity.SearchAssetCondition{
		ID:       c.Query("id"),
		Name:     c.Query("name"),
		Category: c.Query("category"),
		SizeMin:  sizeMin,
		SizeMax:  sizeMax,
		Tag:      c.Query("tag"),
		PageSize: PageSize,
		Page:     Page,
	}

	return data
}