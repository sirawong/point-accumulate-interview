package di

import (
	"net/http"

	"github.com/sirawong/point-accumulate-interview/pkg/config"
)

type Application struct {
	httpServer *http.Server
	Cfg        *config.Config
}
