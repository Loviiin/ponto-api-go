# Structured Logging with slog

Este documento descreve como o sistema de logging estruturado foi implementado usando o pacote `slog` da biblioteca padrão do Go.

## Visão Geral

O sistema de logging foi migrado de `log` para `slog` (structured logging) para fornecer:

- **Logs estruturados em JSON** para fácil integração com ferramentas de monitoramento
- **Campos estruturados** que podem ser pesquisados e filtrados
- **Contexto rico** em cada log entry
- **Compatibilidade** com Datadog, Splunk, ELK Stack, CloudWatch, etc.

## Configuração

### Modos de Operação

O logger suporta dois modos:

1. **Modo Produção** (`ENVIRONMENT=production`):
   - Formato JSON estruturado
   - Otimizado para ferramentas de monitoramento
   - Fácil parsing e indexação

2. **Modo Desenvolvimento** (padrão):
   - Formato texto legível por humanos
   - Melhor para debugging local
   - Mais fácil de ler no console

### Inicialização

No `main.go`:

```go
import "github.com/Loviiin/ponto-api-go/pkg/logger"

// Detecta ambiente baseado na variável ENVIRONMENT
isProduction := os.Getenv("ENVIRONMENT") == "production"
log := logger.NewLogger(isProduction)
logger.SetDefault(log)
```

## Uso

### Logging Básico

```go
import "log/slog"

// Info - informações gerais
slog.Info("servidor iniciado", slog.String("port", "8083"))

// Warn - avisos
slog.Warn("cache miss", slog.String("key", "user:123"))

// Error - erros que precisam atenção
slog.Error("falha ao conectar", slog.Any("error", err))
```

### Campos Estruturados

Sempre use campos estruturados ao invés de interpolação de strings:

```go
// ❌ Evite:
slog.Info(fmt.Sprintf("Usuário %d criado", userID))

// ✅ Use:
slog.Info("usuário criado", slog.Uint64("user_id", uint64(userID)))
```

Tipos de campos disponíveis:
- `slog.String(key, value)` - strings
- `slog.Int(key, value)` - inteiros
- `slog.Uint64(key, value)` - unsigned integers
- `slog.Float64(key, value)` - floats
- `slog.Bool(key, value)` - booleanos
- `slog.Duration(key, value)` - duração de tempo
- `slog.Time(key, value)` - timestamps
- `slog.Any(key, value)` - qualquer tipo (use com moderação)

### Logging de HTTP Requests

O middleware `logger.Middleware` automaticamente loga todas as requisições HTTP:

```go
router.Use(logger.Middleware(log))
```

Cada requisição gera um log com:
- `method` - método HTTP (GET, POST, etc.)
- `path` - caminho da URL
- `status` - código de status HTTP
- `duration` - tempo de processamento
- `client_ip` - IP do cliente
- `user_agent` - user agent do cliente

### Logging com Contexto

Você pode adicionar o logger ao contexto e recuperá-lo depois:

```go
import "github.com/Loviiin/ponto-api-go/pkg/logger"

// Adicionar logger ao contexto
ctx := logger.WithLogger(ctx, log)

// Recuperar logger do contexto
log := logger.FromContext(ctx)
log.Info("operação executada")
```

## Exemplo de Output

### Modo Produção (JSON)

```json
{
  "time": "2025-10-31T21:00:00Z",
  "level": "INFO",
  "msg": "http request",
  "method": "POST",
  "path": "/api/v1/pontos",
  "status": 201,
  "duration": 45000000,
  "client_ip": "192.168.1.1",
  "user_agent": "Mozilla/5.0..."
}
```

### Modo Desenvolvimento (Text)

```
time=2025-10-31T21:00:00Z level=INFO msg="http request" method=POST path=/api/v1/pontos status=201 duration=45ms client_ip=192.168.1.1
```

## Integração com Ferramentas de Monitoramento

### Datadog

Os logs JSON são automaticamente compatíveis com Datadog. Configure o Datadog Agent para ler os logs da aplicação:

```yaml
logs:
  - type: file
    path: /var/log/ponto-api.log
    service: ponto-api
    source: go
```

### Splunk

Configure o Splunk para ingerir logs JSON:

```
[source::ponto-api]
INDEXED_EXTRACTIONS = json
```

### ELK Stack

Use Filebeat ou Logstash para enviar logs para Elasticsearch:

```yaml
filebeat.inputs:
- type: log
  enabled: true
  paths:
    - /var/log/ponto-api.log
  json.keys_under_root: true
```

## Boas Práticas

1. **Use campos estruturados**: Sempre prefira campos estruturados ao invés de mensagens interpoladas
2. **Níveis apropriados**:
   - `Info`: Operações normais e importantes
   - `Warn`: Situações anormais mas recuperáveis
   - `Error`: Erros que precisam atenção
3. **Contexto suficiente**: Inclua IDs relevantes (user_id, empresa_id, request_id)
4. **Evite logs excessivos**: Não logue em loops de alta frequência
5. **Sensibilidade de dados**: Nunca logue senhas, tokens, ou dados sensíveis

## Migração do log para slog

Se você encontrar código antigo usando `log`:

```go
// Antes
import "log"
log.Printf("Erro ao processar: %v", err)

// Depois
import "log/slog"
slog.Error("erro ao processar", slog.Any("error", err))
```

## Variáveis de Ambiente

- `ENVIRONMENT=production` - Ativa modo de produção com logs JSON
- Qualquer outro valor ou não definido - Ativa modo de desenvolvimento com logs texto
