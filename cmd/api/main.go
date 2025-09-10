package main

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/Loviiin/ponto-api-go/docs"
	"github.com/Loviiin/ponto-api-go/internal/config"
	"github.com/Loviiin/ponto-api-go/internal/domain/bancohoras"
	"github.com/Loviiin/ponto-api-go/internal/domain/justificativa"
	"github.com/Loviiin/ponto-api-go/internal/domain/logbancohoras"
	"github.com/Loviiin/ponto-api-go/internal/model"
	"github.com/gin-contrib/cors"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

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

var allowedOriginRegex = regexp.MustCompile(`^https?:\/\/.*\.app\.github\.dev$`)


// @title           Ponto API em Go
// @version         1.0
// @description     API de alta performance para um sistema de Ponto Eletrônico.
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  MIT
// @license.url   https://github.com/Loviiin/ponto-api-go/blob/main/LICENSE

// @host      localhost:8083
// @BasePath  /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description "Digite 'Bearer' seguido de um espaço e o seu token."
// @type apiKey
func main() {
	cfg, err := config.LoadConfig(".")
	if err != nil {
		log.Fatal("Não foi possível carregar as configurações: ", err)
	}

	var dsn string
	if strings.HasPrefix(cfg.DBHost, "/") {
		dsn = fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=America/Sao_Paulo",
			cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode)
	} else {

		dsn = fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=America/Sao_Paulo",
			cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort, cfg.DBSSLMode)
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Falha ao conectar ao banco de dados: ", err)
	}
	log.Println("Conexão com o banco de dados estabelecida com sucesso.")

	// Adicionámos o &model.Permissao{} para a migração automática
	err = db.AutoMigrate(&model.Usuario{}, &model.RegistroPonto{}, &model.Empresa{}, &model.Cargo{}, &model.Permissao{}, &model.Justificativa{}, &model.LogBancoHoras{})
	if err != nil {
		log.Fatal("Falha ao rodar a migração: ", err)
	}
	log.Println("Migração do banco de dados executada com sucesso.")
	log.Println("Migração do banco de dados executada com sucesso.")
	config.SeedPermissions(db)
	config.SeedSuperAdmin(db)

	// --- Inicialização de Serviços e Repositórios ---
	jwtService := jwt.NewJWTService(cfg.JWTSecretKey, "ponto-api-go")
	funcoesService := funcoes.NewFuncoes()

	usuarioRepo := usuario.NewUsuarioRepository(db)
	pontoRepo := ponto.NewPontoRepository(db)
	empresaRepo := empresa.NewEmpresaRepository(db)
	cargoRepo := cargo.NewCargoRepository(db)
	permissaoRepo := permissao.NewRepository(db)
	justificativaRepo := justificativa.NewRepository(db)
	logBancoHorasRepo := logbancohoras.NewRepository(db)

	usuarioService := usuario.NewUsuarioService(usuarioRepo, cargoRepo, empresaRepo)
	authService := auth.NewAuthService(usuarioRepo, jwtService)
	pontoService := ponto.NewPontoService(pontoRepo, usuarioRepo, empresaRepo, db)

	empresaService := empresa.NewEmpresaService(empresaRepo)
	cargoService := cargo.NewCargoService(cargoRepo)
	permissaoService := permissao.NewService(permissaoRepo)
	bancoHorasService := bancohoras.NewBancoHorasService(pontoRepo, usuarioRepo, logBancoHorasRepo, db)
	justificativaService := justificativa.NewService(justificativaRepo, pontoRepo, db)

	usuarioHandler := usuario.NewUsuarioHandler(usuarioService, funcoesService)
	authHandler := auth.NewAuthHandler(authService)
	pontoHandler := ponto.NewPontoHandler(pontoService, justificativaService, funcoesService)

	empresaHandler := empresa.NewEmpresaHandler(empresaService, funcoesService, db)
	cargoHandler := cargo.NewCargoHandler(cargoService, funcoesService)
	permissaoHandler := permissao.NewHandler(permissaoService)
	bancoHorasHandler := bancohoras.NewBancoHorasHandler(bancoHorasService, usuarioService, funcoesService)
	justificativaHandler := justificativa.NewHandler(justificativaService, funcoesService)

	// --- Middlewares ---
	authMiddleware := auth.AuthMiddleware(jwtService)

	// Criamos os nossos middlewares de permissão aqui.
	// Cada um verifica uma permissão específica.
	canEditEmpresa := auth.PermissionMiddleware(usuarioService, funcoesService, permissions.EDITAR_EMPRESA)
	canDeleteEmpresa := auth.PermissionMiddleware(usuarioService, funcoesService, permissions.DELETAR_EMPRESA)
	canDeleteUsuario := auth.PermissionMiddleware(usuarioService, funcoesService, permissions.DELETAR_USUARIO)
	canManageCargos := auth.PermissionMiddleware(usuarioService, funcoesService, permissions.GERENCIAR_CARGOS)
	canEditSaldo := auth.PermissionMiddleware(usuarioService, funcoesService, permissions.EDITAR_SALDO_FUNCIONARIOS)
	canViewPonto := auth.PermissionMiddleware(usuarioService, funcoesService, permissions.VISUALIZAR_PONTO_FUNCIONARIOS)
	canAdjustPonto := auth.PermissionMiddleware(usuarioService, funcoesService, permissions.AJUSTAR_PONTO_FUNCIONARIOS)
	canManageJustificativas := auth.PermissionMiddleware(usuarioService, funcoesService, permissions.GERENCIAR_JUSTIFICATIVAS)

	//	scheduler := scheduler.NewScheduler(bancoHorasService, usuarioService)
	//	scheduler.Start()

	// --- Rotas da API ---
	router := gin.Default()

	router.SetTrustedProxies(nil)

