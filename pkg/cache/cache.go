package cache

import (
	"context"
	"time"
)

// Service é a interface para nossa camada de cache.
type Service interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Delete(ctx context.Context, key string) error
}

// Flusher é uma interface opcional para provedores que suportam limpar todo o cache.
// Implementações que a suportarem podem ser usadas para invalidar tudo em cenários como reset de banco.
type Flusher interface {
	FlushAll(ctx context.Context) error
}
