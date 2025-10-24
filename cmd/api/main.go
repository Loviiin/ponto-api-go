package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/Loviiin/ponto-api-go/docs"
	"github.com/Loviiin/ponto-api-go/internal/config"
	"github.com/Loviiin/ponto-api-go/internal/domain/bancohoras"
	"github.com/Loviiin/ponto-api-go/internal/domain/contrato"
	"github.com/Loviiin/ponto-api-go/internal/domain/justificativa"
	"github.com/Loviiin/ponto-api-go/internal/domain/logbancohoras"
	"github.com/Loviiin/ponto-api-go/internal/domain/relatorio"
	"github.com/Loviiin/ponto-api-go/internal/model"
	"github.com/gin-contrib/cors"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/Loviiin/ponto-api-go/internal/domain/auth"
	"github.com/Loviiin/ponto-api-go/internal/domain/cargo"
	cephandler "github.com/Loviiin/ponto-api-go/internal/domain/cep"
	"github.com/Loviiin/ponto-api-go/internal/domain/empresa"
	"github.com/Loviiin/ponto-api-go/internal/domain/localidade"
	"github.com/Loviiin/ponto-api-go/internal/domain/permissao"
	"github.com/Loviiin/ponto-api-go/internal/domain/ponto"
	"github.com/Loviiin/ponto-api-go/internal/domain/usuario"

	"github.com/Loviiin/ponto-api-go/pkg/brasilapi"
	"github.com/Loviiin/ponto-api-go/pkg/cache"

	"github.com/Loviiin/ponto-api-go/pkg/cep"
	"github.com/Loviiin/ponto-api-go/pkg/distancematrix"
	"github.com/Loviiin/ponto-api-go/pkg/funcoes"
	"github.com/Loviiin/ponto-api-go/pkg/geolocation"
	"github.com/Loviiin/ponto-api-go/pkg/jwt"
	"github.com/Loviiin/ponto-api-go/pkg/scheduler"
	"github.com/Loviiin/ponto-api-go/pkg/viacep"

	// Vamos usar este pacote para as nossas constantes de permissão
	"github.com/Loviiin/ponto-api-go/pkg/permissions"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// NOVO: Função para resetar e popular o banco de dados
func resetAndSeedDatabase(db *gorm.DB) {
	log.Println("Iniciando reset do banco de dados...")

	// Apaga as tabelas na ordem correta para evitar problemas de chave estrangeira
	err := db.Migrator().DropTable(
		"usuario_cargos",   // Tabela de junção para Usuario e Cargo
		"cargo_permissoes", // Tabela de junção para Cargo e Permissao
		&model.RegistroPonto{},
		&model.Justificativa{},
		&model.LogBancoHoras{},
		&model.Contrato{},
		&model.Usuario{},
		&model.Permissao{},
		&model.Cargo{},
		&model.Empresa{},
		&model.Localidade{},
	)

	if err != nil {
		log.Fatalf("Falha ao apagar tabelas: %v", err)
	}
	log.Println("Tabelas antigas removidas.")

	// Recria as tabelas
	log.Println("Recriando tabelas com AutoMigrate...")
	err = db.AutoMigrate(
		&model.Usuario{},
		&model.RegistroPonto{},
		&model.Empresa{},
		&model.Cargo{},
		&model.Permissao{},
		&model.Justificativa{},
		&model.LogBancoHoras{},
		&model.Contrato{},
		&model.Localidade{},
	)
	if err != nil {
		log.Fatal("Falha ao rodar a migração: ", err)
	}
	log.Println("Tabelas recriadas com sucesso.")

	// Popula com dados iniciais (seeding)
	log.Println("Populando o banco de dados com dados iniciais...")
	config.SeedPermissions(db)
	config.SeedSuperAdmin(db)
	log.Println("Banco de dados resetado e populado com sucesso!")
}

