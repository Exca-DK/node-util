package scanner

import (
	"context"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/Exca-DK/node-util/scanner/log"
	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	cl  *redis.Client
	Log log.Logger
}

func (r *RedisClient) Start(addr string, password string) error {
	r.cl = redis.NewClient(&redis.Options{Addr: addr, Password: password, DB: 0})
	_, err := r.cl.Ping(context.Background()).Result()
	if err != nil {
		return err
	}
	go r.runStatus()
	return nil
}

func (r *RedisClient) runStatus() {
	ctx := context.Background()
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	for {
		_, err := r.cl.Ping(ctx).Result()
		if err != nil {
			r.Log.Warn("redis status failure", log.NewErrorField(err))
			os.Exit(1)
		}
	}
}

func (r *RedisClient) cacheScan(ctx context.Context, ip net.IP) bool {
	op := r.cl.SetNX(ctx, fmt.Sprintf("%s.%s", "scan.ip", ip.String()), time.Now().String(), 12*time.Hour)
	return op.Val()
}

func (r *RedisClient) deleteScan(ctx context.Context, ip net.IP) {
	r.cl.Del(ctx, fmt.Sprintf("%s.%s", "scan.ip", ip.String()))
}
