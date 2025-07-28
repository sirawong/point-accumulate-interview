package di

import (
	"github.com/sirawong/point-accumulate-interview/internal/handler/http"
	customerdb "github.com/sirawong/point-accumulate-interview/internal/repository/mongodb/customers"
	rulesdb "github.com/sirawong/point-accumulate-interview/internal/repository/mongodb/rules"
	"github.com/sirawong/point-accumulate-interview/internal/services/accumulatepoints"
	"github.com/sirawong/point-accumulate-interview/pkg/config"
	"github.com/sirawong/point-accumulate-interview/pkg/database/mongodb"
)

func NewNewApplication() (*Application, func(), error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, nil, err
	}

	db, cleanup, err := mongodb.NewMongoConn(cfg)
	if err != nil {
		return nil, nil, err
	}

	customerRepo := customerdb.NewCustomerRepository(db)
	rulesRepo := rulesdb.NewRuleRepository(db)

	accumulatePointsSrv := accumulatepoints.NewAccumulatePointService(rulesRepo, customerRepo, cfg)

	apHandler := http.NewAccumulatePointHandler(accumulatePointsSrv, cfg)
	httpRouter := http.NewRouter(apHandler)
	httpServer := httpRouter.NewServer(cfg)

	return &Application{
			httpServer: httpServer,
			Cfg:        cfg,
		}, func() {
			cleanup()
		}, nil
}
