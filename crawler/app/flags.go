package app

import "github.com/urfave/cli/v2"

var (
	NodesFlag = &cli.StringFlag{
		Name:  "nodes",
		Usage: "remote enodes",
	}
	TimeOutInMinutesFlag = &cli.Uint64Flag{
		Name:  "timeout",
		Usage: "specify timeout in minutes",
	}
	ConcurrencyFactorFlag = &cli.Uint64Flag{
		Name:  "concurrency",
		Usage: "specify the ratio for concurrency",
	}
	DirFlag = &cli.StringFlag{
		Name:  "dir",
		Usage: "specify the directory",
	}
	MetricsFlag = &cli.BoolFlag{
		Name:  "metrics",
		Usage: "enable metrics",
	}
	IdentitiesFlag = &cli.StringSliceFlag{
		Name:  "identities",
		Usage: "nodes without provided ids will be ignored",
	}
	DiscPort = &cli.IntFlag{
		Name:  "discport",
		Usage: "port for discovery",
	}
	FakeName = &cli.StringFlag{
		Name:  "fakename",
		Usage: "name used for faking",
	}
	FakeCaps = &cli.StringSliceFlag{
		Name:  "fakecaps",
		Usage: "caps to be used for rlpx conns",
	}

	//broker flags
	BrokerEnabled  = &cli.BoolFlag{Name: "enable_broker"}
	BrokerEndpoint = &cli.StringFlag{Name: "broker_endpoint"}
	BrokerUser     = &cli.StringFlag{Name: "broker_user"}
	BrokerPassword = &cli.StringFlag{Name: "broker_password"}

	//db flags
	DbSource       = &cli.StringFlag{Name: "db_source"}
	DbMigrationUrl = &cli.StringFlag{Name: "db_migration_url"}
)
