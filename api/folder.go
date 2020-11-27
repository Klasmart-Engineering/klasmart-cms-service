package api

import (
	"github.com/gin-gonic/gin"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/model"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
	"net/http"
)

// @Summary createFolder
// @ID createFolder
// @Description create folder
// @Accept json
// @Produce json
// @Param content body entity.CreateFolderRequest true "create request"
// @Tags folder
// @Success 200 {object} CreateFolderResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Failure 400 {object} BadRequestResponse
// @Router /folders [post]
func (s *Server) createFolder(c *gin.Context) {
	ctx := c.Request.Context()
	op := s.getOperator(c)
	var data entity.CreateFolderRequest
	err := c.ShouldBind(&data)
	if err != nil {
		log.Warn(ctx, "create folder failed", log.Err(err))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}

	cid, err := model.GetFolderModel().CreateFolder(ctx, data, op)
	switch err {
	case nil:
		c.JSON(http.StatusOK, gin.H{
			"id": cid,
		})
	default:
		c.JSON(http.StatusInternalServerError, responseMsg(err.Error()))
	}
}

// @Summary addFolderItem
// @ID addFolderItem
// @Description create folder item
// @Accept json
// @Produce json
// @Param content body entity.CreateFolderItemRequest true "create request"
// @Tags folder
// @Success 200 {object} CreateFolderResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Failure 400 {object} BadRequestResponse
// @Router /folders/items [post]
func (s *Server) addFolderItem(c *gin.Context){
	ctx := c.Request.Context()
	op := s.getOperator(c)
	var data entity.CreateFolderItemRequest
	err := c.ShouldBind(&data)
	if err != nil {
		log.Warn(ctx, "add folder item failed", log.Err(err))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}

	cid, err := model.GetFolderModel().AddItem(ctx, data, op)
	switch err {
	case nil:
		c.JSON(http.StatusOK, gin.H{
			"id": cid,
		})
	default:
		c.JSON(http.StatusInternalServerError, responseMsg(err.Error()))
	}
}

// @Summary removeFolderItem
// @ID removeFolderItem
// @Description remove folder item
// @Accept json
// @Produce json
// @Param content body entity.CreateFolderItemRequest true "create request"
// @Tags folder
// @Success 200 {object} string ok
// @Failure 500 {object} InternalServerErrorResponse
// @Failure 400 {object} BadRequestResponse
// @Router /folders/items/{item_id} [delete]
func (s *Server) removeFolderItem(c *gin.Context){
	ctx := c.Request.Context()
	op := s.getOperator(c)
	fid := c.Param("item_id")
	err := model.GetFolderModel().RemoveItem(ctx, fid, op)

	switch err {
	case nil:
		c.JSON(http.StatusOK, "")
	default:
		c.JSON(http.StatusInternalServerError, responseMsg(err.Error()))
	}
}

// @Summary updateFolderItem
// @ID updateFolderItem
// @Description update folder item info
// @Accept json
// @Produce json
// @Param content body entity.UpdateFolderRequest true "update request"
// @Tags folder
// @Success 200 {object} string ok
// @Failure 500 {object} InternalServerErrorResponse
// @Failure 400 {object} BadRequestResponse
// @Router /folders/items/details/{item_id} [put]
func (s *Server) updateFolderItem(c *gin.Context){
	ctx := c.Request.Context()
	op := s.getOperator(c)
	fid := c.Param("item_id")

	var data entity.UpdateFolderRequest
	err := c.ShouldBind(&data)
	if err != nil {
		log.Warn(ctx, "update folder item failed", log.Err(err))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}
	err = model.GetFolderModel().UpdateFolder(ctx, fid, data, op)

	switch err {
	case nil:
		c.JSON(http.StatusOK, "")
	default:
		c.JSON(http.StatusInternalServerError, responseMsg(err.Error()))
	}
}

