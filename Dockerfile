# --- Estágio 1: Build ---
# Começamos com uma imagem oficial do Go. A tag 'alpine' é de uma versão leve.
FROM golang:1.24.5-alpine AS builder

# Definimos o nosso diretório de trabalho dentro do container
WORKDIR /app

# Copiamos os ficheiros de dependências primeiro. O Docker é inteligente e só vai
# descarregar as dependências de novo se estes ficheiros mudarem.
COPY go.mod go.sum ./
RUN go mod download

# Agora, copiamos todo o resto do código-fonte do nosso projeto para o container
COPY . .

# O comando principal: compilamos a nossa aplicação.
# CGO_ENABLED=0 cria um binário estático, que não depende de bibliotecas do sistema.
# -o ./out/ponto-api diz para colocar o executável compilado na pasta 'out' com o nome 'ponto-api'.
# O alvo é o nosso ficheiro principal.
# -ldflags="-w -s" remove informações de debug para diminuir o binário
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o ./out/ponto-api ./cmd/api/main.go

# --- Estágio 2: Final ---
# Começamos com uma imagem 'alpine', que é uma das menores imagens Linux disponíveis.
FROM alpine:latest

# Instala tzdata para suporte a timezones e ca-certificates para HTTPS
RUN apk add --no-cache tzdata ca-certificates

# Cria um grupo e usuário não-root para rodar a aplicação com mais segurança
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Novamente, definimos o diretório de trabalho.
WORKDIR /app

# A parte mais importante: copiamos APENAS o binário compilado do estágio 'builder'.
COPY --from=builder /app/out/ponto-api .
COPY docs ./docs

# Define o usuário não-root como o usuário padrão para execução
USER appuser

# Expomos a porta 8083
EXPOSE 8083

# O comando final que será executado quando o container iniciar.
CMD ["./ponto-api"]