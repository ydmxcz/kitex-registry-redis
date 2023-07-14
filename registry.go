package redis

import (
	"context"
	"sync"

	"github.com/cloudwego/kitex/pkg/registry"
	redis "github.com/redis/go-redis/v9"
)

var _ registry.Registry = (*redisRegistry)(nil)

type redisRegistry struct {
	client *redis.Client
	rctx   *registryContext
	mu     sync.Mutex
	wg     sync.WaitGroup
}

type registryContext struct {
	ctx    context.Context
	cancel context.CancelFunc
}

// NewRedisRegistry creates a redis registry
func NewRedisRegistry(addr string, opts ...Option) registry.Registry {
	redisOpts := &redis.Options{
		Addr:     addr,
		Password: "",
		DB:       0,
	}
	for _, opt := range opts {
		opt(redisOpts)
	}
	rdb := redis.NewClient(redisOpts)
	return &redisRegistry{
		client: rdb,
	}
}

func (r *redisRegistry) Register(info *registry.Info) error {
	if err := validateRegistryInfo(info); err != nil {
		return err
	}
	rctx := registryContext{}
	rctx.ctx, rctx.cancel = context.WithCancel(context.Background())
	m := newMentor()
	r.wg.Add(1)
	go m.subscribe(rctx.ctx, info, r)
	r.wg.Wait()
	rdb := r.client
	hash, err := prepareRegistryHash(info)
	if err != nil {
		return err
	}
	r.mu.Lock()
	r.rctx = &rctx
	rdb.HSet(rctx.ctx, hash.key, hash.field, hash.value)
	rdb.Expire(rctx.ctx, hash.key, defaultExpireTime)
	rdb.Publish(rctx.ctx, hash.key, generateMsg(register, info.ServiceName, info.Addr.String()))
	r.mu.Unlock()
	go m.monitorTTL(rctx.ctx, hash, info, r)
	go keepAlive(rctx.ctx, hash, r)
	return nil
}

func (r *redisRegistry) Deregister(info *registry.Info) error {
	if err := validateRegistryInfo(info); err != nil {
		return err
	}
	rctx := r.rctx
	rdb := r.client
	hash, err := prepareRegistryHash(info)
	if err != nil {
		return err
	}
	r.mu.Lock()
	rdb.HDel(rctx.ctx, hash.key, hash.field)
	rdb.Publish(rctx.ctx, hash.key, generateMsg(deregister, info.ServiceName, info.Addr.String()))
	rctx.cancel()
	r.mu.Unlock()
	return nil
}
