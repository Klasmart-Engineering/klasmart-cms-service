package route

import (

	"context"
	"errors"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"net/http"
	"strings"
)

var(
	ErrInvalidAction = errors.New("invalid action pairs")
	ErrUnknownActionPrefix = errors.New("unknown action prefix")
)

type Server interface {
	DoHandler(ctx context.Context, action string, body []byte)*entity.Response
	Prefix() string
}

var serverMap map[string]Server

func init(){
	serverMap = make(map[string]Server)

	asset := new(AssetServer)
	serverMap[asset.Prefix()] = asset
}

func actionPairs(action string) (string ,string, error){
	pairs := strings.Split(action, ".")
	if len(pairs) != 2{
		return "", "", ErrInvalidAction
	}
	return pairs[0], pairs[1], nil
}

func Route(ctx context.Context, action string, body []byte)(*entity.Response, error){
	prefix, function, err := actionPairs(action)
	if err != nil{
		return nil, err
	}
	s, ok := serverMap[prefix]
	if !ok {
		return entity.NewErrorResponse(http.StatusNotFound, "Action prefix not found"), nil
	}
	return s.DoHandler(ctx, function, body), nil
}