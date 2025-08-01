package main

import (
	"fmt"
	"log"

	// Nossos pacotes internos que criamos
	"github.com/Loviiin/ponto-api-go/internal/config"
	"github.com/Loviiin/ponto-api-go/internal/handler"
	"github.com/Loviiin/ponto-api-go/internal/model"
	"github.com/Loviiin/ponto-api-go/internal/repository"
	"github.com/Loviiin/ponto-api-go/internal/service"

	// Pacotes externos (nossas dependências)
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// --- PASSO 1: Carregar as Configurações ---
	// Chamamos nossa função LoadConfig para ler o arquivo .env.
	// Se der erro, a aplicação não pode continuar, então usamos log.Fatal para parar tudo.
	cfg, err := config.LoadConfig(".")
	if err != nil {
		log.Fatal("Não foi possível carregar as configurações: ", err)
	}

	// --- PASSO 2: Conectar ao Banco de Dados ---
	// Montamos a String de Conexão (DSN) com os dados que vieram do .env.
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=America/Sao_Paulo",
		cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort)

	// Usamos o GORM para abrir a conexão. Se falhar, a aplicação para.
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Falha ao conectar ao banco de dados: ", err)
	}
	log.Println("Conexão com o banco de dados estabelecida com sucesso.")

	// --- PASSO 3: Rodar a Migração Automática ---
	// O AutoMigrate garante que a tabela 'usuarios' exista no banco e
	// que ela tenha todas as colunas da nossa struct 'model.Usuario'.
	err = db.AutoMigrate(&model.Usuario{})
	if err != nil {
		log.Fatal("Falha ao rodar a migração: ", err)
	}
	log.Println("Migração do banco de dados executada com sucesso.")

	// --- PASSO 4: Montar a "Linha de Montagem" (Injeção de Dependências) ---
	// Aqui criamos as instâncias de cada camada, passando as dependências de "dentro para fora".

	// 1. Camada de Repositório (O "Arquivista")
	// Criamos o repositório e entregamos a ele a conexão com o banco.
	usuarioRepo := repository.UsuarioRepository{Db: db}

	// 2. Camada de Serviço (O "Gerente")
	// Criamos o serviço e entregamos a ele o repositório.
	// Assim, o serviço pode dar ordens para o repositório.
	// Usamos um ponteiro para o repositório, que é uma boa prática.
	usuarioService := service.UsuarioService{UsuarioRepository: &usuarioRepo}

	// 3. Camada de Handler (O "Atendente de Front Desk")
	// Criamos o handler e entregamos a ele o serviço.
	// O handler receberá os pedidos e os passará para o serviço.
	usuarioHandler := handler.NewUsuarioHandler(&usuarioService)

	// --- PASSO 5: Configurar as Rotas (O Painel do Atendente) ---
	// Inicializamos o roteador do Gin.
	router := gin.Default()

	// Agrupamos as rotas da nossa API sob o prefixo /api/v1 para organização.
	apiV1 := router.Group("/api/v1")
	{
		// Criamos a rota para CRIAR um usuário.
		// Quando uma requisição POST chegar em /api/v1/usuarios,
		// o método CriarUsuarioHandler do nosso handler será executado.
		apiV1.POST("/usuarios", usuarioHandler.CriarUsuarioHandler)

		// ... AQUI É ONDE VOCÊ ADICIONARÁ AS OUTRAS ROTAS NO FUTURO ...
		// Ex: apiV1.POST("/login", authHandler.Login)
		// Ex: apiV1.GET("/usuarios/:id", usuarioHandler.BuscarUsuarioPorIDHandler)
	}

	// --- PASSO 6: Ligar o Servidor ---
	// O maestro dá a ordem final, e a orquestra começa a tocar.
	// O servidor começa a ouvir por requisições na porta definida no .env.
	log.Printf("Servidor iniciado e ouvindo na porta %s", cfg.APIPort)
	err = router.Run(":" + cfg.APIPort)
	if err != nil {
		log.Fatal("Falha ao iniciar o servidor: ", err)
	}
}
