package scanner

import (
	"context"
	"net"
	"testing"
	"time"
)

func TestScanPortRange(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	openPorts := make([]int, 0)
	for i := 2000; i < 2600; i += 100 {
		go listen(ctx, i)
		openPorts = append(openPorts, i)
	}

	scanner := NewScannerController(1, 10, 10*time.Microsecond, nil)

	ports, err := scanner.Scan(ctx, net.IPv4(127, 0, 0, 1), 1, MaxTcpPort)
	if err != nil {
		t.Fatalf("scan failure. error: %v", err)
	}

	for _, openport := range openPorts {
		found := false
		for _, port := range ports {
			if port == uint16(openport) {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("missing open port: %v. got %v", openport, ports)
		}
	}
}

func listen(ctx context.Context, port int) error {
	listener, err := net.ListenTCP("tcp", &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: port})
	if err != nil {
		panic(err)
	}
	go func() {
		<-ctx.Done()
		listener.Close()
	}()
	for {
		_, err := listener.Accept()
		if err != nil {
			if net.ErrClosed == err {
				return nil
			}
			return err
		}
	}
}
