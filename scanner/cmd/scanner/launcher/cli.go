package launcher

import (
	"errors"
	"net"

	application "github.com/Exca-DK/node-util/scanner/app"
	"github.com/Exca-DK/node-util/scanner/log"
	"github.com/Exca-DK/node-util/scanner/scanner"
	"github.com/urfave/cli/v2"
)

var (
	EventScanCommand = &cli.Command{
		Name:        "EventScan",
		Description: "scans new node events",
		Action:      runEventScanner,
		Flags:       []cli.Flag{},
	}
	IpScanCommand = &cli.Command{
		Name:        "IpScan",
		Description: "scans ips",
		Action:      runIpScanner,
		Flags:       []cli.Flag{},
	}
)

func init() {
	baseFlags := make([]cli.Flag, 0)
	baseFlags = append(baseFlags, application.Verbose)
	baseFlags = append(baseFlags, application.IntervalFlag)
	baseFlags = append(baseFlags, application.BurstFlag)
	baseFlags = append(baseFlags, application.IpsFlag)
	baseFlags = append(baseFlags, application.ScanPortStart)
	baseFlags = append(baseFlags, application.ScanPortEnd)

	EventScanCommand.Flags = append(EventScanCommand.Flags, application.MetricsFlag)
	EventScanCommand.Flags = append(EventScanCommand.Flags, application.BrokerFlags...)
	EventScanCommand.Flags = append(EventScanCommand.Flags, application.CacheFlags...)
	EventScanCommand.Flags = append(EventScanCommand.Flags, application.DbFlags...)
	EventScanCommand.Flags = append(EventScanCommand.Flags, baseFlags...)

	IpScanCommand.Flags = append(IpScanCommand.Flags, baseFlags...)
}

func runEventScanner(ctx *cli.Context) error {
	cfg := application.DefaultConfig()
	cfg, err := application.Parse(ctx, cfg)
	if err != nil {
		return err
	}
	if err := cfg.Validate(); err != nil {
		return err
	}
	app := application.NewScanApp(cfg)
	if err := app.Setup(); err != nil {
		return err
	}
	err = app.Run(ctx.Context)
	app.Stop()
	return err
}

func runIpScanner(ctx *cli.Context) error {
	logger := log.Root()
	cfg := application.DefaultConfig()
	cfg, err := application.Parse(ctx, cfg)
	if err != nil {
		return err
	}

	if len(cfg.Ips) == 0 {
		return errors.New("running scanner without ip is useless")
	}

	scanner := scanner.NewScannerController(cfg.Threads, cfg.Burst, cfg.Interval, nil)
	chErrs := make(chan error, len(cfg.Ips))
	for _, ip := range cfg.Ips {
		go func(ip net.IP) {
			ports, err := scanner.Scan(ctx.Context, ip, uint16(cfg.PortStart), uint16(cfg.PortEnd))
			if err == nil {
				logger.Info("scanned ip", log.NewStringField("ip", ip.String()), log.NewTField("ips", ports))
			} else {
				logger.Warn("scan failure", log.NewStringField("ip", ip.String()), log.NewErrorField(err))
			}
			chErrs <- err
		}(ip)
	}

	var chError error
	for i := 0; i < len(cfg.Ips); i++ {
		select {
		case <-ctx.Context.Done():
			return ctx.Err()
		case err := <-chErrs:
			if err != nil {
				chError = err
			}
		}
	}
	return chError
}
