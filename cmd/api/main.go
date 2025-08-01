package main

import (
	"fmt"
	"github.com/Loviiin/ponto-api-go/pkg/jwt"
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

	cfg, err := config.LoadConfig(".")
	if err != nil {
		log.Fatal("Não foi possível carregar as configurações: ", err)
	}

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=America/Sao_Paulo",
		cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Falha ao conectar ao banco de dados: ", err)
	}
	log.Println("Conexão com o banco de dados estabelecida com sucesso.")

	err = db.AutoMigrate(&model.Usuario{})
	if err != nil {
		log.Fatal("Falha ao rodar a migração: ", err)
	}
	log.Println("Migração do banco de dados executada com sucesso.")

	jwtService := jwt.NewJWTService(cfg.JWTSecretKey, "ponto-api-go")
	usuarioRepo := repository.NewUsuarioRepository(db)

	usuarioService := service.NewUsuarioService(usuarioRepo)
	authService := service.NewAuthService(usuarioRepo, jwtService)

	usuarioHandler := handler.NewUsuarioHandler(usuarioService)
	authHandler := handler.NewAuthHandler(authService)

	router := gin.Default()

	// Agrupamos as rotas da nossa API sob o prefixo /api/v1 para organização.
	apiV1 := router.Group("/api/v1")
	{

		apiV1.POST("/usuarios", usuarioHandler.CriarUsuarioHandler)
		apiV1.POST("/auth/login", authHandler.Login)

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