// --- CONFIGURAÇÃO DE CORS OTIMIZADA E SEGURA ---
configCORS := cors.DefaultConfig()
configCORS.AllowCredentials = true

// Para produção (ex: Cloud Run), você usaria uma origem específica.
// Para desenvolvimento no Codespaces, usamos a função com o regex pré-compilado.
configCORS.AllowOriginFunc = func(origin string) bool {
    // A variável 'allowedOriginRegex' deve ser definida no topo do seu arquivo main.go:
    // var allowedOriginRegex = regexp.MustCompile(`^https?:\/\/.*\.app\.github\.dev$`)
    return allowedOriginRegex.MatchString(origin)
}

configCORS.AllowHeaders = []string{"Authorization", "Content-Type", "Origin"}
router.Use(cors.New(configCORS))


// --- CONFIGURAÇÃO DINÂMICA DO SWAGGER ---
// Verifica a variável de ambiente para determinar o ambiente de execução.
if os.Getenv("ENVIRONMENT") == "production" {
    // Em produção (Cloud Run, Codespaces), apaga o host para usar um caminho relativo.
    docs.SwaggerInfo.Host = ""
}

docs.SwaggerInfo.BasePath = "/api/v1"
router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

// Define o grupo de rotas da API.
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

		// Rota para tarefas internas, a ser chamada pelo Cloud Scheduler
		apiV1.POST("/tasks/fechamento-diario", bancoHorasHandler.ExecutarFechamentoDiario)

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
			rotasProtegidas.GET("/pontos/meus-registros", pontoHandler.GetMeusRegistos)
			rotasProtegidas.GET("/pontos/usuario/:id", canViewPonto, pontoHandler.GetRegistosPorUsuarioID)
			rotasProtegidas.POST("/pontos/ajuste", canAdjustPonto, pontoHandler.AjustarPonto)
			rotasProtegidas.PUT("/pontos/:pontoId", canAdjustPonto, pontoHandler.EditarPonto)

			// Rotas de Empresa (Ações gerais)
			rotasProtegidas.GET("/empresas", empresaHandler.GetAllEmpresasHandler)

			// Rotas de Empresa (Ações Administrativas, protegidas por permissão)
			rotasProtegidas.PUT("/empresas/:id", canEditEmpresa, empresaHandler.UpdateEmpresaHandler)
			rotasProtegidas.DELETE("/empresas/:id", canDeleteEmpresa, empresaHandler.DeleteEmpresaHandler)

			// A gestão de cargos (apagar, atualizar, adicionar permissões) continua protegida.
			rotasProtegidas.GET("/cargos", cargoHandler.GetAllCargos)
			rotasProtegidas.PUT("/cargos/:id", canManageCargos, cargoHandler.UpdateCargo)
			rotasProtegidas.DELETE("/cargos/:id", canManageCargos, cargoHandler.DeleteCargo)

			rotasProtegidas.GET("/bancohoras/saldo/usuario/:id", bancoHorasHandler.GetSaldoDoDia)
			rotasProtegidas.POST("/bancohoras/fechamento/usuario/:id", canEditSaldo, bancoHorasHandler.FecharDia)

			// --- NOVAS ROTAS DE JUSTIFICATIVAS ---
			// Rota para o funcionário criar uma solicitação
			rotasProtegidas.POST("/justificativas", justificativaHandler.SolicitarAjuste)

			// Rotas para o admin/gestor gerir as solicitações
			rotasProtegidas.GET("/justificativas/pendentes", canManageJustificativas, justificativaHandler.ListarPendentes)
			rotasProtegidas.POST("/justificativas/:id/processar", canManageJustificativas, justificativaHandler.AprovarReprovar)

		}
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = cfg.APIPort
	}
	log.Printf("Servidor iniciado e ouvindo na porta %s", port)
	err = router.Run(":" + port)
	if err != nil {
		log.Fatal("Falha ao iniciar o servidor: ", err)
	}
}
