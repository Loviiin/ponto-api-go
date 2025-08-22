# --- Estágio 1: Build ---
# Começamos com uma imagem oficial do Go. A tag 'alpine' é de uma versão leve.
FROM golang:1.24.5-alpine AS builder

# Definimos o nosso diretório de trabalho dentro do container
WORKDIR /app

# Copiamos os ficheiros de dependências primeiro. O Docker é inteligente e só vai
# descarregar as dependências de novo se estes ficheiros mudarem.
COPY go.mod go.sum ./
RUN go mod download

# Agora, copiamos tod o resto do código-fonte do nosso projeto para o container
COPY . .

# O comando principal: compilamos a nossa aplicação.
# CGO_ENABLED=0 cria um binário estático, que não depende de bibliotecas do sistema.
# -o ./out/ponto-api diz para colocar o executável compilado na pasta 'out' com o nome 'ponto-api'.
# O alvo é o nosso ficheiro principal.
RUN CGO_ENABLED=0 GOOS=linux go build -o ./out/ponto-api ./cmd/api/main.go

# --- Estágio 2: Final ---
# Começamos com uma imagem 'alpine', que é uma das menores imagens Linux disponíveis.
FROM alpine:latest

RUN apk add --no-cache tzdata

# Novamente, definimos o diretório de trabalho.
WORKDIR /app

# A parte mais importante: copiamos APENAS o binário compilado do estágio 'builder'.
COPY --from=builder /app/out/ponto-api .

# Também precisamos do nosso ficheiro de configuração .env para a aplicação saber
# como se conectar ao banco de dados, em que porta rodar, etc.
COPY .env .

# Expomos a porta 8083, que é a que a sua API usa (de acordo com o README).
# Isto diz ao Docker que o container vai "ouvir" nesta porta.
EXPOSE 8083

# O comando final que será executado quando o container iniciar.
# Simplesmente executa o nosso binário.
CMD ["./ponto-api"]