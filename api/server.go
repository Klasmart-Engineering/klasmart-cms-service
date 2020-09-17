package api

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
)

// Server api server
type Server struct {
	engine *gin.Engine
}

// NewServer create api server
func NewServer() *Server {
	gin.SetMode(gin.ReleaseMode)
	server := &Server{
		engine: gin.New(),
	}

	log.Debug(context.TODO(), "init gin success")

	server.engine.Use(logger(), recovery())

	server.registeRoute()

	log.Debug(context.TODO(), "register route success")

	return server
}

func (s Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.engine.ServeHTTP(w, r)
}
