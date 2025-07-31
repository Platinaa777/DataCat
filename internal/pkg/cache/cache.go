package cache

import (
	"context"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

type Cache interface {
	Get(context.Context, string) ([]byte, bool)
	Set(context.Context, string, []byte, time.Duration)
}
type entry struct {
	data    []byte
	expires time.Time
}
type Memory struct {
	mu       sync.Mutex
	entries  map[string]entry
	capacity int
}

func NewMemory(capacity int) *Memory { return &Memory{entries: map[string]entry{}, capacity: capacity} }
func (m *Memory) Get(_ context.Context, key string) ([]byte, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	v, ok := m.entries[key]
	if !ok || time.Now().After(v.expires) {
		delete(m.entries, key)
		return nil, false
	}
	return append([]byte(nil), v.data...), true
}
func (m *Memory) Set(_ context.Context, key string, data []byte, ttl time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.entries) >= m.capacity {
		var oldest string
		var expires time.Time
		for k, v := range m.entries {
			if oldest == "" || v.expires.Before(expires) {
				oldest = k
				expires = v.expires
			}
		}
		delete(m.entries, oldest)
	}
	m.entries[key] = entry{append([]byte(nil), data...), time.Now().Add(ttl)}
}

type Redis struct{ client *redis.Client }

func NewRedis(raw string) (*Redis, error) {
	options, err := redis.ParseURL(raw)
	if err != nil {
		return nil, err
	}
	options.DialTimeout = 2 * time.Second
	options.ReadTimeout = time.Second
	options.WriteTimeout = time.Second
	return &Redis{redis.NewClient(options)}, nil
}
func (r *Redis) Get(ctx context.Context, key string) ([]byte, bool) {
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	b, err := r.client.Get(ctx, "datacat:"+key).Bytes()
	return b, err == nil
}
func (r *Redis) Set(ctx context.Context, key string, data []byte, ttl time.Duration) {
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	r.client.Set(ctx, "datacat:"+key, data, ttl)
}
func (r *Redis) Close() error { return r.client.Close() }