// @Summary moveFolderItem
// @ID moveFolderItem
// @Description move folder item
// @Accept json
// @Produce json
// @Param content body entity.MoveFolderRequest true "move folder request"
// @Tags folder
// @Success 200 {object} string ok
// @Failure 500 {object} InternalServerErrorResponse
// @Failure 400 {object} BadRequestResponse
// @Failure 406 {object} BadRequestResponse
// @Router /folders/items/move [put]
func (s *Server) moveFolderItem(c *gin.Context){
	ctx := c.Request.Context()
	op := s.getOperator(c)
	var data entity.MoveFolderRequest
	err := c.ShouldBind(&data)
	if err != nil {
		log.Warn(ctx, "update folder item failed", log.Err(err))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}

	err = model.GetFolderModel().MoveItem(ctx, data, op)
	switch err {
	case nil:
		c.JSON(http.StatusOK, "")
	case model.ErrMoveToChild:
		c.JSON(http.StatusNotAcceptable, L(GeneralUnknown))
	default:
		c.JSON(http.StatusInternalServerError, responseMsg(err.Error()))
	}
}



// @Summary moveFolderItemBulk
// @ID moveFolderItemBulk
// @Description bulk move folder item
// @Accept json
// @Produce json
// @Param content body entity.MoveFolderIDBulkRequest true "move folder item buck request"
// @Tags folder
// @Success 200 {object} string ok
// @Failure 500 {object} InternalServerErrorResponse
// @Failure 400 {object} BadRequestResponse
// @Failure 406 {object} BadRequestResponse
// @Router /folders/items/bulk/move [put]
func (s *Server) moveFolderItemBulk(c *gin.Context){
	ctx := c.Request.Context()
	op := s.getOperator(c)
	var data entity.MoveFolderIDBulkRequest
	err := c.ShouldBind(&data)
	if err != nil {
		log.Warn(ctx, "update folder item failed", log.Err(err))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}

	err = model.GetFolderModel().MoveItemBulk(ctx, data, op)
	switch err {
	case nil:
		c.JSON(http.StatusOK, "")
	case model.ErrMoveToChild:
		c.JSON(http.StatusNotAcceptable, L(GeneralUnknown))
	default:
		c.JSON(http.StatusInternalServerError, responseMsg(err.Error()))
	}
}

// @Summary listFolderItems
// @ID listFolderItems
// @Description list folder items
// @Accept json
// @Produce json
// @Param item_type query integer false "list items type. 1.folder 2.file"
// @Tags folder
// @Success 200 {object} FolderItemsResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /folders/items/list/{folder_id} [get]
func (s *Server) listFolderItems(c *gin.Context){
	ctx := c.Request.Context()
	op := s.getOperator(c)
	fid := c.Param("folder_id")
	itemType := utils.ParseInt(ctx, c.Query("item_type"))
	items, err := model.GetFolderModel().ListItems(ctx, fid, entity.NewItemType(itemType), op)
	switch err {
	case nil:
		c.JSON(http.StatusOK, FolderItemsResponse{Items: items})
	default:
		c.JSON(http.StatusInternalServerError, responseMsg(err.Error()))
	}
}

// @Summary searchPrivateFolderItems
// @ID searchPrivateFolderItems
// @Description search user's private folder items
// @Accept json
// @Produce json
// @Param name query string false "search content name"
// @Param item_type query integer false "list items type. 1.folder 2.file"
// @Param owner_type query integer false "list items owner type. 1.org folder 2.private folder"
// @Param partition query string false "list items type. [assets, plans and materials]"
// @Param parent_id query string false "list items from parent"
// @Param path query string false "list items in path"
// @Param order_by query string false "search content order by column name" Enums(id, -id, create_at, -create_at, update_at, -update_at)
// @Param page query int false "content list page index"
// @Param page_size query int false "content list page size"
// @Tags folder
// @Success 200 {object} FolderItemsResponseWithTotal
// @Failure 500 {object} InternalServerErrorResponse
// @Router /folders/items/search/private [get]
func (s *Server) searchPrivateFolderItems(c *gin.Context){
	ctx := c.Request.Context()
	op := s.getOperator(c)
	condition := s.buildFolderCondition(c)

	total, items, err := model.GetFolderModel().SearchPrivateFolder(ctx, *condition, op)
	switch err {
	case nil:
		c.JSON(http.StatusOK, FolderItemsResponseWithTotal{Items: items, Total: total})
	default:
		c.JSON(http.StatusInternalServerError, responseMsg(err.Error()))
	}
}

