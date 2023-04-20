package scanner

import (
	"context"
	"net"

	"github.com/Exca-DK/node-util/scanner/log"

	"github.com/Exca-DK/node-util/scanner/broker"
	store "github.com/Exca-DK/node-util/service/db"
	db "github.com/Exca-DK/node-util/service/db/sqlc"
	"github.com/ethereum/go-ethereum/p2p/enode"
)

func NewNodeScanner(controller *ScannerController, log log.Logger, store store.Store, broker broker.Broker, cache *RedisClient) nodeScanner {
	return nodeScanner{
		controller: controller,
		log:        log,
		store:      store,
		broker:     broker,
		cache:      cache,
	}
}

type nodeScanner struct {
	controller *ScannerController
	log        log.Logger
	store      store.Store
	broker     broker.Broker
	cache      *RedisClient
}

func (scanner *nodeScanner) Start(ctx context.Context) error {
	ch, sub, err := scanner.broker.SubscribeNewNodes(ctx)
	if err != nil {
		return err
	}
	defer sub.Unsubscribe()
	callback := func(item *scannedItem, node string, ip net.IP) {
		if item == nil {
			scanner.log.Info("scan crashed", log.NewStringField("node", node))
			scanner.cache.deleteScan(context.Background(), ip)
			return
		}
		scanner.store.CreateScanRecord(ctx, db.CreateScanRecordParams{
			Node:  node,
			Ip:    item.Src.Ip.String(),
			Ports: convertArray(item.Open),
		})
		scanner.log.Info("scanned node",
			log.NewStringField("node", node),
			log.NewTField("open", item.Open),
			log.NewStringField("duration", item.FinishedAt().Sub(item.StartedAt()).String()))
	}
	scanner.log.Info("waiting for nodes")
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case node := <-ch:
			scanner.log.Info("got node", log.NewTField("nmode", node))
			enr, err := enode.ParseV4(node.ENR)
			if err != nil {
				scanner.log.Warn("invalid enr received", log.NewStringField("enr", node.ENR), log.NewErrorField(err))
				continue
			}
			scanner.log.Debug("recv node to scan", log.NewStringField("enr", node.ENR))
			ip := enr.IP()
			if !ip.IsGlobalUnicast() {
				scanner.log.Warn("invalid ip received", log.NewStringField("enr", node.ENR), log.NewTField("ip", ip.String()))
				continue
			}
			if !scanner.cache.cacheScan(ctx, ip) {
				scanner.log.Info("cache hit. ignoring", log.NewStringField("enr", node.ENR), log.NewStringField("ip", ip.String()))
				continue
			}
			err = scanner.controller.ScanNoWait(ctx, ip, 22, MaxTcpPort, func(item *scannedItem) { callback(item, node.ENR, ip) })
			if err != nil {
				scanner.log.Warn("scan failure", log.NewStringField("enr", node.ENR), log.NewErrorField(err))
				scanner.cache.deleteScan(ctx, ip)
				continue
			}
		}
	}
}

func convertArray(uints []uint16) []int32 {
	arr := make([]int32, len(uints))
	for i, v := range uints {
		arr[i] = int32(v)
	}
	return arr
}
