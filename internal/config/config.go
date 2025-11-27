// Em internal/config/config.go

package config

import (
	// Viper é a biblioteca que vamos usar para ler o arquivo .env
	"strings" // <-- NOVO IMPORT

	"github.com/spf13/viper"
)

// Config é a struct que vai armazenar todas as configurações da nossa aplicação.
// As 'tags' `mapstructure` dizem ao Viper qual variável de ambiente
// corresponde a qual campo da struct.
type Config struct {
	// Porta onde a API vai rodar
	APIPort string `mapstructure:"API_PORT"`

	// Configurações de conexão com o banco de dados PostgreSQL
	DBHost     string `mapstructure:"DB_HOST"`
	DBPort     string `mapstructure:"DB_PORT"`
	DBUser     string `mapstructure:"DB_USER"`
	DBPassword string `mapstructure:"DB_PASSWORD"`
	DBName     string `mapstructure:"DB_NAME"`
	// NOVO CAMPO para o modo SSL
	DBSSLMode string `mapstructure:"DB_SSLMODE"`

	// Chave secreta para assinar os tokens JWT (usaremos mais tarde)
	JWTSecretKey string `mapstructure:"JWT_SECRET_KEY"`
	// Chave da API do Distance Matrix AI para geocodificação
	DistanceMatrixAPIKey string `mapstructure:"DISTANCEMATRIX_API_KEY"`
	BrasilApiUrl         string `mapstructure:"BRASILAPI_URL"`

	// Cache: Upstash REST
	UpstashRedisRestURL   string `mapstructure:"UPSTASH_REDIS_REST_URL"`
	UpstashRedisRestToken string `mapstructure:"UPSTASH_REDIS_REST_TOKEN"`

	// Cache: Redis nativo
	RedisAddr     string `mapstructure:"REDIS_ADDR"`
	RedisPassword string `mapstructure:"REDIS_PASSWORD"`
	RedisDB       int    `mapstructure:"REDIS_DB"`

	// Cloudinary para upload de avatares
	CloudinaryURL string `mapstructure:"CLOUDINARY_URL"`

	// SMTP Configuration
	SMTPHost      string `mapstructure:"SMTP_HOST"`
	SMTPPort      string `mapstructure:"SMTP_PORT"`
	SMTPUser      string `mapstructure:"SMTP_USER"`
	SMTPPassword  string `mapstructure:"SMTP_PASSWORD"`
	SMTPFromName  string `mapstructure:"SMTP_FROM_NAME"`
	SMTPFromEmail string `mapstructure:"SMTP_FROM_EMAIL"`

	// Google OAuth
	GoogleClientID     string `mapstructure:"GOOGLE_CLIENT_ID"`
	GoogleClientSecret string `mapstructure:"GOOGLE_CLIENT_SECRET"`
	GoogleRedirectURL  string `mapstructure:"GOOGLE_REDIRECT_URL"`

	// Frontend URL
	FrontendURL string `mapstructure:"FRONTEND_URL"`
}

// --- FUNÇÃO LoadConfig COMPLETAMENTE NOVA ---
// LoadConfig lê as configurações. É flexível para ambientes locais e de produção.
func LoadConfig(path string) (config Config, err error) {
	// --- INÍCIO DA CORREÇÃO ---
	// Configura o Viper para ler variáveis de ambiente (ex: API_PORT)
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// Define os valores padrão (opcional, mas boa prática)
	viper.SetDefault("API_PORT", "8083")
	// Definimos 'require' como padrão para o SSL, que é o mais seguro e exigido pelo Neon
	viper.SetDefault("DB_SSLMODE", "require")
	viper.SetDefault("REDIS_DB", 0)

	// Tenta ler o ficheiro .env (para desenvolvimento local)
	// Se não encontrar, não há problema, continuará com as variáveis de ambiente.
	viper.AddConfigPath(path)
	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	_ = viper.ReadInConfig() // Ignora o erro se o ficheiro não for encontrado

	// Faz o "bind" explícito de cada variável. ISTO É O MAIS IMPORTANTE.
	// Isto garante que viper.Unmarshal encontrará os valores.
	err = viper.BindEnv("API_PORT")
	if err != nil {
		return
	}
	err = viper.BindEnv("DB_HOST")
	if err != nil {
		return
	}
	err = viper.BindEnv("DB_PORT")
	if err != nil {
		return
	}
	err = viper.BindEnv("DB_USER")
	if err != nil {
		return
	}
	err = viper.BindEnv("DB_PASSWORD")
	if err != nil {
		return
	}
	err = viper.BindEnv("DB_NAME")
	if err != nil {
		return
	}
	err = viper.BindEnv("DB_SSLMODE")
	if err != nil {
		return
	}
	err = viper.BindEnv("JWT_SECRET_KEY")
	if err != nil {
		return
	}
	// Geocoding provider (Distance Matrix AI)
	if err = viper.BindEnv("DISTANCEMATRIX_API_KEY"); err != nil {
		return
	}
	err = viper.BindEnv("BRASILAPI_URL")
	if err != nil {
		return
	}
	// Cache: Upstash REST
	if err = viper.BindEnv("UPSTASH_REDIS_REST_URL"); err != nil {
		return
	}
	if err = viper.BindEnv("UPSTASH_REDIS_REST_TOKEN"); err != nil {
		return
	}
	// Cache: Redis nativo
	if err = viper.BindEnv("REDIS_ADDR"); err != nil {
		return
	}
	if err = viper.BindEnv("REDIS_PASSWORD"); err != nil {
		return
	}
	if err = viper.BindEnv("REDIS_DB"); err != nil {
		return
	}
	// Cloudinary
	if err = viper.BindEnv("CLOUDINARY_URL"); err != nil {
		return
	}
	// SMTP
	if err = viper.BindEnv("SMTP_HOST"); err != nil {
		return
	}
	if err = viper.BindEnv("SMTP_PORT"); err != nil {
		return
	}
	if err = viper.BindEnv("SMTP_USER"); err != nil {
		return
	}
	if err = viper.BindEnv("SMTP_PASSWORD"); err != nil {
		return
	}
	if err = viper.BindEnv("SMTP_FROM_NAME"); err != nil {
		return
	}
	if err = viper.BindEnv("SMTP_FROM_EMAIL"); err != nil {
		return
	}
	// Google OAuth
	if err = viper.BindEnv("GOOGLE_CLIENT_ID"); err != nil {
		return
	}
	if err = viper.BindEnv("GOOGLE_CLIENT_SECRET"); err != nil {
		return
	}
	if err = viper.BindEnv("GOOGLE_REDIRECT_URL"); err != nil {
		return
	}
	// Frontend URL
	if err = viper.BindEnv("FRONTEND_URL"); err != nil {
		return
	}
	// --- FIM DA CORREÇÃO ---

	// "Deserializa" os valores lidos para dentro da nossa struct 'config'.
	err = viper.Unmarshal(&config)

	// Retorna a struct preenchida e um erro (que será 'nil' se tudo deu certo).
	return
}
