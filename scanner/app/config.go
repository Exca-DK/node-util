package app

import (
	"errors"
	"net"
	"runtime"
	"time"

	"github.com/Exca-DK/node-util/scanner/log"
	"github.com/Exca-DK/node-util/scanner/scanner"
)

type Config struct {
	Ips       []net.IP
	Threads   int
	Burst     int
	Interval  time.Duration
	PortStart int
	PortEnd   int

	Metrics bool

	StoreEnabled   bool
	StoreMigration string
	StoreSource    string

	CacheEnabled  bool
	CacheEndpoint string
	CachePassword string

	BrokerEnabled  bool
	BrokerEndpoint string
	BrokerUser     string
	BrokerPassword string
}

func DefaultConfig() *Config {
	cfg := &Config{}
	cfg.Threads = runtime.GOMAXPROCS(0)
	cfg.Burst = 10
	cfg.Interval = 1 * time.Second / 10
	cfg.PortEnd = scanner.MaxTcpPort
	return cfg
}

func (conf *Config) Validate() error {
	if len(conf.Ips) == 0 && !conf.BrokerEnabled {
		return errors.New("running scanner without broker or ips is useless")
	}
	logger := log.Root()
	if conf.Burst <= 0 {
		logger.Warn("Burst below 0. reverting to 1")
		conf.Burst = 1
	}
	if conf.Interval < 1*time.Millisecond {
		logger.Warn("Burst below 10ms. clamping...")
		conf.Interval = 1 * time.Millisecond
	}
	if conf.Threads <= 0 {
		logger.Warn("threads below 1. reverting to 1")
		conf.Threads = 1
	}
	if conf.PortStart < 0 {
		logger.Warn("Port start below 0. reverting to 0")
		conf.PortStart = 0
	}
	if conf.PortEnd <= 0 {
		logger.Warn("Port end above maximum. reverting to max")
		conf.PortStart = scanner.MaxTcpPort
	}
	return nil
}
