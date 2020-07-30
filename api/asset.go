package api

import (
	"github.com/gin-gonic/gin"
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
	data.Id = id

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

func (s *Server) getAssetById(c *gin.Context) {
	id := c.Param("id")
	assetInfo, err := model.GetAssetModel().GetAssetById(c.Request.Context(), id)
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
	assetsList, err := model.GetAssetModel().SearchAssets(c.Request.Context(), data)
	if err != nil{
		c.JSON(http.StatusInternalServerError, responseMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"assets": assetsList,
	})
}

func (s *Server) getAssetUploadPath(c *gin.Context) {
	ext := c.Param("ext")

	path, err := model.GetAssetModel().GetAssetUploadPath(c.Request.Context(), ext)
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
func buildAssetSearchCondition(c *gin.Context) *model.SearchAssetCondition{
	sizeMin, _ := strconv.Atoi("size_min")
	sizeMax, _ := strconv.Atoi("size_max")
	PageSize, _ := strconv.Atoi("page_size")
	Page, _ := strconv.Atoi("page")

	data := &model.SearchAssetCondition{
		Id:       c.Query("id"),
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