// @Summary searchOrgFolderItems
// @ID searchOrgFolderItems
// @Description search folder items in org
// @Accept json
// @Produce json
// @Param name query string false "search content name"
// @Param item_type query integer false "list items type. 1.folder 2.file"
// @Param owner_type query integer false "list items owner type. 1.org folder 2.private folder"
// @Param partition query string false "list items type. [assets, plans and materials]"
// @Param parent_id query string false "list items from parent"
// @Param path query string false "list items in path"
// @Param order_by query string false "search content order by column name" Enums(id, -id, create_at, -create_at, update_at, -update_at)
// @Param page query int false "content list page index"
// @Param page_size query int false "content list page size"
// @Tags folder
// @Success 200 {object} FolderItemsResponseWithTotal
// @Failure 500 {object} InternalServerErrorResponse
// @Router /folders/items/search/org [get]
func (s *Server) searchOrgFolderItems(c *gin.Context){
	ctx := c.Request.Context()
	op := s.getOperator(c)
	condition := s.buildFolderCondition(c)

	total, items, err := model.GetFolderModel().SearchOrgFolder(ctx, *condition, op)
	switch err {
	case nil:
		c.JSON(http.StatusOK, FolderItemsResponseWithTotal{Items: items, Total: total})
	default:
		c.JSON(http.StatusInternalServerError, responseMsg(err.Error()))
	}
}

// @Summary getFolderItemByID
// @ID getFolderItemByID
// @Description get a folder item by id
// @Accept json
// @Produce json
// @Tags folder
// @Success 200 {object} entity.FolderItemInfo
// @Failure 500 {object} InternalServerErrorResponse
// @Failure 404 {object} BadRequestResponse
// @Router /folders/items/details/{folder_id} [get]
func (s *Server) getFolderItemByID(c *gin.Context){
	ctx := c.Request.Context()
	op := s.getOperator(c)
	fid := c.Param("folder_id")
	item, err := model.GetFolderModel().GetFolderByID(ctx, fid, op)
	switch err {
	case nil:
		c.JSON(http.StatusOK, item)
	case model.ErrNoFolder:
		c.JSON(http.StatusNotFound, L(GeneralUnknown))
	default:
		c.JSON(http.StatusInternalServerError, responseMsg(err.Error()))
	}
}

func (s *Server) buildFolderCondition(c *gin.Context) *entity.SearchFolderCondition{
	ctx := c.Request.Context()
	//OwnerType OwnerType
	ownerType := utils.ParseInt(ctx, c.Query("owner_type"))
	itemType := utils.ParseInt(ctx, c.Query("item_type"))
	parentID := c.Query("parent_id")
	path := c.Query("path")
	name := c.Query("name")
	orderBy := c.Query("order_by")
	pageSize := utils.ParseInt64(ctx, c.Query("page_size"))
	pageIndex := utils.ParseInt64(ctx, c.Query("page"))
	partition := c.Query("partition")
	//Pager   utils.Pager
	return &entity.SearchFolderCondition{
		IDs:       nil,
		OwnerType: entity.NewOwnerType(ownerType),
		ItemType:  entity.NewItemType(itemType),
		ParentID:  parentID,
		Path:      path,
		Name:      name,
		Partition: partition,
		OrderBy:   orderBy,
		Pager:     utils.Pager{
			PageIndex: pageIndex,
			PageSize:  pageSize,
		},
	}
}
type FolderItemsResponse struct {
	Items []*entity.FolderItem `json:"items"`
}
type FolderItemsResponseWithTotal struct {
	Items []*entity.FolderItem `json:"items"`
	Total int `json:"total"`
}
