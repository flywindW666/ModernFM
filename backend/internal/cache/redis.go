package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisCache struct {
	client *redis.Client
	ctx    context.Background()
}

func NewRedisCache(addr string) *RedisCache {
	rdb := redis.NewClient(&redis.Options{
		Addr: addr,
	})
	return &RedisCache{
		client: rdb,
		ctx:    context.Background(),
	}
}

// SetDirCache 缓存目录列表数据，默认过期时间 1 小时
func (r *RedisCache) SetDirCache(path string, data interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	// 使用 key 格式 "dir:v1:/movies"
	return r.client.Set(r.ctx, "dir:v1:"+path, jsonData, 1*time.Hour).Err()
}

// GetDirCache 获取目录缓存
func (r *RedisCache) GetDirCache(path string) (string, error) {
	return r.client.Get(r.ctx, "dir:v1:"+path).Result()
}

// InvalidatePath 当文件发生变动时，使相关路径缓存失效
func (r *RedisCache) InvalidatePath(path string) {
	// 简单逻辑：删除该路径及其父目录的缓存
	r.client.Del(r.ctx, "dir:v1:"+path)
	parent := "/"+path // 简化处理
	r.client.Del(r.ctx, "dir:v1:"+parent)
}
