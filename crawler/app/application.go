package app

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/Exca-DK/node-util/crawler/broker"
	querier "github.com/Exca-DK/node-util/crawler/net"
	base_broker "github.com/Exca-DK/node-util/service/broker"
	store "github.com/Exca-DK/node-util/service/db"

	crawl "github.com/Exca-DK/node-util/crawler/crawler"
	"github.com/Exca-DK/node-util/crawler/log"
	"github.com/Exca-DK/node-util/crawler/metrics"
	"github.com/Exca-DK/node-util/crawler/net/discovery"
	"github.com/Exca-DK/node-util/crawler/net/discovery/interfaces"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/gin-gonic/gin"
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

type Config struct {
	Metrics   bool
	Port      int
	Bootnodes []*enode.Node
	Handler   discovery.SubHandler
	Conn      net.PacketConn
	Qurier    *querier.PeerQuierier

	Keys  []string
	Store store.Store

	BrokerEnabled  bool
	BrokerEndpoint string
	BrokerUser     string
	BrokerPassword string
}

func NewCrawlApp(cfg Config) *Application {
	app := &Application{Config: cfg, log: log.NewLogger()}

	engine := gin.Default()
	if cfg.Metrics {
		app.log.Info("metrics enabled")
		metrics.RegisterMetricsHandler(engine)
	}

	//setup base generic broker
	if cfg.BrokerEnabled {
		app.log.Info("broker enabled")
		broker, err := base_broker.NewRabbitBroker(cfg.BrokerEndpoint, cfg.BrokerUser, cfg.BrokerPassword, base_broker.NewRecorder(metrics.Factory))
		if err != nil {
			panic(fmt.Errorf("rabbit broker initializationn failure. error: %v", err))
		}
		app.broker = broker
	} else {
		app.broker = base_broker.NewVoidBroker()
	}

	app.http = http.Server{Addr: fmt.Sprintf(":%v", cfg.Port), Handler: engine}
	db, _ := enode.OpenDB("")
	handler, listener, err := MakeHandler(cfg.Conn, cfg.Bootnodes, db, cfg.Handler)
	if err != nil {
		panic(fmt.Errorf("failed making handler. error: %v", err))
	}
	app.listener = listener
	app.handler = handler
	return app
}

type Application struct {
	Config
	http     http.Server
	handler  discovery.IHandler
	listener discovery.Listener

	broker base_broker.Broker

	log log.Logger

	wg sync.WaitGroup
}

func (app *Application) Run(ctx context.Context) {
	defer app.listener.Stop()
	app.wg.Add(1)
	go app.runHttp(ctx)
	app.wg.Add(1)
	go app.runCrawler(ctx)
	go app.runHandler(ctx)
	app.wg.Wait()
}

func (app *Application) runCrawler(ctx context.Context) {
	defer app.wg.Done()
	broker := broker.NewBroker(app.broker) //wrap generic broker into concrete type
	crawler := crawl.NewCrawler(app.handler, app.Qurier, broker, app.Store)
	for _, key := range app.Keys {
		crawler.AddEnrKey(key)
	}
	crawler.Crawl(ctx)
	crawler.Wait()
}

func (app *Application) runHttp(ctx context.Context) {
	defer app.wg.Done()
	app.http.ListenAndServe()
}

func (app *Application) runHandler(ctx context.Context) {
	defer app.wg.Done()
	app.handler.Run()
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			app.handler.LookupRandom()
		}
	}
}

func MakeHandler(conn net.PacketConn, boot []*enode.Node, db *enode.DB, subhandler discovery.SubHandler) (discovery.IHandler, discovery.Listener, error) {
	writer := discovery.BuildWriter(conn.(*net.UDPConn), log.NewLogger()).
		ThrottleIp(1*time.Second, 32).        //be nice and dont request node more than 32 times a second
		ThrottlePackets(1*time.Second, 8192). //more than that can lead to network abuse...
		Build()
	handler, err := discovery.NewHandler(writer, subhandler, db, discovery.Config{Bootnodes: boot, RunTab: false, IgnoreLimits: true})
	if err != nil {
		return nil, nil, err
	}

	return handler, discovery.NewListener(conn.(*net.UDPConn), log.NewLogger(), discovery.NewPoolAcceptor(handler, 4)), nil
}

func NewQuierier(id interfaces.IdentityHolder) *querier.PeerQuierier {
	return querier.NewPeerQuierier(id, log.NewLogger())
}
