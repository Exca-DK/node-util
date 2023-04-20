package v4

import (
	"fmt"
	"net"
	"time"

	application "github.com/Exca-DK/node-util/crawler/app"
	v4 "github.com/Exca-DK/node-util/crawler/net/discovery/v4"
	lendr "github.com/Exca-DK/node-util/crawler/net/enr"
	"github.com/ethereum/go-ethereum/p2p/enode"

	"github.com/Exca-DK/node-util/crawler/log"
	"github.com/Exca-DK/node-util/crawler/net/discovery"
	"github.com/urfave/cli/v2"
)

func discV4Ping(ctx *cli.Context) error {
	logger := log.NewLogger()
	nodes, err := application.ParseNodes(ctx)
	if err != nil {
		return err
	}

	handler, listener, err := makeHandler()
	if err != nil {
		return err
	}
	defer listener.Stop()
	for _, node := range nodes {
		val, err := handler.Ping(node)
		if err != nil {
			return err
		}
		logger.Info("recv pong", log.NewStringField("from", node.String()), log.NewTField("value", val))
	}
	return nil
}

func discV4AskRandom(ctx *cli.Context) error {
	logger := log.NewLogger()

	nodes, err := application.ParseNodes(ctx)
	if err != nil {
		return err
	}

	handler, listener, err := makeHandler()
	if err != nil {
		return err
	}
	defer listener.Stop()

	for _, node := range nodes {
		nodes, err := handler.AskForRandomNodes(node)
		if err != nil {
			return err
		}
		logger.Info("recv nodes", log.NewStringField("from", node.String()), log.NewTField("nodes", nodes))
	}
	return nil
}

func discV4AskEnr(ctx *cli.Context) error {
	logger := log.NewLogger()

	nodes, err := application.ParseNodes(ctx)
	if err != nil {
		return err
	}

	handler, listener, err := makeHandler()
	if err != nil {
		return err
	}
	defer listener.Stop()

	for _, node := range nodes {
		enr, err := handler.RequestENR(node)
		if err != nil {
			return err
		}
		logger.Info("recv enr", log.NewStringField("from", node.String()), log.NewTField("enr", lendr.RetrieveFieldNames(enr.Record())))
	}
	return nil
}

func discV4PeekDB(ctx *cli.Context) error {
	logger := log.NewLogger()

	nodes, err := application.ParseNodes(ctx)
	if err != nil {
		return err
	}

	factor := application.ParseConcurrency(ctx)
	raceCtx, cancel := application.ParseTimeout(ctx)
	defer cancel()

	handler, listener, err := makeHandler()
	if err != nil {
		return err
	}
	defer listener.Stop()

	type nodeScanResult struct {
		scanned []*enode.Node
		target  *enode.Node
	}

	chScan := make(chan nodeScanResult, len(nodes))
	for _, node := range nodes {
		go func(node *enode.Node) {
			ticker := time.NewTicker(30 * time.Second)
			defer ticker.Stop()

			cache := make(map[enode.ID]*enode.Node, 1024)
			var nodesAmount, duplicates, queries uint16
			var exhaused uint8
			resultCh := make(chan []*enode.Node, factor)
			for i := 0; i < factor; i++ {
				go func(i int) {
					time.Sleep(time.Duration(i) * time.Second)
					errors := 0
					for {
						if errors > 10 {
							logger.Warn("AskForRandomNodes failure")
							return
						}
						result, err := handler.AskForRandomNodes(node)
						if err != nil {
							errors++
							time.Sleep(30 * time.Second)
						} else {
							errors = 0
						}
						resultCh <- result
					}
				}(i)
			}
		OUTER:
			for {
				select {
				case <-ticker.C:
					logger.Info("stats info",
						log.NewTField("queries", queries),
						log.NewTField("duplicates", duplicates),
						log.NewTField("new", nodesAmount),
						log.NewTField("total", len(cache)),
						log.NewTField("exhausted", exhaused),
						log.NewStringField("node", node.String()))
					if nodesAmount == duplicates {
						exhaused++
					}
					if exhaused > 3 {
						break OUTER
					}
					nodesAmount, queries, duplicates = 0, 0, 0
				case <-raceCtx.Done():
					break OUTER
				case <-ctx.Done():
					break OUTER
				case result := <-resultCh:
					queries++
					for i := range result {
						_node := result[i]
						_, ok := cache[_node.ID()]
						if ok {
							duplicates++
							continue
						}
						cache[_node.ID()] = _node
					}
					nodesAmount += uint16(len(result))
				}
			}

			scanned := make([]*enode.Node, 0, len(cache))
			for id := range cache {
				scanned = append(scanned, cache[id])
			}

			chScan <- nodeScanResult{scanned: scanned, target: node}
		}(node)
	}

	scanSet := make(scanSet)
	for i := 0; i < len(nodes); i++ {
		scan := <-chScan
		logger.Info("node scan info", log.NewTField("nodes", len(scan.scanned)), log.NewStringField("node", scan.target.String()))
		scanSet[scan.target.String()] = scan.scanned
	}

	path := fmt.Sprintf("%v/scan.json", application.ParseDir(ctx))
	logger.Info(fmt.Sprintf("dumping scan at %v", path))
	writeScanJSON(path, scanSet)
	return nil
}

func makeHandler() (discovery.IHandler, discovery.Listener, error) {
	db, err := enode.OpenDB("")
	if err != nil {
		return nil, nil, err
	}
	socket, err := net.ListenPacket("udp4", "0.0.0.0:")
	if err != nil {
		return nil, nil, err
	}

	addr := socket.LocalAddr().(*net.UDPAddr)
	holder := discovery.NewRandomIdentity("", addr.IP, addr.AddrPort().Port(), addr.AddrPort().Port(), log.NewLogger())

	return application.MakeHandler(socket, make([]*enode.Node, 0), db, v4.NewHandler(holder, log.NewLogger(), v4.DefaultConfig()))
}
