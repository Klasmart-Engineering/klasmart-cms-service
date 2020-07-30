package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
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

	server.registeRoute()

	return server
}

func (s Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.engine.ServeHTTP(w, r)
}
