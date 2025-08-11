package main

import (
	"fmt"
	auth2 "github.com/Loviiin/ponto-api-go/internal/domain/auth"
	"github.com/Loviiin/ponto-api-go/internal/domain/empresa"
	ponto2 "github.com/Loviiin/ponto-api-go/internal/domain/ponto"
	usuario2 "github.com/Loviiin/ponto-api-go/internal/domain/usuario"
	"github.com/Loviiin/ponto-api-go/pkg/funcoes"
	"github.com/Loviiin/ponto-api-go/pkg/jwt"
	"log"

	// Nossos pacotes internos que criamos
	"github.com/Loviiin/ponto-api-go/internal/config"
	"github.com/Loviiin/ponto-api-go/internal/model"
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

	err = db.AutoMigrate(&model.Usuario{}, &model.RegistroPonto{}, &model.Empresa{}, &model.Cargo{})
	if err != nil {
		log.Fatal("Falha ao rodar a migração: ", err)
	}
	log.Println("Migração do banco de dados executada com sucesso.")

	jwtService := jwt.NewJWTService(cfg.JWTSecretKey, "ponto-api-go")
	usuarioRepo := usuario2.NewUsuarioRepository(db)
	pontoRepo := ponto2.NewPontoRepository(db)
	empresaRepo := empresa.NewEmpresaRepository(db)
	funcoesService := funcoes.NewFuncoes()

	usuarioService := usuario2.NewUsuarioService(usuarioRepo)
	authService := auth2.NewAuthService(usuarioRepo, jwtService)
	pontoService := ponto2.NewPontoService(pontoRepo, usuarioRepo, empresaRepo)
	empresaService := empresa.NewEmpresaService(empresaRepo)

	usuarioHandler := usuario2.NewUsuarioHandler(usuarioService, funcoesService)
	authHandler := auth2.NewAuthHandler(authService)
	pontoHandler := ponto2.NewPontoHandler(pontoService)
	empresaHandler := empresa.NewEmpresaHandler(empresaService, funcoesService, usuarioService)

	authMiddleware := auth2.AuthMiddleware(jwtService)
	//adminAuthMiddleware: = auth2.RoleAuthMiddleware(usuarioService, funcoesService, "ADMIN")

	router := gin.Default()

	// Agrupamos as rotas da nossa API sob o prefixo /api/v1 para organização.
	apiV1 := router.Group("/api/v1")
	{
		apiV1.POST("/auth/login", authHandler.Login)
		apiV1.POST("/usuarios", usuarioHandler.CriarUsuarioHandler)
		apiV1.POST("/empresas", empresaHandler.CriarEmpresaHandler)
		apiV1.GET("/empresas", empresaHandler.GetAllEmpresasHandler)
		apiV1.GET("/empresas/:id", empresaHandler.GetEmpresaByIDHandler)

		rotasProtegidas := apiV1.Group("")
		rotasProtegidas.Use(authMiddleware)
		{
			// Agora, mova as rotas que você quer proteger para dentro deste bloco.
			rotasProtegidas.GET("/usuarios", usuarioHandler.GetAllUsuariosHandler)
			rotasProtegidas.GET("/usuarios/:id", usuarioHandler.GetByIdHandler)
			rotasProtegidas.PUT("/usuarios/:id", usuarioHandler.UpdateUsuarioHandler)
			rotasProtegidas.GET("/usuarios/me", usuarioHandler.GetMeuPerfil)
			rotasProtegidas.DELETE("/usuarios/:id", usuarioHandler.DeleteHandler)
			rotasProtegidas.POST("/pontos", pontoHandler.BaterPonto)
			rotasProtegidas.PUT("/empresas/:id", empresaHandler.UpdateEmpresaHandler)
			rotasProtegidas.DELETE("/empresas/:id", empresaHandler.DeleteEmpresaHandler)
			//		rotasProtegidas.PUT("/empresas/:id", adminAuthMiddleware, empresaHandler.UpdateEmpresaHandler)
			//		rotasProtegidas.DELETE("/empresas/:id", adminAuthMiddleware, empresaHandler.DeleteEmpresaHandler)
			//adicionar para quando tiver a verificação de cargos
		}
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
