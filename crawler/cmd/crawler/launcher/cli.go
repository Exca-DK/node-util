package launcher

import (
	"errors"
	"fmt"
	"net"

	application "github.com/Exca-DK/node-util/crawler/app"
	"github.com/Exca-DK/node-util/crawler/log"
	"github.com/Exca-DK/node-util/crawler/net/discovery"
	v4 "github.com/Exca-DK/node-util/crawler/net/discovery/v4"
	"github.com/urfave/cli/v2"
)

var (
	CrawlerCommand = &cli.Command{
		Name:        "crawl",
		Description: "crawls network using v4 discovery protocol",
		Action:      runCrawler,
		Flags: []cli.Flag{application.NodesFlag,
			application.MetricsFlag,
			application.IdentitiesFlag,
			application.DiscPort,
			application.FakeName,
			application.FakeCaps,
			application.BrokerEnabled,
			application.BrokerEndpoint,
			application.BrokerPassword,
			application.BrokerUser,
			application.DbSource,
			application.DbMigrationUrl,
		},
	}
)

func runCrawler(ctx *cli.Context) error {
	logger := log.NewLogger()
	fakename := application.ParseFakeName(ctx)
	ids := application.ParseIdentities(ctx)
	metrics := application.ParseMetrics(ctx)
	bootnodes, err := application.ParseNodes(ctx)
	if err != nil {
		return err
	} else if len(bootnodes) == 0 {
		return errors.New("atleast 1 bootnode required")
	}
	fakecaps, err := application.ParseFakeCaps(ctx)
	if err != nil {
		return err
	}
	broker_username, broker_password, broker_endpoint, broker_enabled, err := application.ParseBroker(ctx)
	if err != nil {
		return err
	}

	store, err := application.ParseDb(ctx)
	if err != nil {
		return err
	}

	socket, err := net.ListenPacket("udp4", fmt.Sprintf("0.0.0.0:%v", application.ParseDiscPort(ctx)))
	if err != nil {
		return err
	}

	addr := socket.LocalAddr().(*net.UDPAddr)
	holder := discovery.NewRandomIdentity(fakename, addr.IP, addr.AddrPort().Port(), addr.AddrPort().Port(), log.NewLogger())
	handler := v4.NewHandler(holder, log.NewLogger(), v4.DefaultConfig())
	quierier := application.NewQuierier(holder)
	for _, flag := range fakecaps {
		quierier.AddCap(flag.Name, flag.Version)
	}
	logger.Info("crawling starting",
		log.NewTField("caps", fakecaps),
		log.NewStringField("name", holder.FakeName()),
		log.NewTField("targetedIdentities", ids))
	crawlApp := application.NewCrawlApp(application.Config{
		Port:      6061,
		Bootnodes: bootnodes,
		Handler:   handler,
		Conn:      socket,
		Metrics:   metrics,
		Keys:      ids,
		Qurier:    quierier,
		Store:     store,

		//broker
		BrokerEnabled:  broker_enabled,
		BrokerUser:     broker_username,
		BrokerPassword: broker_password,
		BrokerEndpoint: broker_endpoint,
	})
	crawlApp.Run(ctx.Context)
	return nil
}
