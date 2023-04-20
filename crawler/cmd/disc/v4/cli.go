package v4

import (
	application "github.com/Exca-DK/node-util/crawler/app"
	"github.com/urfave/cli/v2"
)

var (
	Discv4Command = &cli.Command{
		Name:  "discv4",
		Usage: "Node Discovery v4 tools",
		Subcommands: []*cli.Command{
			discv4PingCommand,
			discv4GetRandomNodes,
			discv4GetEnr,
			discv4PeekDb,
		},
	}
	discv4PingCommand = &cli.Command{
		Name:      "ping",
		Usage:     "Sends ping to nodes",
		Action:    discV4Ping,
		ArgsUsage: "<nodes>",
		Flags:     []cli.Flag{application.NodesFlag},
	}
	discv4GetRandomNodes = &cli.Command{
		Name:      "random",
		Usage:     "Get random nodes from nodes",
		Action:    discV4AskRandom,
		ArgsUsage: "<nodes>",
		Flags:     []cli.Flag{application.NodesFlag},
	}
	discv4GetEnr = &cli.Command{
		Name:      "enr",
		Usage:     "Get enr nodes from nodes",
		Action:    discV4AskEnr,
		ArgsUsage: "<nodes>",
		Flags:     []cli.Flag{application.NodesFlag},
	}
	discv4PeekDb = &cli.Command{
		Name:      "scan",
		Usage:     "Try to scan whole enode.db of targeted peer",
		Action:    discV4PeekDB,
		ArgsUsage: "<nodes>",
		Flags:     []cli.Flag{application.NodesFlag, application.TimeOutInMinutesFlag, application.ConcurrencyFactorFlag, application.DirFlag},
	}
)
