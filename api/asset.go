package api

import (
	"context"
	"encoding/json"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/model"
	"net/http"
)

type AssetServer struct {

}

func (as *AssetServer) DoHandler(ctx context.Context, action string, body []byte)*entity.Response{
	switch action {
	case "create":
		return as.createAsset(ctx, body)
	case "delete":
		return as.deleteAsset(ctx, body)
	case "update":
		return as.updateAsset(ctx, body)
	case "get":
		return as.getAssetById(ctx, body)
	case "search":
		return as.searchAssets(ctx, body)
	case "upload":
		return as.getAssetUploadPath(ctx, body)
	}
	return entity.NewErrorResponse(http.StatusNotFound, "Action not found")
}

func (as AssetServer) Prefix() string{
	return "asset"
}

func (as *AssetServer) createAsset(ctx context.Context, body []byte) *entity.Response{
	data := new(entity.AssetObject)
	err := json.Unmarshal(body, data)
	if err != nil{
		return entity.NewErrorResponse(http.StatusBadRequest, err.Error())
	}

	id, err := model.GetAssetModel().CreateAsset(ctx, *data)
	if err != nil{
		return entity.NewErrorResponse(http.StatusInternalServerError, err.Error())
	}
	return &entity.Response{
		StatusCode: http.StatusOK,
		StatusMsg: entity.IdMsg{Id: id},
	}
}
func (as *AssetServer) updateAsset(ctx context.Context, body []byte) *entity.Response{
	data := new(entity.UpdateAssetRequest)
	err := json.Unmarshal(body, data)
	if err != nil{
		return entity.NewErrorResponse(http.StatusBadRequest, err.Error())
	}
	err = model.GetAssetModel().UpdateAsset(ctx, *data)
	if err != nil{
		return entity.NewErrorResponse(http.StatusInternalServerError, err.Error())
	}
	return &entity.Response{
		StatusCode: http.StatusOK,
	}
}
func (as *AssetServer) deleteAsset(ctx context.Context, body []byte) *entity.Response{
	data := new(entity.IdMsg)
	err := json.Unmarshal(body, data)
	if err != nil{
		return entity.NewErrorResponse(http.StatusBadRequest, err.Error())
	}
	err = model.GetAssetModel().DeleteAsset(ctx, data.Id)
	if err != nil{
		return entity.NewErrorResponse(http.StatusInternalServerError, err.Error())
	}
	return &entity.Response{
		StatusCode: http.StatusOK,
	}
}

func (as *AssetServer) getAssetById(ctx context.Context, body []byte) *entity.Response{
	data := new(entity.IdMsg)
	err := json.Unmarshal(body, data)
	if err != nil{
		return entity.NewErrorResponse(http.StatusBadRequest, err.Error())
	}
	assetInfo, err := model.GetAssetModel().GetAssetById(ctx, data.Id)
	if err != nil{
		return entity.NewErrorResponse(http.StatusInternalServerError, err.Error())
	}
	return &entity.Response{
		StatusCode: http.StatusOK,
		StatusMsg: assetInfo,
	}
}
func (as *AssetServer) searchAssets(ctx context.Context, body []byte) *entity.Response{
	data := new(model.SearchAssetCondition)
	err := json.Unmarshal(body, data)
	if err != nil{
		return entity.NewErrorResponse(http.StatusBadRequest, err.Error())
	}
	assetsList, err := model.GetAssetModel().SearchAssets(ctx, data)
	if err != nil{
		return entity.NewErrorResponse(http.StatusInternalServerError, err.Error())
	}
	return &entity.Response{
		StatusCode: http.StatusOK,
		StatusMsg: assetsList,
	}
}

func (as *AssetServer) getAssetUploadPath(ctx context.Context, body []byte) *entity.Response{
	data := new(entity.FileExtensionRequest)
	err := json.Unmarshal(body, data)
	if err != nil{
		return entity.NewErrorResponse(http.StatusBadRequest, err.Error())
	}
	path, err := model.GetAssetModel().GetAssetUploadPath(ctx, data.Extension)
	if err != nil{
		return entity.NewErrorResponse(http.StatusInternalServerError, err.Error())
	}
	return &entity.Response{
		StatusCode: http.StatusOK,
		StatusMsg: entity.PathRequest{Path: path},
	}
}