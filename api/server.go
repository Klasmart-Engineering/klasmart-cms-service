package api

import (
	"context"
	"net/http"
	"time"

	_ "net/http/pprof"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/pprof"
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

	if config.Get().NewRelic.Enable() {
		// New Relic middle should insert in the very beginning as the new relic doc says.
		// And newRelicMiddlewareRectifier retrieve the txn applied by previous new relic middleware,
		// and then apply this txn to context of c.Request. Why do this is because there are a lot of
		// code in this repo use `ctx := c.Request.Context()`, which cause to the txn be lost,
		// (txn was applied to c, not context on c.Request). So we also need to apply txn to the
		// context of c.Request, the second middleware: newRelicMiddlewareRectifier do this.
		server.engine.Use(server.getNewRelicMiddleware(), server.newRelicMiddlewareRectifier())
		log.Info(context.TODO(), "new relic plugin enabled")
	} else {
		log.Warn(context.TODO(), "new relic plugin disabled because the necessary environment variables are missing!")
	}

	server.engine.Use(server.logger(), server.recovery(), server.contextStopwatch())

	// CORS
	if len(config.Get().CORS.AllowOrigins) > 0 {
		server.engine.Use(cors.New(cors.Config{
			AllowOrigins:     config.Get().CORS.AllowOrigins,
			AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD"},
			AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type"},
			AllowCredentials: true,
			AllowWildcard:    true,
			MaxAge:           12 * time.Hour,
		}))
	}

	pprof.Register(server.engine, "/v1/pprof")

	server.registeRoute()

	log.Debug(context.TODO(), "register route success")

	return server
}

func (s Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.engine.ServeHTTP(w, r)
}
