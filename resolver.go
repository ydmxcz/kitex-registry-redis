package redis

import (
	"context"
	"encoding/json"

	"github.com/cloudwego/kitex/pkg/discovery"
	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	redis "github.com/redis/go-redis/v9"
)

var _ discovery.Resolver = (*redisResolver)(nil)

type redisResolver struct {
	client *redis.Client
}

// NewRedisResolver creates a redis resolver
func NewRedisResolver(addr string, opts ...Option) discovery.Resolver {
	redisOpts := &redis.Options{Addr: addr}
	for _, opt := range opts {
		opt(redisOpts)
	}
	rdb := redis.NewClient(redisOpts)
	return &redisResolver{
		client: rdb,
	}
}

func (r *redisResolver) Target(ctx context.Context, target rpcinfo.EndpointInfo) (description string) {
	return target.ServiceName()
}

// Resolve returns a list of instances for the given description of a target.
func (r *redisResolver) Resolve(ctx context.Context, desc string) (result discovery.Result, err error) {
	rdb := r.client
	fvs := rdb.HGetAll(ctx, generateKey(desc, server)).Val()
	var its []discovery.Instance
	for f, v := range fvs {
		var ri registryInfo
		err := json.Unmarshal([]byte(v), &ri)
		if err != nil {
			klog.Warnf("HERTZ: fail to unmarshal with err: %v, ignore instance Addr: %v", err, f)
			continue
		}
		weight := ri.Weight
		if weight <= 0 {
			weight = defaultWeight
		}
		its = append(its, discovery.NewInstance(tcp, ri.Addr, weight, ri.Tags))
	}
	return discovery.Result{
		Cacheable: true,
		CacheKey:  desc,
		Instances: its,
	}, nil
}

// Diff computes the difference between two results.
// When `next` is cacheable, the Change should be cacheable, too. And the `Result` field's CacheKey in
// the return value should be set with the given cacheKey.
func (r *redisResolver) Diff(cacheKey string, prev, next discovery.Result) (change discovery.Change, ok bool) {
	return discovery.DefaultDiff(cacheKey, prev, next)
}

// Name returns the name of the resolver.
func (r *redisResolver) Name() string {
	return "redis"
}

// func (r *redisResolver) Target(_ context.Context, target *discovery.TargetInfo) string {
// 	return target.Host
// }

// func (r *redisResolver) Resolve(ctx context.Context, desc string) (discovery.Result, error) {
// 	rdb := r.client
// 	fvs := rdb.HGetAll(ctx, generateKey(desc, server)).Val()
// 	var its []discovery.Instance
// 	for f, v := range fvs {
// 		var ri registryInfo
// 		err := json.Unmarshal([]byte(v), &ri)
// 		if err != nil {
// 			hlog.Warnf("HERTZ: fail to unmarshal with err: %v, ignore instance Addr: %v", err, f)
// 			continue
// 		}
// 		weight := ri.Weight
// 		if weight <= 0 {
// 			weight = defaultWeight
// 		}
// 		its = append(its, discovery.NewInstance(tcp, ri.Addr, weight, ri.Tags))
// 	}
// 	return discovery.Result{
// 		CacheKey:  desc,
// 		Instances: its,
// 	}, nil
// }

// func (r *redisResolver) Name() string {
// 	return Redis
// }
