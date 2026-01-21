package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"regexp"

	"github.com/Loviiin/ponto-api-go/internal/config"
	"github.com/Loviiin/ponto-api-go/internal/constants"
	"github.com/Loviiin/ponto-api-go/internal/model"
	"github.com/Loviiin/ponto-api-go/pkg/cache"
	"github.com/Loviiin/ponto-api-go/pkg/email"
	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Regex para permitir todos os subdomínios de desenvolvimento do GitHub Codespaces
var allowedOriginRegex = regexp.MustCompile(`^https?:\/\/.*\.app\.github\.dev$`)

// --- CONFIGURAÇÃO DE CORS ---
var allowedOrigins = map[string]bool{
	"http://localhost:3000/":                                           true,
	"http://localhost:3000":                                            true,
	"https://nexora-app.vercel.app":                                    true,
	"https://meu-ponto-frontend.vercel.app":                            true,
	"https://meu-ponto-frontend-git-main-loviins-projects.vercel.app":  true,
	"https://meu-ponto-frontend-n9lx9odbj-loviins-projects.vercel.app": true,
}

func CarregaConfig() *config.Config {
	cfg, err := config.LoadConfig(".")
	if err != nil {
		log.Fatal("Não foi possível carregar as configurações: ", err)
	}
	log.Printf("Configurações carregadas: %+v\n", cfg)
	return &cfg
}

func conectaBD(cfg *config.Config) *gorm.DB {
	dbconf := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=%s",
		cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode, constants.TimezoneBR)

	db, err := gorm.Open(postgres.Open(dbconf), &gorm.Config{})
	if err != nil {
		log.Fatal("Não foi possivel Conectar ao banco de dados: ", err)
	}
	log.Println("Conexão com o banco de dados estabelecida com sucesso.")
	return db
}

func resetaBanco(db *gorm.DB) {
	resetVal := viper.GetString("RESET_DB_ON_START")
	if resetVal == "" {
		resetVal = os.Getenv("RESET_DB_ON_START")
	}

	if resetVal == "true" {
		log.Println("RESET_DB_ON_START está definido como true. Resetando o banco de dados...")
		resetAndSeedDatabase(db)
	} else {
		log.Printf("RESET_DB_ON_START='%s'. Pulando reset do banco de dados.", resetVal)
	}
}

func inicializaCache(cfg *config.Config) cache.Service {
	if cs, err := initUpstashCache(cfg); err == nil {
		log.Println("Cache: Usando Upstash REST API")
		return cs
	} else if cfg.UpstashRedisRestURL != "" {
		log.Printf("Aviso: Falha ao conectar Upstash (%v). Tentando Redis local...", err)
	}

	cs, err := initRedisCache(cfg)
	if err != nil {
		log.Panicf("CRÍTICO: Falha ao conectar em qualquer cache (Redis/Upstash): %v", err)
	}

	log.Println("Cache: Usando Redis Nativo")
	return cs
}

func initUpstashCache(cfg *config.Config) (cache.Service, error) {
	if cfg.UpstashRedisRestURL == "" || cfg.UpstashRedisRestToken == "" {
		return nil, fmt.Errorf("credenciais do Upstash não configuradas")
	}
	return cache.NewUpstashService(cfg.UpstashRedisRestURL, cfg.UpstashRedisRestToken)
}

func initRedisCache(cfg *config.Config) (cache.Service, error) {
	return cache.NewRedisService(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
}

func resetaCacheComBD(cacheService cache.Service) {
	flusher := cacheService.(cache.Flusher)
	if os.Getenv("RESET_DB_ON_START") == "true" {
		if err := flusher.FlushAll(context.Background()); err != nil {
			log.Printf("[cache] falha ao limpar cache após reset BancoDeDados: %v", err)
		} else {
			log.Println("[cache] cache limpo após reset BancoDeDados")
		}
	}
}

func criaSchemaDatabase(db *gorm.DB) {
	err := db.AutoMigrate(
		&model.Usuario{},
		&model.PasswordResetToken{},
		&model.RegistroPonto{},
		&model.Empresa{},
		&model.Cargo{},
		&model.Permissao{},
		&model.Justificativa{},
		&model.LogBancoHoras{},
		&model.Contrato{},
		&model.Localidade{},
		&model.AuditLog{},
	)
	if err != nil {
		log.Fatal("Falha ao rodar a migração: ", err)
	}
	log.Println("Migração do banco de dados executada com sucesso.")
}

func inicializaEmailService(cfg *config.Config) *email.EmailService {
	svc := email.NewEmailService(
		cfg.SMTPHost,
		cfg.SMTPPort,
		cfg.SMTPUser,
		cfg.SMTPPassword,
		cfg.SMTPFromName,
		cfg.SMTPFromEmail,
		cfg.FrontendURL,
	)

	if cfg.SMTPHost == "" || cfg.SMTPPort == "0" {
		slog.Warn("Configurações de SMTP ausentes. O serviço de email será desativado.")
		return nil
	}

	if svc == nil {
		slog.Warn("Serviço de email não configurado (verifique .env ou configurações SMTP)")
		return nil
	}

	return svc
}

func resetAndSeedDatabase(db *gorm.DB) {
	log.Println("Iniciando reset do banco de dados...")

	err := db.Migrator().DropTable(
		"usuario_cargos",   // Tabela de junção para Usuario e Cargo
		"cargo_permissoes", // Tabela de junção para Cargo e Permissao
		&model.AuditLog{},  // Precisa ser dropado antes de Usuario
		&model.RegistroPonto{},
		&model.Justificativa{},
		&model.LogBancoHoras{},
		&model.Contrato{},
		&model.PasswordResetToken{}, // NOVO
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

	// Recria as tabelas (ordem importa: dependências primeiro!)
	log.Println("Recriando tabelas com AutoMigrate...")
	err = db.AutoMigrate(
		&model.Empresa{},
		&model.Permissao{},
		&model.Cargo{},
		&model.Localidade{},
		&model.Usuario{},
		&model.PasswordResetToken{},
		&model.Contrato{},
		&model.AuditLog{},
		&model.RegistroPonto{},
		&model.Justificativa{},
		&model.LogBancoHoras{},
	)
	if err != nil {
		log.Fatal("Falha ao rodar a migração: ", err)
	}
	log.Println("Tabelas recriadas com sucesso.")

	log.Println("Populando o banco de dados com dados iniciais...")
	config.SeedPermissions(db)
	config.SeedSuperAdmin(db)
	log.Println("Banco de dados resetado e populado com sucesso!")
}
