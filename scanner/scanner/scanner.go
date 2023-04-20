package scanner

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/Exca-DK/node-util/scanner/metrics"

	"github.com/Exca-DK/node-util/scanner/log"
	"golang.org/x/time/rate"
)

const (
	MaxTcpPort  = 65_535
	logInterval = 15 * time.Second
	dialTimeout = 1 * time.Second / 2
)

var (
	ErrScannerStopped = errors.New("scanner stopped")
)

type scanItem struct {
	Ip        net.IP
	PortStart uint16
	PortEnd   uint16
	callback  chan *scannedItem
}

type scannedItem struct {
	Src        scanItem
	Open       []uint16
	startedAt  int64
	finishedAt int64
}

func (item scannedItem) StartedAt() time.Time {
	return time.UnixMilli(item.startedAt)
}

func (item scannedItem) FinishedAt() time.Time {
	return time.UnixMilli(item.finishedAt)
}

func NewScannerController(threads int, maxBurst int, maxInterval time.Duration, logger log.Logger) *ScannerController {
	if logger == nil {
		logger = log.NewLogger()
	}
	logger.Info("scan controller info",
		log.NewTField("threads", threads),
		log.NewTField("maxBurst", maxBurst),
		log.NewTField("maxInterval", maxInterval))
	controller := ScannerController{
		in:           make(chan scanItem),
		maxBurst:     maxBurst,
		maxFrequency: maxInterval,
		threads:      threads,
		sig:          make(chan struct{}),
		log:          logger,
		dialTimeout:  dialTimeout,
		recorder:     newRecorder(metrics.Factory),
	}
	for i := 0; i < threads; i++ {
		go controller.runScanner()
	}
	go func() {
		<-controller.sig
		close(controller.done)
	}()
	return &controller
}

type ScannerController struct {
	log          log.Logger
	dialTimeout  time.Duration
	maxBurst     int
	maxFrequency time.Duration
	threads      int

	recorder Recorder

	in   chan scanItem
	done chan struct{}
	sig  chan struct{}
}

func (scanner *ScannerController) Scan(ctx context.Context, ip net.IP, start uint16, end uint16) ([]uint16, error) {
	item := scanItem{
		Ip:        ip,
		PortStart: start,
		PortEnd:   end,
		callback:  make(chan *scannedItem, 1),
	}

	select {
	case scanner.in <- item:
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-scanner.done:
		return nil, ErrScannerStopped
	}

	select {
	case result, ok := <-item.callback:
		if !ok {
			return nil, fmt.Errorf("scan failure. ip:%v err:%v", ip, errors.New("canceled"))
		}
		return result.Open, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-scanner.done:
		return nil, ErrScannerStopped
	}
}

func (scanner *ScannerController) ScanNoWait(ctx context.Context, ip net.IP, start uint16, end uint16, callback func(*scannedItem)) error {
	ch := make(chan *scannedItem, 1)
	go func() {
		select {
		case <-ctx.Done():
			return
		case item := <-ch:
			callback(item)
		}
	}()
	item := scanItem{
		Ip:        ip,
		PortStart: start,
		PortEnd:   end,
		callback:  ch,
	}

	select {
	case scanner.in <- item:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	case <-scanner.done:
		return ErrScannerStopped
	}
}

func (controller *ScannerController) runScanner() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	scanner := NewScanner(ctx, controller.maxBurst*controller.threads, controller.maxFrequency/time.Duration(controller.threads)+1, controller.dialTimeout, controller.log)
	scanner.recorder = controller.recorder
	for {
		select {
		case <-controller.done:
			return
		case item := <-controller.in:
			controller.log.Debug("running ip scan", log.NewStringField("ip", item.Ip.String()))
			result, index := scanner.ScanItem(item)
			controller.log.Debug("ip scan finished",
				log.NewStringField("ip", item.Ip.String()),
				log.NewStringField("elapsed", time.Since(result.StartedAt()).String()),
				log.NewTField("full", index == result.Src.PortEnd))
			controller.recorder.RecordScan(time.Since(result.StartedAt()))
			// ctx has been canceled
			if index != result.Src.PortEnd {
				close(item.callback)
				return
			}
			item.callback <- &result
		}
	}
}

func (controller *ScannerController) Stop() {
	select {
	case controller.sig <- struct{}{}:
	case <-controller.done:
	}
}

func NewScanner(ctx context.Context, burst int, interval time.Duration, timeout time.Duration, log log.Logger) scanner {
	limiter := rate.NewLimiter(rate.Every(interval), burst)
	scanner := scanner{limiter: limiter, ctx: ctx, timeout: timeout, log: log}
	return scanner
}

type scanner struct {
	log      log.Logger
	limiter  *rate.Limiter
	ctx      context.Context
	timeout  time.Duration
	recorder Recorder
}

// scans port range of the ip untill the context has been canceled
// returns scanned result and port at which the job has ended
func (scanner *scanner) ScanItem(item scanItem) (scannedItem, uint16) {
	tcpAddr := &net.TCPAddr{IP: item.Ip, Port: 0}
	result := scannedItem{Src: item, Open: []uint16{}, startedAt: time.Now().UnixMilli()}
	if item.PortEnd > MaxTcpPort {
		item.PortEnd = MaxTcpPort
	}
	defer func(ts time.Time) { result.finishedAt = ts.UnixMilli() }(time.Now())
	ts := time.Now()
	for i := int(item.PortStart); i <= int(item.PortEnd); i++ {
		if err := scanner.limiter.Wait(scanner.ctx); err != nil {
			scanner.log.Warn("encountered limiting error when scanning",
				log.NewStringField("ip", item.Ip.String()),
				log.NewTField("at", i),
				log.NewTField("open_so_far", result.Open),
				log.NewErrorField(err))
			return result, uint16(i)
		}
		open := scanner.IsOpen(tcpAddr, i)
		if open {
			result.Open = append(result.Open, uint16(i))
		}
		if scanner.recorder != nil {
			scanner.recorder.RecordPortScan(open)
		}
		if time.Since(ts) > logInterval {
			ts = time.Now()
			scanner.log.Debug("scanning progress debug info",
				log.NewStringField("ip", item.Ip.String()),
				log.NewTField("at", i),
				log.NewTField("open_so_far", result.Open))
		}
	}
	return result, result.Src.PortEnd
}

func (h *scanner) IsOpen(addr *net.TCPAddr, port int) bool {
	addr.Port = port
	conn, err := net.DialTimeout("tcp", addr.String(), h.timeout)
	if err != nil {
		return false
	}
	defer conn.Close()

	return true
}
