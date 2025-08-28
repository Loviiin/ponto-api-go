// Em internal/config/config.go

package config

import (
	// Viper é a biblioteca que vamos usar para ler o arquivo .env
	"github.com/spf13/viper"
	"strings" // <-- NOVO IMPORT
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
	err = viper.Unmarshal(&config)
	.
	return
}