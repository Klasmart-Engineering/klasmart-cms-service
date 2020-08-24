package api

import (
	"github.com/gin-gonic/gin"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/model"
	"net/http"
	"strconv"
	"strings"
)

func (s *Server) createAsset(c *gin.Context) {
	data := new(entity.CreateAssetData)
	err := c.ShouldBind(data)
	if err != nil{
		c.JSON(http.StatusBadRequest, responseMsg(err.Error()))
		return
	}

	id, err := model.GetAssetModel().CreateAsset(c.Request.Context(), *data, s.getOperator(c))
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

	err = model.GetAssetModel().UpdateAsset(c.Request.Context(), *data, s.getOperator(c))
	if err == model.ErrNoAuth{
		c.JSON(http.StatusForbidden, err.Error())
		return
	}
	if err != nil{
		c.JSON(http.StatusInternalServerError, responseMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, responseMsg("success"))
}

func (s *Server) deleteAsset(c *gin.Context) {
	id := c.Param("id")

	err := model.GetAssetModel().DeleteAsset(c.Request.Context(), id, s.getOperator(c))
	if err == model.ErrNoAuth{
		c.JSON(http.StatusForbidden, err.Error())
		return
	}
	if err != nil{
		c.JSON(http.StatusInternalServerError, responseMsg(err.Error()))
		return
	}

	c.JSON(http.StatusOK, responseMsg("success"))
}

func (s *Server) getAssetByID(c *gin.Context) {
	id := c.Param("id")
	assetInfo, err := model.GetAssetModel().GetAssetByID(c.Request.Context(), id, s.getOperator(c))
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
	count, assetsList, err := model.GetAssetModel().SearchAssets(c.Request.Context(), data, s.getOperator(c))
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

	resource, err := model.GetAssetModel().GetAssetUploadPath(c.Request.Context(), ext, s.getOperator(c))
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
	name := c.Param("resource_name")

	path, err := model.GetAssetModel().GetAssetResourcePath(c.Request.Context(), name, s.getOperator(c))
	if err != nil{
		c.JSON(http.StatusInternalServerError, responseMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"path": path,
	})
}


func (s *Server) getOperator(c *gin.Context) entity.Operator{
	return entity.Operator{}
}


func responseMsg(msg string) interface{}{
	return gin.H{
		"msg": msg,
	}
}
func buildAssetSearchCondition(c *gin.Context) *entity.SearchAssetCondition{
	PageSize, _ := strconv.Atoi("page_size")
	Page, _ := strconv.Atoi("page")
	rawSearchWord := c.Query("search_words")
	isSelfStr := c.Query("is_self")
	fuzzyQuery := c.Query("fuzzy_query")
	orderBy := c.Query("order_by")

	searchWords := strings.Split(rawSearchWord, " ")
	isSelf, _ := strconv.ParseBool(isSelfStr)

	data := &entity.SearchAssetCondition{
		SearchWords:  searchWords,
		FuzzyQuery: fuzzyQuery,
		IsSelf:  isSelf,
		OrderBy: orderBy,
		PageSize: PageSize,
		Page:     Page,
	}

	return data
}