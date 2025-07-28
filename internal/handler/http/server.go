package http

import (
	"fmt"
	"net/http"

	"github.com/sirawong/point-accumulate-interview/pkg/config"

	"github.com/gin-gonic/gin"
)

type HttpServer struct {
	*gin.Engine
}

func NewRouter(apHandler *AccumulatePointHandler) *HttpServer {
	router := gin.New()

	router.Use(gin.Recovery())
	router.Use(gin.Logger())

	router.POST("/api/v1/point/accumulate/upload", apHandler.UploadCSV)

	return &HttpServer{router}
}

func (h HttpServer) NewServer(cfg *config.Config) *http.Server {
	return &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.HttpServerPort),
		Handler: h,
	}
}
