package app

import "github.com/urfave/cli/v2"

var (
	MetricsFlag = &cli.BoolFlag{Name: "metrics", Usage: "enable metrics"}
	//broker flags
	BrokerFlags    = []cli.Flag{brokerEnabled, brokerEndpoint, brokerUser, brokerPassword}
	brokerEnabled  = &cli.BoolFlag{Name: "broker_enabled"}
	brokerEndpoint = &cli.StringFlag{Name: "broker_endpoint"}
	brokerUser     = &cli.StringFlag{Name: "broker_user"}
	brokerPassword = &cli.StringFlag{Name: "broker_password"}

	DbFlags        = []cli.Flag{dbEnabled, dbSource, dbMigrationUrl}
	dbEnabled      = &cli.BoolFlag{Name: "db_enabled"}
	dbSource       = &cli.StringFlag{Name: "db_source"}
	dbMigrationUrl = &cli.StringFlag{Name: "db_migration_url"}

	CacheFlags    = []cli.Flag{cacheEnabled, cacheEndpoint, cachePassword}
	cacheEnabled  = &cli.BoolFlag{Name: "cache_enabled"}
	cacheEndpoint = &cli.StringFlag{Name: "cache_endpoint"}
	cachePassword = &cli.StringFlag{Name: "cache_password"}

	IpsFlag       = &cli.StringSliceFlag{Name: "ips", Usage: "scans provided ips"}
	BurstFlag     = &cli.StringFlag{Name: "burst", Usage: "maximum amount of packets at once"}
	IntervalFlag  = &cli.UintFlag{Name: "interval_in_ms", Usage: "can't scan faster than that"}
	ScanPortStart = &cli.UintFlag{Name: "port_start", Usage: "which port to use as starting point"}
	ScanPortEnd   = &cli.UintFlag{Name: "port_end", Usage: "which port to use as ending point"}
	Verbose       = &cli.BoolFlag{Name: "verbose", Usage: "use detailed logging"}
)
