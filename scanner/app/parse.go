package app

import (
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/Exca-DK/node-util/scanner/log"
	"github.com/urfave/cli/v2"
)

func Parse(ctx *cli.Context, cfg *Config) (*Config, error) {
	parseLogging(ctx)
	ParseMetrics(ctx, cfg)
	ParseInterval(ctx, cfg)
	ParseIps(ctx, cfg)
	if err := ParseBroker(ctx, cfg); err != nil {
		return cfg, err
	}
	if err := ParseDb(ctx, cfg); err != nil {
		return cfg, err
	}
	if err := ParseCache(ctx, cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}

func parseLogging(ctx *cli.Context) {
	if ctx.IsSet(Verbose.Name) {
		verbose := ctx.Bool(Verbose.Name)
		if verbose {
			log.SetLoggingLvl(log.DEBUG)
		}
	}
}

func ParseMetrics(ctx *cli.Context, cfg *Config) {
	if ctx.IsSet(MetricsFlag.Name) {
		cfg.Metrics = ctx.Bool(MetricsFlag.Name)
	}
}

func ParseBroker(ctx *cli.Context, cfg *Config) error {
	if !ctx.IsSet(brokerEnabled.Name) {
		return nil
	}
	if !ctx.IsSet(brokerUser.Name) {
		return errors.New("broker enabled without username")
	}
	if !ctx.IsSet(brokerPassword.Name) {
		return errors.New("broker enabled without password")
	}
	if !ctx.IsSet(brokerEndpoint.Name) {
		return errors.New("broker enabled without endpoint")
	}
	cfg.BrokerEnabled = ctx.Bool(brokerEnabled.Name)
	cfg.BrokerUser = ctx.String(brokerUser.Name)
	cfg.BrokerPassword = ctx.String(brokerPassword.Name)
	cfg.BrokerEndpoint = ctx.String(brokerEndpoint.Name)
	return nil
}

func ParseDb(ctx *cli.Context, cfg *Config) error {
	if !ctx.IsSet(dbEnabled.Name) {
		return nil
	}
	if !ctx.IsSet(dbSource.Name) {
		return fmt.Errorf("db enabled without source: %v", dbSource.Name)
	}
	if !ctx.IsSet(dbMigrationUrl.Name) {
		return fmt.Errorf("db enabled without migration: %v", dbMigrationUrl.Name)
	}

	cfg.StoreEnabled = ctx.Bool(dbEnabled.Name)
	cfg.StoreSource = ctx.String(dbSource.Name)
	cfg.StoreMigration = ctx.String(dbMigrationUrl.Name)
	return nil
}

func ParseCache(ctx *cli.Context, cfg *Config) error {
	if !ctx.IsSet(cacheEnabled.Name) {
		return nil
	}

	if !ctx.IsSet(cacheEndpoint.Name) {
		return fmt.Errorf("flag missing: %v", cacheEndpoint.Name)
	}
	if !ctx.IsSet(cachePassword.Name) {
		return fmt.Errorf("flag missing: %v", cacheEndpoint.Name)
	}
	cfg.CacheEnabled = ctx.Bool(cacheEnabled.Name)
	cfg.CacheEndpoint = ctx.String(cacheEndpoint.Name)
	cfg.CachePassword = ctx.String(cachePassword.Name)
	return nil
}

func ParseIps(ctx *cli.Context, cfg *Config) error {
	if !ctx.IsSet(IpsFlag.Name) {
		return nil
	}
	rawIps := ctx.StringSlice(IpsFlag.Name)
	for _, raw := range rawIps {
		ip := net.ParseIP(raw)
		if ip == nil {
			return fmt.Errorf("invalid ip. ip: %v", raw)
		}
		cfg.Ips = append(cfg.Ips, ip)
	}
	if len(rawIps) == 0 {
		return errors.New("ips set without ip")
	}
	return nil
}

func ParseInterval(ctx *cli.Context, cfg *Config) {
	if !ctx.IsSet(IntervalFlag.Name) {
		return
	}
	cfg.Interval = time.Millisecond * time.Duration(ctx.Int(IntervalFlag.Name))
}

func ParsePorts(ctx *cli.Context, cfg *Config) {
	if ctx.IsSet(ScanPortStart.Name) {
		cfg.PortStart = int(ctx.Uint(ScanPortStart.Name))
	}
	if ctx.IsSet(ScanPortStart.Name) {
		cfg.PortEnd = int(ctx.Uint(ScanPortEnd.Name))
	}
}
