package api

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"
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

	server.engine.Use(server.logger(), server.recovery())

	// CORS
	if len(config.Get().CORS.AllowOrigins) > 0 {
		// server.engine.Use(cors.New(cors.Config{
		// 	AllowOrigins:     config.Get().CORS.AllowOrigins,
		// 	AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD"},
		// 	AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type"},
		// 	AllowCredentials: true,
		// 	MaxAge:           12 * time.Hour,
		// }))
	}

	server.registeRoute()

	log.Debug(context.TODO(), "register route success")

	return server
}

func (s Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.engine.ServeHTTP(w, r)
}
