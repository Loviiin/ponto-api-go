package main

import (
	"fmt"
	"log"

	"github.com/Loviiin/ponto-api-go/internal/config"
	"github.com/Loviiin/ponto-api-go/internal/model"

	"github.com/Loviiin/ponto-api-go/internal/domain/auth"
	"github.com/Loviiin/ponto-api-go/internal/domain/cargo"
	"github.com/Loviiin/ponto-api-go/internal/domain/empresa"
	"github.com/Loviiin/ponto-api-go/internal/domain/permissao"
	"github.com/Loviiin/ponto-api-go/internal/domain/ponto"
	"github.com/Loviiin/ponto-api-go/internal/domain/usuario"

	"github.com/Loviiin/ponto-api-go/pkg/funcoes"
	"github.com/Loviiin/ponto-api-go/pkg/jwt"
	// Vamos usar este pacote para as nossas constantes de permissão
	"github.com/Loviiin/ponto-api-go/pkg/permissions"

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

	// Adicionámos o &model.Permissao{} para a migração automática
	err = db.AutoMigrate(&model.Usuario{}, &model.RegistroPonto{}, &model.Empresa{}, &model.Cargo{}, &model.Permissao{})
	if err != nil {
		log.Fatal("Falha ao rodar a migração: ", err)
	}
	log.Println("Migração do banco de dados executada com sucesso.")
	log.Println("Migração do banco de dados executada com sucesso.")
	config.SeedPermissions(db)

	// --- Inicialização de Serviços e Repositórios ---
	jwtService := jwt.NewJWTService(cfg.JWTSecretKey, "ponto-api-go")
	funcoesService := funcoes.NewFuncoes()

	usuarioRepo := usuario.NewUsuarioRepository(db)
	pontoRepo := ponto.NewPontoRepository(db)
	empresaRepo := empresa.NewEmpresaRepository(db)
	cargoRepo := cargo.NewCargoRepository(db)
	permissaoRepo := permissao.NewRepository(db)

	usuarioService := usuario.NewUsuarioService(usuarioRepo)
	authService := auth.NewAuthService(usuarioRepo, jwtService)
	pontoService := ponto.NewPontoService(pontoRepo, usuarioRepo, empresaRepo)
	empresaService := empresa.NewEmpresaService(empresaRepo)
	cargoService := cargo.NewCargoService(cargoRepo)
	permissaoService := permissao.NewService(permissaoRepo)

	usuarioHandler := usuario.NewUsuarioHandler(usuarioService, empresaService, cargoService, funcoesService)
	authHandler := auth.NewAuthHandler(authService)
	pontoHandler := ponto.NewPontoHandler(pontoService)
	empresaHandler := empresa.NewEmpresaHandler(empresaService, funcoesService, db)
	cargoHandler := cargo.NewCargoHandler(cargoService, funcoesService)
	permissaoHandler := permissao.NewHandler(permissaoService)

	// --- Middlewares ---
	authMiddleware := auth.AuthMiddleware(jwtService)

	// Criamos os nossos middlewares de permissão aqui.
	// Cada um verifica uma permissão específica.
	canEditEmpresa := auth.PermissionMiddleware(usuarioService, funcoesService, permissions.EDITAR_EMPRESA)
	canDeleteEmpresa := auth.PermissionMiddleware(usuarioService, funcoesService, permissions.DELETAR_EMPRESA)
	canDeleteUsuario := auth.PermissionMiddleware(usuarioService, funcoesService, permissions.DELETAR_USUARIO)
	canManageCargos := auth.PermissionMiddleware(usuarioService, funcoesService, permissions.GERENCIAR_CARGOS)

	// --- Rotas da API ---
	router := gin.Default()
	apiV1 := router.Group("/api/v1")
	{
		// Rotas Públicas
		apiV1.POST("/auth/login", authHandler.Login)
		apiV1.POST("/usuarios", usuarioHandler.CriarUsuarioHandler)

		// Rotas para Super-Admin (no futuro, proteger com um middleware de "SuperAdmin")
		apiV1.POST("/permissoes", permissaoHandler.Create)
		apiV1.GET("/permissoes", permissaoHandler.FindAll)

		apiV1.POST("/empresas", empresaHandler.CriarEmpresaHandler)
		apiV1.GET("/empresas/:id", empresaHandler.GetEmpresaByIDHandler)
		apiV1.POST("/cargos", cargoHandler.CreateCargo)
		apiV1.POST("/cargos/:id/permissoes/:permissaoId", cargoHandler.AddPermissionToCargo)

		// Rotas Protegidas (requerem login básico)
		rotasProtegidas := apiV1.Group("")
		rotasProtegidas.Use(authMiddleware)
		{
			// Rotas de Usuário
			rotasProtegidas.GET("/usuarios", usuarioHandler.GetAllUsuariosHandler)
			rotasProtegidas.GET("/usuarios/:id", usuarioHandler.GetByIdHandler)
			rotasProtegidas.PUT("/usuarios/:id", usuarioHandler.UpdateUsuarioHandler) // Utilizador só pode alterar a si mesmo
			rotasProtegidas.GET("/usuarios/me", usuarioHandler.GetMeuPerfil)

			// Agora, para apagar um utilizador, é preciso a permissão DELETAR_USUARIO
			rotasProtegidas.DELETE("/usuarios/:id", canDeleteUsuario, usuarioHandler.DeleteHandler)

			// Rota de Ponto
			rotasProtegidas.POST("/pontos", pontoHandler.BaterPonto)

			// Rotas de Empresa (Ações gerais)
			rotasProtegidas.GET("/empresas", empresaHandler.GetAllEmpresasHandler)

			// Rotas de Empresa (Ações Administrativas, protegidas por permissão)
			rotasProtegidas.PUT("/empresas/:id", canEditEmpresa, empresaHandler.UpdateEmpresaHandler)
			rotasProtegidas.DELETE("/empresas/:id", canDeleteEmpresa, empresaHandler.DeleteEmpresaHandler)

			// A gestão de cargos (apagar, atualizar, adicionar permissões) continua protegida.
			rotasProtegidas.GET("/cargos", cargoHandler.GetAllCargos)
			rotasProtegidas.PUT("/cargos/:id", canManageCargos, cargoHandler.UpdateCargo)
			rotasProtegidas.DELETE("/cargos/:id", canManageCargos, cargoHandler.DeleteCargo)
		}
	}

	log.Printf("Servidor iniciado e ouvindo na porta %s", cfg.APIPort)
	err = router.Run(":" + cfg.APIPort)
	if err != nil {
		log.Fatal("Falha ao iniciar o servidor: ", err)
	}
}
