package main

import (
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/Loviiin/ponto-api-go/docs"
	"github.com/Loviiin/ponto-api-go/internal/config"
	"github.com/Loviiin/ponto-api-go/internal/domain/auth"
	"github.com/Loviiin/ponto-api-go/internal/domain/bancohoras"
	"github.com/Loviiin/ponto-api-go/internal/domain/cargo"
	cephandler "github.com/Loviiin/ponto-api-go/internal/domain/cep"
	"github.com/Loviiin/ponto-api-go/internal/domain/contrato"
	"github.com/Loviiin/ponto-api-go/internal/domain/empresa"
	"github.com/Loviiin/ponto-api-go/internal/domain/justificativa"
	"github.com/Loviiin/ponto-api-go/internal/domain/localidade"
	"github.com/Loviiin/ponto-api-go/internal/domain/logbancohoras"
	"github.com/Loviiin/ponto-api-go/internal/domain/permissao"
	"github.com/Loviiin/ponto-api-go/internal/domain/ponto"
	"github.com/Loviiin/ponto-api-go/internal/domain/profile"
	"github.com/Loviiin/ponto-api-go/internal/domain/relatorio"
	"github.com/Loviiin/ponto-api-go/internal/domain/usuario"
	"github.com/Loviiin/ponto-api-go/pkg/audit"
	"github.com/Loviiin/ponto-api-go/pkg/brasilapi"
	"github.com/Loviiin/ponto-api-go/pkg/cep"
	"github.com/Loviiin/ponto-api-go/pkg/cloudinary"
	"github.com/Loviiin/ponto-api-go/pkg/distancematrix"
	"github.com/Loviiin/ponto-api-go/pkg/funcoes"
	"github.com/Loviiin/ponto-api-go/pkg/geolocation"
	"github.com/Loviiin/ponto-api-go/pkg/jwt"
	"github.com/Loviiin/ponto-api-go/pkg/logger"
	"github.com/Loviiin/ponto-api-go/pkg/permissions"
	"github.com/Loviiin/ponto-api-go/pkg/scheduler"
	"github.com/Loviiin/ponto-api-go/pkg/viacep"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

)


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
	logger.InitLogger()
	funcoesService := funcoes.NewFuncoes()

	//Carrega as configurações de ambiente
	cfg := CarregaConfig()

	//Conecta ao banco de dados
	BancoDeDados := conectaBD(cfg)

	// Inicializa o serviço de cache: preferir Upstash REST se variáveis estiverem presentes
	cacheService := inicializaCache(cfg)

	// Verifica se deve resetar o banco de dados
	resetaBanco(BancoDeDados)

	resetaCacheComBD(cacheService)

	criaSchemaDatabase(BancoDeDados)

	// Popula permissões padrão e super-admin
	config.SeedPermissions(BancoDeDados)
	config.SeedSuperAdmin(BancoDeDados)
	if err := config.EnsureDemoWorkspace(BancoDeDados); err != nil {
		log.Printf("AVISO: não foi possível preparar o demo: %v", err)
	}


	// Inicializar Serviço de Email
	emailService := inicializaEmailService(cfg)

	usuarioRepo := usuario.NewUsuarioRepository(BancoDeDados, cacheService)
	passwordResetRepo := auth.NewPasswordResetRepository(BancoDeDados)
	pontoRepo := ponto.NewPontoRepository(BancoDeDados)
	empresaRepo := empresa.NewEmpresaRepository(BancoDeDados, cacheService)
	cargoRepo := cargo.NewCargoRepository(BancoDeDados, cacheService)
	permissaoRepo := permissao.NewRepository(BancoDeDados, cacheService)
	justificativaRepo := justificativa.NewRepository(BancoDeDados, cacheService)
	logBancoHorasRepo := logbancohoras.NewRepository(BancoDeDados)
	contratoRepo := contrato.NewContratoRepository(BancoDeDados)
	localidadeRepo := localidade.NewRepository(BancoDeDados)

	jwtService := jwt.NewJWTService(cfg.JWTSecretKey, "ponto-api-go")

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

	usuarioService := usuario.NewUsuarioService(BancoDeDados, usuarioRepo, cargoRepo, empresaRepo, contratoRepo, localidadeRepo)
	// Passando emailService para AuthService
	authService := auth.NewAuthService(usuarioRepo, empresaRepo, cargoRepo, contratoRepo, localidadeRepo, passwordResetRepo, geoService, jwtService, cacheService, BancoDeDados, emailService)
	pontoService := ponto.NewPontoServiceWithCache(pontoRepo, usuarioRepo, localidadeRepo, BancoDeDados, cacheService)

	empresaService := empresa.NewEmpresaService(empresaRepo)
	cargoService := cargo.NewCargoService(cargoRepo)
	permissaoService := permissao.NewService(permissaoRepo)
	bancoHorasService := bancohoras.NewBancoHorasServiceWithCache(pontoRepo, usuarioRepo, logBancoHorasRepo, BancoDeDados, cacheService)
	justificativaService := justificativa.NewService(justificativaRepo, pontoRepo, BancoDeDados)
	localidadeService := localidade.NewService(localidadeRepo, geoService, cacheService)

	// CEP handler para consulta direta por CEP
	cepHandler := cephandler.NewHandler(geoService)

	// Relatório (espelho de ponto) - após bancoHorasService e justificativaRepo
	relatorioService := relatorio.NewServiceWithCache(pontoRepo, usuarioRepo, logBancoHorasRepo, justificativaRepo, bancoHorasService, cacheService)
	relatorioHandler := relatorio.NewHandler(relatorioService, funcoesService)

	usuarioHandler := usuario.NewUsuarioHandlerWithBancoHoras(usuarioService, bancoHorasService, funcoesService)
	authHandler := auth.NewAuthHandler(authService)
	pontoHandler := ponto.NewPontoHandler(pontoService, justificativaService, bancoHorasService, funcoesService)

	empresaHandler := empresa.NewEmpresaHandler(empresaService, funcoesService, BancoDeDados)
	cargoHandler := cargo.NewCargoHandler(cargoService, funcoesService)
	permissaoHandler := permissao.NewHandler(permissaoService)
	bancoHorasHandler := bancohoras.NewBancoHorasHandler(bancoHorasService, usuarioService, funcoesService)
	justificativaHandler := justificativa.NewHandler(justificativaService, funcoesService)
	localidadeHandler := localidade.NewHandler(localidadeService, funcoesService)

	// Cloudinary Service - Upload de avatares
	cloudinaryService, err := cloudinary.NewService(cfg.CloudinaryURL)
	if err != nil {
		log.Printf("AVISO: Cloudinary não inicializado: %v (upload de avatares desabilitado)", err)
	}

	// Audit Logger - Serviço de auditoria
	auditLogger := audit.NewService(BancoDeDados)

	// Profile - Handler de perfil de usuário
	profileRepo := profile.NewRepository(BancoDeDados, cacheService)
	profileService := profile.NewService(profileRepo, BancoDeDados, cacheService, auditLogger)
	profileHandler := profile.NewHandler(profileService, cloudinaryService)

	// --- Middlewares ---
	authMiddleware := auth.AuthMiddleware(jwtService)

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
	canCreateJustificativa := auth.PermissionMiddleware(usuarioService, funcoesService, permissions.CRIAR_JUSTIFICATIVA_PROPRIA)
	canAprovarJustificativas := auth.PermissionMiddleware(usuarioService, funcoesService, permissions.APROVAR_JUSTIFICATIVAS)
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
		apiV1.POST("/auth/demo", authHandler.Demo)
		apiV1.POST("/auth/signup", authHandler.SignUp)
		apiV1.POST("/auth/login", authHandler.Login)

		// Password Reset
		apiV1.POST("/auth/forgot-password", authHandler.RequestPasswordReset)
		apiV1.POST("/auth/reset-password", authHandler.ResetPassword)
		apiV1.GET("/auth/validate-reset-token", authHandler.ValidateResetToken)

		// Google OAuth (Login)
		// apiV1.GET("/auth/google/login", authHandler.GoogleLogin)
		// apiV1.GET("/auth/google/callback", authHandler.GoogleCallback)

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
			rotasProtegidas.PATCH("/usuarios/:id", usuarioHandler.PatchUsuarioHandler) // Atualização parcial com suporte a cargo
			rotasProtegidas.GET("/usuarios/me", usuarioHandler.GetMeuPerfil)           // Agora, para apagar um utilizador, é preciso a permissão DELETAR_USUARIO
			rotasProtegidas.DELETE("/usuarios/:id", canDeleteUsuario, usuarioHandler.DeleteHandler)

			// Gestão de Cargos (criar, listar, atualizar, deletar)
			rotasProtegidas.POST("/cargos", canManageCargos, cargoHandler.CreateCargo)
			rotasProtegidas.GET("/cargos", cargoHandler.GetAllCargos)
			rotasProtegidas.GET("/cargos/:id", cargoHandler.GetCargoByID)
			rotasProtegidas.PUT("/cargos/:id", canManageCargos, cargoHandler.UpdateCargo)
			rotasProtegidas.DELETE("/cargos/:id", canManageCargos, cargoHandler.DeleteCargo)

			// Gestão de Permissões de Cargos
			rotasProtegidas.GET("/cargos/:id/permissoes", cargoHandler.GetPermissionsByCargo)
			rotasProtegidas.POST("/cargos/:id/permissoes/:permissaoId", canManageCargos, cargoHandler.AddPermissionToCargo)
			rotasProtegidas.DELETE("/cargos/:id/permissoes/:permissaoId", canManageCargos, cargoHandler.RemovePermissionFromCargo)

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
			// Rota para o funcionário criar uma solicitação de ponto faltante (requer permissão CRIAR_JUSTIFICATIVA_PROPRIA)
			rotasProtegidas.POST("/justificativas", canCreateJustificativa, justificativaHandler.SolicitarAjuste)
			// Rota para o funcionário solicitar correção de ponto existente (requer permissão CRIAR_JUSTIFICATIVA_PROPRIA)
			rotasProtegidas.POST("/justificativas/solicitar-correcao", canCreateJustificativa, justificativaHandler.SolicitarCorrecaoPonto)
			// Rota para o funcionário ver suas próprias justificativas (não requer permissão especial, apenas autenticação)
			rotasProtegidas.GET("/justificativas/minhas", justificativaHandler.ListarMinhas)
			// Rota para o funcionário cancelar sua própria solicitação pendente (não requer permissão especial)
			rotasProtegidas.DELETE("/justificativas/:id/cancelar", justificativaHandler.CancelarSolicitacao)

			// Rotas para o admin/gestor gerir as solicitações (requer permissão APROVAR_JUSTIFICATIVAS)
			rotasProtegidas.GET("/justificativas/pendentes", canAprovarJustificativas, justificativaHandler.ListarPendentes)
			rotasProtegidas.POST("/justificativas/:id/processar", canAprovarJustificativas, justificativaHandler.AprovarReprovar)

			//rotas de localidade
			rotasProtegidas.POST("/localidades", canManageLocalidades, localidadeHandler.Create)
			rotasProtegidas.GET("/localidades", canManageLocalidades, localidadeHandler.ListarLocalidades)
			rotasProtegidas.GET("/empresas/:id/localidades", canManageLocalidades, localidadeHandler.GetAllByEmpresa)

			// --- ROTAS DE PERFIL ---
			// Perfil do usuário autenticado
			rotasProtegidas.GET("/profile/me", profileHandler.GetMyProfile)
			rotasProtegidas.PATCH("/profile/me", profileHandler.UpdateProfile)           // PATCH para updates parciais
			rotasProtegidas.PATCH("/profile/me/password", profileHandler.ChangePassword) // PATCH semântico
			// rotasProtegidas.POST("/profile/me/avatar", profileHandler.UploadAvatar) // Removed as field is no longer in model

			// Estatísticas e permissões
			rotasProtegidas.GET("/profile/me/stats", profileHandler.GetMyStats)
			rotasProtegidas.GET("/profile/me/permissions", profileHandler.GetMyPermissions)

			// Calendário e atividades recentes
			rotasProtegidas.GET("/profile/me/calendar", profileHandler.GetCalendar)
			rotasProtegidas.GET("/profile/me/recent-activity", profileHandler.GetRecentActivity)

			// --- ROTAS DE ADMIN ---
			// Atualizar CPF (requer permissão EDITAR_USUARIO)
			rotasProtegidas.PATCH("/admin/users/:user_id/cpf", canEditUsuario, profileHandler.UpdateCPF)

			// --- Google OAuth (Vincular Conta) ---
			// rotasProtegidas.GET("/auth/google/link", authHandler.LinkGoogleAccount)
			// rotasProtegidas.POST("/auth/google/link/callback", authHandler.LinkGoogleCallback)
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