// --- CONFIGURAÇÃO DE CORS ---
// Lista explícita de domínios permitidos (mais seguro e legível)
var allowedOrigins = map[string]bool{
	"https://nexora-app.vercel.app":                                    true,
	"https://meu-ponto-frontend.vercel.app":                            true,
	"https://meu-ponto-frontend-git-main-loviins-projects.vercel.app":  true,
	"https://meu-ponto-frontend-n9lx9odbj-loviins-projects.vercel.app": true,
}

// Regex para permitir todos os subdomínios de desenvolvimento do GitHub Codespaces
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

	resetDB := os.Getenv("RESET_DB_ON_START") == "true"
	if resetDB {
		resetAndSeedDatabase(db)
	} else {
		// Adicionámos o &model.Permissao{} para a migração automática
		err = db.AutoMigrate(&model.Usuario{}, &model.RegistroPonto{}, &model.Empresa{}, &model.Cargo{}, &model.Permissao{}, &model.Justificativa{}, &model.LogBancoHoras{}, &model.Contrato{}, &model.Localidade{})
		if err != nil {
			log.Fatal("Falha ao rodar a migração: ", err)
		}
		log.Println("Migração do banco de dados executada com sucesso.")
		config.SeedPermissions(db)
		config.SeedSuperAdmin(db)
	}

	// --- Inicialização de Serviços e Repositórios ---
	jwtService := jwt.NewJWTService(cfg.JWTSecretKey, "ponto-api-go")
	funcoesService := funcoes.NewFuncoes()

	// Inicializa o serviço de cache: preferir Upstash REST se variáveis estiverem presentes
	var cacheService cache.Service
	upstashURL := cfg.UpstashRedisRestURL
	upstashToken := cfg.UpstashRedisRestToken
	if upstashURL != "" && upstashToken != "" {
		cs, cErr := cache.NewUpstashService(upstashURL, upstashToken)
		if cErr != nil {
			log.Panicf("Falha ao inicializar Upstash REST: %v", cErr)
		}
		cacheService = cs
		log.Println("Cache: usando Upstash REST API")
	} else {
		cs, cErr := cache.NewRedisService(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
		if cErr != nil {
			log.Panicf("Falha ao conectar ao Redis: %v", cErr)
		}
		cacheService = cs
		log.Println("Cache: usando Redis nativo")
	}

	// Se resetDB == true, limpa também o cache se o provider suportar.
	if resetDB {
		if flusher, ok := cacheService.(cache.Flusher); ok {
			if err := flusher.FlushAll(context.Background()); err != nil {
				log.Printf("[cache] falha ao limpar cache após reset DB: %v", err)
			} else {
				log.Println("[cache] cache limpo após reset DB")
			}
		} else {
			log.Println("[cache] provider não suporta FlushAll; considere invalidar por prefixo")
		}
	}

	usuarioRepo := usuario.NewUsuarioRepository(db, cacheService)
	pontoRepo := ponto.NewPontoRepository(db)
	empresaRepo := empresa.NewEmpresaRepository(db, cacheService)
	cargoRepo := cargo.NewCargoRepository(db, cacheService)
	permissaoRepo := permissao.NewRepository(db, cacheService)
	justificativaRepo := justificativa.NewRepository(db)
	logBancoHorasRepo := logbancohoras.NewRepository(db)
	contratoRepo := contrato.NewContratoRepository(db)
	localidadeRepo := localidade.NewRepository(db)

	// CEP providers (BrasilAPI primário + ViaCEP fallback)
	brasilAPIClient := brasilapi.NewClient(cfg.BrasilApiUrl)
	viacepURL := "https://viacep.com.br/ws"
	viaCEPClient := viacep.NewClient(viacepURL)

	// Distance Matrix AI client
	distanceMatrixClient := distancematrix.NewClient(cfg.DistanceMatrixAPIKey)

	// Serviço de CEP com fallback (ViaCEP primário + BrasilAPI fallback)
	cepService := cep.NewService(brasilAPIClient, viaCEPClient)

	// Serviço de geolocalização: usa cadeia ViaCEP+DistanceMatrix primário e BrasilAPI como fallback
	geoService := geolocation.NewService(cepService, distanceMatrixClient, brasilAPIClient)

	usuarioService := usuario.NewUsuarioService(db, usuarioRepo, cargoRepo, empresaRepo, contratoRepo, localidadeRepo)
	authService := auth.NewAuthService(usuarioRepo, empresaRepo, cargoRepo, contratoRepo, localidadeRepo, geoService, jwtService, db)
	pontoService := ponto.NewPontoService(pontoRepo, usuarioRepo, localidadeRepo, db)

	empresaService := empresa.NewEmpresaService(empresaRepo)
	cargoService := cargo.NewCargoService(cargoRepo)
	permissaoService := permissao.NewService(permissaoRepo)
	bancoHorasService := bancohoras.NewBancoHorasServiceWithCache(pontoRepo, usuarioRepo, logBancoHorasRepo, db, cacheService)
	justificativaService := justificativa.NewService(justificativaRepo, pontoRepo, db)
	localidadeService := localidade.NewService(localidadeRepo, geoService)

	// CEP handler para consulta direta por CEP
	cepHandler := cephandler.NewHandler(geoService)

	// Relatório (espelho de ponto) - após bancoHorasService e justificativaRepo
	relatorioService := relatorio.NewServiceWithCache(pontoRepo, usuarioRepo, logBancoHorasRepo, justificativaRepo, bancoHorasService, cacheService)
	relatorioHandler := relatorio.NewHandler(relatorioService, funcoesService)

	usuarioHandler := usuario.NewUsuarioHandlerWithBancoHoras(usuarioService, bancoHorasService, funcoesService)
	authHandler := auth.NewAuthHandler(authService)
	pontoHandler := ponto.NewPontoHandler(pontoService, justificativaService, bancoHorasService, funcoesService)

	empresaHandler := empresa.NewEmpresaHandler(empresaService, funcoesService, db)
	cargoHandler := cargo.NewCargoHandler(cargoService, funcoesService)
	permissaoHandler := permissao.NewHandler(permissaoService)
	bancoHorasHandler := bancohoras.NewBancoHorasHandler(bancoHorasService, usuarioService, funcoesService)
	justificativaHandler := justificativa.NewHandler(justificativaService, funcoesService)
	localidadeHandler := localidade.NewHandler(localidadeService, funcoesService)

	// --- Middlewares ---
	authMiddleware := auth.AuthMiddleware(jwtService)

	// Criamos os nossos middlewares de permissão aqui.
	// Cada um verifica uma permissão específica.
	canEditEmpresa := auth.PermissionMiddleware(usuarioService, funcoesService, permissions.EDITAR_EMPRESA)
	canDeleteEmpresa := auth.PermissionMiddleware(usuarioService, funcoesService, permissions.DELETAR_EMPRESA)
	canEditUsuario := auth.PermissionMiddleware(usuarioService, funcoesService, permissions.EDITAR_USUARIO)
	canDeleteUsuario := auth.PermissionMiddleware(usuarioService, funcoesService, permissions.DELETAR_USUARIO)
	canManageCargos := auth.PermissionMiddleware(usuarioService, funcoesService, permissions.GERENCIAR_CARGOS)
	canViewSaldo := auth.PermissionMiddleware(usuarioService, funcoesService, permissions.VER_SALDO_FUNCIONARIOS)
	canEditSaldo := auth.PermissionMiddleware(usuarioService, funcoesService, permissions.EDITAR_SALDO_FUNCIONARIOS)
	canViewPonto := auth.PermissionMiddleware(usuarioService, funcoesService, permissions.VISUALIZAR_PONTO_FUNCIONARIOS)
	canAdjustPonto := auth.PermissionMiddleware(usuarioService, funcoesService, permissions.AJUSTAR_PONTO_FUNCIONARIOS)
	canManageJustificativas := auth.PermissionMiddleware(usuarioService, funcoesService, permissions.GERENCIAR_JUSTIFICATIVAS)
	canManageLocalidades := auth.PermissionMiddleware(usuarioService, funcoesService, permissions.GERENCIAR_LOCALIDADES)
	canViewRelatoriosGerais := auth.PermissionMiddleware(usuarioService, funcoesService, permissions.VISUALIZAR_RELATORIOS_GERAIS)

	// Scheduler controlado por variável de ambiente ENABLE_SCHEDULER=true
	if enable, err := strconv.ParseBool(os.Getenv("ENABLE_SCHEDULER")); err == nil && enable {
		log.Println("Scheduler habilitado via ENABLE_SCHEDULER=true")
		scheduler := scheduler.NewScheduler(bancoHorasService, usuarioService)
		scheduler.Start()
	} else {
		log.Println("Scheduler desabilitado (defina ENABLE_SCHEDULER=true para ativar)")
	}

	// --- Rotas da API ---
	router := gin.Default()

	router.SetTrustedProxies(nil)

	// --- CONFIGURAÇÃO DE CORS OTIMIZADA E SEGURA ---
	configCORS := cors.DefaultConfig()
	configCORS.AllowCredentials = true

	// A nossa nova função de verificação de origem
	configCORS.AllowOriginFunc = func(origin string) bool {
		// Primeiro, verifica se a origem está na nossa lista de domínios permitidos.
		if allowedOrigins[origin] {
			return true
		}
		// Se não estiver, verifica se corresponde ao padrão do GitHub Codespaces.
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
		// Endpoint público para testar consulta de CEP
		apiV1.GET("/cep/v2/:cep", cepHandler.GetByCEPV2)

		apiV1.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "UP"})
		})
		// Rotas Públicas
		apiV1.POST("/auth/signup", authHandler.SignUp)
		apiV1.POST("/auth/login", authHandler.Login)

		// Rotas para Super-Admin (no futuro, proteger com um middleware de "SuperAdmin")
		apiV1.POST("/permissoes", permissaoHandler.Create)
		apiV1.GET("/permissoes", permissaoHandler.FindAll)

		apiV1.POST("/empresas", empresaHandler.CriarEmpresaHandler)
		apiV1.GET("/empresas/:id", empresaHandler.GetEmpresaByIDHandler)

		// Rota para tarefas internas, a ser chamada pelo Cloud Scheduler
		apiV1.POST("/tasks/fechamento-diario", bancoHorasHandler.ExecutarFechamentoDiario)

		// Rotas Protegidas (requerem login básico)
		rotasProtegidas := apiV1.Group("")
		rotasProtegidas.Use(authMiddleware)
		{
			// Rotas de Usuário
			// Criação de usuário precisa estar autenticada para capturar o id do requisitante do token
			// Apenas usuários com permissão EDITAR_USUARIO podem criar novos usuários
			rotasProtegidas.POST("/usuarios", canEditUsuario, usuarioHandler.CriarUsuarioHandler)
			rotasProtegidas.GET("/usuarios", usuarioHandler.GetAllUsuariosHandler)
			rotasProtegidas.GET("/usuarios/:id", usuarioHandler.GetByIdHandler)
			// Update e Delete já têm validação de permissão interna no handler
			rotasProtegidas.PUT("/usuarios/:id", usuarioHandler.UpdateUsuarioHandler)
			rotasProtegidas.GET("/usuarios/me", usuarioHandler.GetMeuPerfil)

			// Agora, para apagar um utilizador, é preciso a permissão DELETAR_USUARIO
			rotasProtegidas.DELETE("/usuarios/:id", canDeleteUsuario, usuarioHandler.DeleteHandler)

			// Gestão de Cargos (criar, listar, atualizar, deletar)
			rotasProtegidas.POST("/cargos", canManageCargos, cargoHandler.CreateCargo)
			rotasProtegidas.GET("/cargos", cargoHandler.GetAllCargos)
			rotasProtegidas.GET("/cargos/:id", cargoHandler.GetCargoByID)
			rotasProtegidas.PUT("/cargos/:id", canManageCargos, cargoHandler.UpdateCargo)
			rotasProtegidas.DELETE("/cargos/:id", canManageCargos, cargoHandler.DeleteCargo)
			rotasProtegidas.POST("/cargos/:id/permissoes/:permissaoId", canManageCargos, cargoHandler.AddPermissionToCargo)

			// Rota de Ponto
			rotasProtegidas.POST("/pontos", pontoHandler.BaterPonto)
			rotasProtegidas.GET("/pontos/meus-registros", pontoHandler.GetMeusRegistos)
			rotasProtegidas.GET("/pontos/usuario/:id", canViewPonto, pontoHandler.GetRegistosPorUsuarioID)
			rotasProtegidas.POST("/pontos/ajuste", canAdjustPonto, pontoHandler.AjustarPonto)
			rotasProtegidas.PUT("/pontos/:pontoId", canAdjustPonto, pontoHandler.EditarPonto)

			// Rotas de Relatórios de Ponto (exportação CSV/PDF)
			rotasProtegidas.GET("/relatorios/ponto/meus-registros/export", pontoHandler.ExportarRelatorio)
			rotasProtegidas.GET("/relatorios/ponto/usuario/:id/export", canViewPonto, pontoHandler.ExportarRelatorio)

			// Espelho de Ponto
			rotasProtegidas.GET("/relatorios/ponto/espelho/me", relatorioHandler.GetEspelhoMe)
			rotasProtegidas.GET("/relatorios/ponto/espelho/usuario/:id", canViewPonto, relatorioHandler.GetEspelhoUsuario)

			// Relatório Geral de Ponto
			rotasProtegidas.GET("/relatorios/geral", canViewRelatoriosGerais, relatorioHandler.GerarRelatorioGeral)
			rotasProtegidas.GET("/relatorios/geral/export", canViewRelatoriosGerais, relatorioHandler.ExportarRelatorioGeral)

			// Rotas de Empresa (Ações gerais)
			rotasProtegidas.GET("/empresas", empresaHandler.GetAllEmpresasHandler)

			// Rotas de Empresa (Ações Administrativas, protegidas por permissão)
			rotasProtegidas.PUT("/empresas/:id", canEditEmpresa, empresaHandler.UpdateEmpresaHandler)
			rotasProtegidas.DELETE("/empresas/:id", canDeleteEmpresa, empresaHandler.DeleteEmpresaHandler)

			rotasProtegidas.GET("/bancohoras/saldo/usuario/:id", bancoHorasHandler.GetSaldoDoDia)
			rotasProtegidas.POST("/bancohoras/fechamento/usuario/:id", canEditSaldo, bancoHorasHandler.FecharDia)
			// Dashboard de banco de horas do usuário autenticado
			rotasProtegidas.GET("/bancohoras/dashboard/me", bancoHorasHandler.GetDashboard)
			// Saldo de banco de horas de um usuário específico (requer permissão VER_SALDO_FUNCIONARIOS)
			rotasProtegidas.GET("/bancohoras/dashboard/:userId", canViewSaldo, bancoHorasHandler.GetSaldoUsuario)

			// --- NOVAS ROTAS DE JUSTIFICATIVAS ---
			// Rota para o funcionário criar uma solicitação
			rotasProtegidas.POST("/justificativas", justificativaHandler.SolicitarAjuste)

			// Rotas para o admin/gestor gerir as solicitações
			rotasProtegidas.GET("/justificativas/pendentes", canManageJustificativas, justificativaHandler.ListarPendentes)
			rotasProtegidas.POST("/justificativas/:id/processar", canManageJustificativas, justificativaHandler.AprovarReprovar)

			//rotas de localodade
			rotasProtegidas.POST("/localidades", canManageLocalidades, localidadeHandler.Create)
			rotasProtegidas.GET("/empresas/:id/localidades", canManageLocalidades, localidadeHandler.GetAllByEmpresa)
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
