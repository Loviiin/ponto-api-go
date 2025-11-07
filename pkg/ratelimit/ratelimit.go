package ratelimit

import (
	"fmt"
	"sync"
	"time"
)

// RateLimiter implementa um rate limiter simples usando token bucket
type RateLimiter struct {
	mu       sync.RWMutex
	attempts map[string]*attemptInfo
	maxAttempts int
	windowSize time.Duration
}

type attemptInfo struct {
	count     int
	firstTime time.Time
	blockedUntil time.Time
}

// NewRateLimiter cria um novo rate limiter
// maxAttempts: número máximo de tentativas permitidas
// windowSize: janela de tempo para contar as tentativas
func NewRateLimiter(maxAttempts int, windowSize time.Duration) *RateLimiter {
	rl := &RateLimiter{
		attempts: make(map[string]*attemptInfo),
		maxAttempts: maxAttempts,
		windowSize: windowSize,
	}
	
	// Limpar tentativas antigas a cada 5 minutos
	go rl.cleanupOldAttempts()
	
	return rl
}

// IsAllowed verifica se a tentativa é permitida para a chave (ex: email ou IP)
// Retorna: (allowed, timeUntilRetry)
func (rl *RateLimiter) IsAllowed(key string) (bool, time.Duration) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	
	now := time.Now()
	info, exists := rl.attempts[key]
	
	// Se não existe registro, criar novo
	if !exists {
		rl.attempts[key] = &attemptInfo{
			count: 1,
			firstTime: now,
			blockedUntil: time.Time{}, // Não bloqueado
		}
		return true, 0
	}
	
	// Se está bloqueado, verificar se desbloqueou
	if !info.blockedUntil.IsZero() {
		if now.Before(info.blockedUntil) {
			// Ainda está bloqueado
			timeUntilRetry := info.blockedUntil.Sub(now)
			return false, timeUntilRetry
		}
		// Bloqueio expirou, resetar
		info.count = 1
		info.firstTime = now
		info.blockedUntil = time.Time{}
		return true, 0
	}
	
	// Verificar se a janela de tempo expirou
	if now.Sub(info.firstTime) > rl.windowSize {
		// Janela expirou, resetar
		info.count = 1
		info.firstTime = now
		info.blockedUntil = time.Time{}
		return true, 0
	}
	
	// Incrementar contador
	info.count++
	
	// Verificar se atingiu limite
	if info.count > rl.maxAttempts {
		// Bloquear por 15 minutos
		info.blockedUntil = now.Add(15 * time.Minute)
		timeUntilRetry := info.blockedUntil.Sub(now)
		return false, timeUntilRetry
	}
	
	return true, 0
}

// GetAttemptCount retorna o número de tentativas para uma chave
func (rl *RateLimiter) GetAttemptCount(key string) int {
	rl.mu.RLock()
	defer rl.mu.RUnlock()
	
	if info, exists := rl.attempts[key]; exists {
		// Se está bloqueado ou janela ainda está aberta
		if !info.blockedUntil.IsZero() || time.Since(info.firstTime) <= rl.windowSize {
			return info.count
		}
	}
	
	return 0
}

// Reset reseta o contador para uma chave
func (rl *RateLimiter) Reset(key string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	
	delete(rl.attempts, key)
}

// cleanupOldAttempts remove tentativas expiradas periodicamente
func (rl *RateLimiter) cleanupOldAttempts() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	
	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		
		for key, info := range rl.attempts {
			// Remover se a janela expirou há mais de 15 minutos
			if now.Sub(info.firstTime) > rl.windowSize+15*time.Minute {
				delete(rl.attempts, key)
			}
		}
		
		rl.mu.Unlock()
	}
}

// GetBlockedUntil retorna quando a chave será desbloqueada (ou tempo zero se não bloqueada)
func (rl *RateLimiter) GetBlockedUntil(key string) time.Time {
	rl.mu.RLock()
	defer rl.mu.RUnlock()
	
	if info, exists := rl.attempts[key]; exists {
		return info.blockedUntil
	}
	
	return time.Time{}
}

// GetStatus retorna status legível
func (rl *RateLimiter) GetStatus(key string) string {
	rl.mu.RLock()
	defer rl.mu.RUnlock()
	
	info, exists := rl.attempts[key]
	if !exists {
		return "Sem tentativas"
	}
	
	now := time.Now()
	
	if !info.blockedUntil.IsZero() && now.Before(info.blockedUntil) {
		minutosRestantes := int(info.blockedUntil.Sub(now).Minutes())
		return fmt.Sprintf("Bloqueado por %d minutos", minutosRestantes)
	}
	
	if now.Sub(info.firstTime) > rl.windowSize {
		return "Janela expirada"
	}
	
	tentativasRestantes := rl.maxAttempts - info.count + 1
	if tentativasRestantes < 0 {
		tentativasRestantes = 0
	}
	
	return fmt.Sprintf("%d tentativas (máximo %d)", info.count, rl.maxAttempts)
}
