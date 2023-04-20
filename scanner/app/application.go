package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	base_broker "github.com/Exca-DK/node-util/service/broker"
	store "github.com/Exca-DK/node-util/service/db"
	"github.com/gin-gonic/gin"

	"github.com/Exca-DK/node-util/scanner/broker"
	"github.com/Exca-DK/node-util/scanner/log"
	"github.com/Exca-DK/node-util/scanner/metrics"
	"github.com/Exca-DK/node-util/scanner/scanner"
)

var (
	VersionMajor = 0     // Major version component of the current release
	VersionMinor = 0     // Minor version component of the current release
	VersionPatch = 1     // Patch version component of the current release
	VersionMeta  = "dev" // Version metadata to append to the version string
)

func Version() string {
	return fmt.Sprintf("%d.%d.%d", VersionMajor, VersionMinor, VersionPatch)
}

func VersionWithMeta() string {
	v := Version()
	if VersionMeta != "" {
		v += "-" + VersionMeta
	}
	return v
}

func NewScanApp(cfg *Config) *Application {
	app := &Application{Config: cfg, log: log.NewLogger()}
	return app
}

type Application struct {
	*Config
	http *http.Server

	broker base_broker.Broker
	cache  *scanner.RedisClient
	store  store.Store

	log log.Logger
}

func (app *Application) Setup() error {
	app.log.Info("running setup")
	if app.Config.Metrics {
		engine := gin.Default()
		app.log.Info("metrics enabled")
		metrics.RegisterMetricsHandler(engine)
		app.http = &http.Server{Addr: fmt.Sprintf(":%v", 6061), Handler: engine}
		go app.http.ListenAndServe()
	}

	if app.Config.BrokerEnabled {
		broker, err := base_broker.NewRabbitBroker(app.Config.BrokerEndpoint, app.Config.BrokerUser, app.Config.BrokerPassword, base_broker.NewRecorder(metrics.Factory))
		if err != nil {
			return fmt.Errorf("broker initializationn failure. error: %v", err)
		}
		app.broker = broker
	} else {
		return errors.New("broker is required")
	}

	if app.Config.CacheEnabled {
		cache := scanner.RedisClient{Log: app.log}
		if err := cache.Start(app.CacheEndpoint, app.CachePassword); err != nil {
			return err
		}
		app.cache = &cache
	} else {
		return errors.New("cache is required")
	}

	if app.Config.StoreEnabled {
		store, err := store.NewStore(store.Config{Source: app.Config.StoreSource, MigrationPath: app.Config.StoreMigration})
		if err != nil {
			return err
		}
		app.store = store
	} else {
		return errors.New("store is required")
	}

	app.log.Info("setup finished")
	return nil
}

func (app *Application) Run(ctx context.Context) error {
	controller := scanner.NewScannerController(app.Threads, app.Burst, app.Interval, app.log)
	scanner := scanner.NewNodeScanner(controller, log.NewLogger(), app.store, broker.NewBroker(app.broker), app.cache)
	return scanner.Start(ctx)
}

func (app *Application) Stop() {
	if app.http == nil {
		return
	}
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(15*time.Second))
	defer cancel()
	app.http.Shutdown(ctx)
}
