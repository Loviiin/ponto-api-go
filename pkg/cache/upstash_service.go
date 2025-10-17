package cache

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// upstashService implementa Service usando a API REST do Upstash.
type upstashService struct {
	httpClient *http.Client
	baseURL    string
	token      string
}

// NewUpstashService cria um serviço de cache que usa a API REST do Upstash.
// baseURL: UPSTASH_REDIS_REST_URL, token: UPSTASH_REDIS_REST_TOKEN.
func NewUpstashService(baseURL, token string) (Service, error) {
	if baseURL == "" || token == "" {
		return nil, errors.New("upstash baseURL e token são obrigatórios")
	}
	return &upstashService{
		httpClient: &http.Client{Timeout: 5 * time.Second},
		baseURL:    strings.TrimRight(baseURL, "/"),
		token:      token,
	}, nil
}

func (u *upstashService) do(ctx context.Context, body interface{}, v interface{}) error {
	b, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.baseURL, strings.NewReader(string(b)))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+u.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := u.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("upstash status %d", resp.StatusCode)
	}
	if v == nil {
		return nil
	}
	dec := json.NewDecoder(resp.Body)
	return dec.Decode(v)
}

type upstashResp struct {
	Result interface{} `json:"result"`
}

func (u *upstashService) Get(ctx context.Context, key string) (string, error) {
	var out upstashResp
	// GET key
	if err := u.do(ctx, []interface{}{"GET", key}, &out); err != nil {
		return "", err
	}
	if out.Result == nil {
		return "", nil
	}
	switch v := out.Result.(type) {
	case string:
		return v, nil
	default:
		// Upstash pode retornar base64 em binários; tentar converter
		if m, ok := out.Result.(map[string]interface{}); ok {
			if s, ok := m["data"].(string); ok {
				if dec, err := base64.StdEncoding.DecodeString(s); err == nil {
					return string(dec), nil
				}
			}
		}
		return "", nil
	}
}

func (u *upstashService) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	// SET key value EX seconds
	secs := int(expiration.Seconds())
	args := []interface{}{"SET", key, value}
	if secs > 0 {
		args = append(args, "EX", secs)
	}
	return u.do(ctx, args, nil)
}

func (u *upstashService) Delete(ctx context.Context, key string) error {
	// DEL key
	return u.do(ctx, []interface{}{"DEL", key}, nil)
}

// FlushAll tenta executar um script EVAL para limpar as chaves.
// Aviso: Upstash REST não expõe FLUSHDB direto; esta abordagem lista e deleta por padrão.
func (u *upstashService) FlushAll(ctx context.Context) error {
	// Scan + Del via pipeline simples (limite para evitar tempo excessivo)
	// Aqui executamos um EVAL que deleta por padrão todas as chaves encontradas por KEYS * (uso cuidadoso em prod)
	// Alternativa: manter prefixos de chave e excluir por prefixo.
	// KEYS * pode ser caro; use com cautela e apenas em cenários controlados (reset DB dev/hml como solicitado).
	return u.do(ctx, []interface{}{"EVAL", "for _,k in ipairs(redis.call('keys','*')) do redis.call('del',k) end return 1", 0}, nil)
}
