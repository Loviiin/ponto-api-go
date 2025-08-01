package config

import (
	// Viper é a biblioteca que vamos usar para ler o arquivo .env
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

	// Chave secreta para assinar os tokens JWT (usaremos mais tarde)
	JWTSecretKey string `mapstructure:"JWT_SECRET_KEY"`
}

// LoadConfig é a função que lê as configurações do arquivo .env no caminho especificado.
func LoadConfig(path string) (config Config, err error) {
	// Diz ao Viper para procurar por arquivos de configuração no caminho fornecido.
	// O "." significa o diretório atual.
	viper.AddConfigPath(path)

	// Define o nome do arquivo de configuração (sem a extensão).
	viper.SetConfigName(".env")

	// Define o tipo do arquivo de configuração.
	viper.SetConfigType("env")

	// viper.AutomaticEnv() permite que o Viper também leia variáveis
	// do ambiente do sistema, que podem sobrescrever as do arquivo .env.
	viper.AutomaticEnv()

	// Tenta ler o arquivo de configuração.
	err = viper.ReadInConfig()
	if err != nil {
		// Se houver um erro na leitura (ex: arquivo não encontrado), a função retorna o erro.
		return
	}

	// "Deserializa" os valores lidos do arquivo para dentro da nossa struct 'config'.
	err = viper.Unmarshal(&config)

	// Retorna a struct preenchida e um erro (que será 'nil' se tudo deu certo).
	return
}
