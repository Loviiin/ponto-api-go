package cache

import (
	"context"
	"time"

	redis "github.com/go-redis/redis/v8"
)

// redisService implementa a interface Service usando o cliente go-redis.
type redisService struct {
	client *redis.Client
}

// NewRedisService cria um novo serviço de cache baseado em Redis.
// addr: endereço do Redis (ex: "localhost:6379").
// password: senha do Redis, se houver.
// db: índice do banco de dados Redis a ser usado.
func NewRedisService(addr, password string, db int) (Service, error) {
	opts := &redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	}

	client := redis.NewClient(opts)

	// Testa a conexão com um Ping com timeout curto.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &redisService{client: client}, nil
}

// Get retorna o valor de uma chave do cache.
// Se a chave não existir, retorna "" e erro nil.
func (r *redisService) Get(ctx context.Context, key string) (string, error) {
	val, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", nil
	}
	return val, err
}

// Set define um valor no cache com uma expiração.
func (r *redisService) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return r.client.Set(ctx, key, value, expiration).Err()
}

// Delete remove uma chave do cache.
func (r *redisService) Delete(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

// FlushAll limpa todas as chaves do DB atual.
func (r *redisService) FlushAll(ctx context.Context) error {
	return r.client.FlushDB(ctx).Err()
}
