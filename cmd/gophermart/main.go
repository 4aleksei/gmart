package main

import (
	"time"

	"github.com/4aleksei/gmart/internal/common/logger"
	"github.com/4aleksei/gmart/internal/common/store"

	"github.com/4aleksei/gmart/internal/common/store/pg"
	"github.com/4aleksei/gmart/internal/common/utils"
	"github.com/4aleksei/gmart/internal/gophermart/config"
	"github.com/4aleksei/gmart/internal/gophermart/handlers"
	"github.com/4aleksei/gmart/internal/gophermart/service"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
)

func main() {
	setupFX().Run()
}

func setupFX() *fx.App {
	app := fx.New(
		fx.Supply(logger.Config{Level: "debug"}),
		fx.StopTimeout(1*time.Minute),
		fx.Provide(
			logger.New,
			config.GetConfig,
			fx.Annotate(pg.New,
				fx.As(new(service.ServiceStore)), fx.As(new(store.Store))),

			service.NewService,
			handlers.NewHTTPServer,

			/*fx.Annotate(memstoragemux.NewStoreMux,
				fx.As(new(service.AgentMetricsStorage))),
			service.NewHandlerStore,
			gather.NewAppGather,
			gatherps.NewGather,
			httpclientpool.NewHandler,
			handlers.NewApp,*/
		),

		fx.WithLogger(func(log *logger.ZapLogger) fxevent.Logger {
			return &fxevent.ZapLogger{Logger: log.Logger}
		}),
		fx.Invoke(
			registerSetLoggerLevel,

			gooseUP,

			registerStorePg,

			registerHTTPServer,
		),
	)
	return app
}

func gooseUP(cfg *config.Config, ll *logger.ZapLogger) {
	if err := migrate(cfg.DatabaseURI, ll); err != nil {
		ll.Logger.Fatal("migrate fatal", zap.Error(err))
	}
}

func registerStorePg(ss store.Store, cfg *config.Config, lc fx.Lifecycle) {
	switch v := ss.(type) {
	case *pg.PgStore:
		v.DatabaseURI = cfg.DatabaseURI
		lc.Append(utils.ToHook(v))
	default:
	}
}

func registerHTTPServer(hh *handlers.HandlersServer, lc fx.Lifecycle) {
	lc.Append(utils.ToHook(hh))
}

func registerSetLoggerLevel(ll *logger.ZapLogger, cfg *config.Config, lc fx.Lifecycle) {
	ll.SetLevel(cfg.LCfg)
	lc.Append(utils.ToHook(ll))
}
