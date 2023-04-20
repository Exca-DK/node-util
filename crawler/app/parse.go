package app

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Exca-DK/node-util/crawler/net"
	store "github.com/Exca-DK/node-util/service/db"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/urfave/cli/v2"
)

// parseNodes parses node records from provider context
func ParseNodes(ctx *cli.Context) ([]*enode.Node, error) {
	var s []string
	if ctx.IsSet(NodesFlag.Name) {
		input := ctx.String(NodesFlag.Name)
		if input == "" {
			return nil, nil
		}
		s = strings.Split(input, ",")
	}
	nodes := make([]*enode.Node, 0, len(s))
	for _, record := range s {
		node, err := ParseNode(record)
		if err != nil {
			return nil, fmt.Errorf("invalid bootstrap node: %v", err)
		}
		nodes = append(nodes, node)
	}
	return nodes, nil
}

func ParseNode(source string) (*enode.Node, error) {
	if strings.HasPrefix(source, "enode://") {
		return enode.ParseV4(source)
	}
	return nil, errors.New("unsupported node type")
}

func ParseMetrics(ctx *cli.Context) bool {
	if ctx.IsSet(MetricsFlag.Name) {
		return ctx.Bool(MetricsFlag.Name)
	}
	return false
}

func ParseIdentities(ctx *cli.Context) []string {
	var s []string
	if ctx.IsSet(IdentitiesFlag.Name) {
		input := ctx.StringSlice(IdentitiesFlag.Name)
		s = append(s, input...)
	}

	if len(s) == 0 {
		s = append(s, "*")
	}

	return s
}

func ParseDiscPort(ctx *cli.Context) int {
	var port int
	if ctx.IsSet(DiscPort.Name) {
		port = ctx.Int(DiscPort.Name)
	}
	if port < 1000 {
		return 10001
	}
	return port
}

func ParseTimeout(ctx *cli.Context) (context.Context, func()) {
	if !ctx.IsSet(TimeOutInMinutesFlag.Name) {
		return context.Background(), func() {}
	}
	input := ctx.Uint64(TimeOutInMinutesFlag.Name)
	if input == 0 {
		return context.Background(), func() {}
	}

	return context.WithTimeout(context.Background(), time.Duration(input)*time.Minute)
}

func ParseConcurrency(ctx *cli.Context) int {
	if !ctx.IsSet(ConcurrencyFactorFlag.Name) {
		return 8
	}
	factor := ctx.Uint64(ConcurrencyFactorFlag.Name)
	return int(factor)
}

const defaultDirectory = "."

func ParseDir(ctx *cli.Context) string {
	if !ctx.IsSet(DirFlag.Name) {
		return defaultDirectory
	}
	input := ctx.String(DirFlag.Name)
	if input == "" {
		return defaultDirectory
	}
	return input
}

func ParseFakeName(ctx *cli.Context) string {
	if ctx.IsSet(FakeName.Name) {
		return ctx.String(FakeName.Name)

	}
	return ""
}

func ParseFakeCaps(ctx *cli.Context) ([]p2p.Cap, error) {
	var rawflags []string
	if ctx.IsSet(FakeCaps.Name) {
		rawflags = ctx.StringSlice(FakeCaps.Name)
	}
	flags := make([]p2p.Cap, len(rawflags))
	for i, flag := range rawflags {
		sflag := strings.Split(flag, ":")
		if len(sflag) != 2 {
			return nil, fmt.Errorf("invalid cap provided. got %v, wants NAME:VALUE", flag)
		}
		value, err := strconv.Atoi(sflag[1])
		if err != nil {
			return nil, err
		}
		flags[i] = p2p.Cap{
			Name:    sflag[0],
			Version: uint(value),
		}
	}
	if len(flags) == 0 {
		flags = append(flags, net.DefaultQuieries...)
	}
	return flags, nil
}

func ParseBroker(ctx *cli.Context) (username string, password string, endpoint string, enabled bool, err error) {
	if !ctx.IsSet(BrokerEnabled.Name) {
		enabled = false
		return
	}
	enabled = true

	if !ctx.IsSet(BrokerUser.Name) {
		err = errors.New("broker enabled without username")
		return
	}
	if !ctx.IsSet(BrokerPassword.Name) {
		err = errors.New("broker enabled without password")
		return
	}
	if !ctx.IsSet(BrokerEndpoint.Name) {
		err = errors.New("broker enabled without endpoint")
		return
	}

	username = ctx.String(BrokerUser.Name)
	password = ctx.String(BrokerPassword.Name)
	endpoint = ctx.String(BrokerEndpoint.Name)
	return
}

func ParseDb(ctx *cli.Context) (store.Store, error) {
	if !ctx.IsSet(DbSource.Name) {
		return nil, fmt.Errorf("flag missing: %v", DbSource.Name)
	}

	if !ctx.IsSet(DbMigrationUrl.Name) {
		return nil, fmt.Errorf("flag missing: %v", DbMigrationUrl.Name)
	}

	var source, migration string
	source = ctx.String(DbSource.Name)
	migration = ctx.String(DbMigrationUrl.Name)
	return store.NewStore(store.Config{
		Source:        source,
		MigrationPath: migration,
	})
}
