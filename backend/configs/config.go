package configs

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Config representa a configuração da aplicação
type Config struct {
	Server     ServerConfig
	Database   DatabaseConfig
	JWT        JWTConfig
	Encryption EncryptionConfig
	Audit      AuditConfig
	Providers  map[string]ProviderConfig
}

// ServerConfig configurações do servidor
type ServerConfig struct {
	Port            int
	Environment     string
	AllowedOrigins  []string
	RateLimitRPS    int
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
}

// DatabaseConfig configurações do banco de dados
type DatabaseConfig struct {
	Host            string
	Port            int
	User            string
	Password        string
	Database        string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// JWTConfig configurações JWT
type JWTConfig struct {
	SecretKey       string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

// EncryptionConfig configurações de criptografia
type EncryptionConfig struct {
	Key string // Base64 encoded 32-byte key
}

// AuditConfig configurações de auditoria
type AuditConfig struct {
	Enabled        bool
	RetentionYears int
	AsyncLogging   bool
}

// ProviderConfig configurações de providers
type ProviderConfig struct {
	BaseURL      string
	AuthURL      string
	SandboxURL   string
	Timeout      int
	MaxRetries   int
	RequiresMTLS bool
}

// LoadConfig carrega a configuração da aplicação
func LoadConfig(configPath string) (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(configPath)
	viper.AddConfigPath(".")
	viper.AddConfigPath("./configs")

	// Configurações padrão
	setDefaults()

	// Ler variáveis de ambiente
	viper.AutomaticEnv()

	// Ler arquivo de configuração
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("erro ao ler arquivo de configuração: %w", err)
		}
	}

	config := &Config{}

	// Server
	config.Server = ServerConfig{
		Port:            viper.GetInt("server.port"),
		Environment:     viper.GetString("server.environment"),
		AllowedOrigins:  viper.GetStringSlice("server.allowed_origins"),
		RateLimitRPS:    viper.GetInt("server.rate_limit_rps"),
		ReadTimeout:     viper.GetDuration("server.read_timeout"),
		WriteTimeout:    viper.GetDuration("server.write_timeout"),
		ShutdownTimeout: viper.GetDuration("server.shutdown_timeout"),
	}

	// Database
	config.Database = DatabaseConfig{
		Host:            viper.GetString("database.host"),
		Port:            viper.GetInt("database.port"),
		User:            viper.GetString("database.user"),
		Password:        viper.GetString("database.password"),
		Database:        viper.GetString("database.database"),
		SSLMode:         viper.GetString("database.sslmode"),
		MaxOpenConns:    viper.GetInt("database.max_open_conns"),
		MaxIdleConns:    viper.GetInt("database.max_idle_conns"),
		ConnMaxLifetime: viper.GetDuration("database.conn_max_lifetime"),
	}

	// JWT
	config.JWT = JWTConfig{
		SecretKey:       viper.GetString("jwt.secret_key"),
		AccessTokenTTL:  viper.GetDuration("jwt.access_token_ttl"),
		RefreshTokenTTL: viper.GetDuration("jwt.refresh_token_ttl"),
	}

	// Encryption
	config.Encryption = EncryptionConfig{
		Key: viper.GetString("encryption.key"),
	}

	// Audit
	config.Audit = AuditConfig{
		Enabled:        viper.GetBool("audit.enabled"),
		RetentionYears: viper.GetInt("audit.retention_years"),
		AsyncLogging:   viper.GetBool("audit.async_logging"),
	}

	// Providers
	config.Providers = make(map[string]ProviderConfig)
	providersMap := viper.GetStringMap("providers")
	for providerName := range providersMap {
		prefix := fmt.Sprintf("providers.%s", providerName)
		config.Providers[providerName] = ProviderConfig{
			BaseURL:      viper.GetString(prefix + ".base_url"),
			AuthURL:      viper.GetString(prefix + ".auth_url"),
			SandboxURL:   viper.GetString(prefix + ".sandbox_url"),
			Timeout:      viper.GetInt(prefix + ".timeout"),
			MaxRetries:   viper.GetInt(prefix + ".max_retries"),
			RequiresMTLS: viper.GetBool(prefix + ".requires_mtls"),
		}
	}

	return config, nil
}

func setDefaults() {
	// Server defaults
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.environment", "development")
	viper.SetDefault("server.allowed_origins", []string{"*"})
	viper.SetDefault("server.rate_limit_rps", 100)
	viper.SetDefault("server.read_timeout", 30*time.Second)
	viper.SetDefault("server.write_timeout", 30*time.Second)
	viper.SetDefault("server.shutdown_timeout", 10*time.Second)

	// Database defaults
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.user", "postgres")
	viper.SetDefault("database.database", "pixsaas")
	viper.SetDefault("database.sslmode", "disable")
	viper.SetDefault("database.max_open_conns", 25)
	viper.SetDefault("database.max_idle_conns", 5)
	viper.SetDefault("database.conn_max_lifetime", 5*time.Minute)

	// JWT defaults
	viper.SetDefault("jwt.access_token_ttl", 15*time.Minute)
	viper.SetDefault("jwt.refresh_token_ttl", 7*24*time.Hour)

	// Audit defaults
	viper.SetDefault("audit.enabled", true)
	viper.SetDefault("audit.retention_years", 5)
	viper.SetDefault("audit.async_logging", true)
}

// GetDSN retorna a string de conexão do banco de dados
func (c *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Database, c.SSLMode,
	)
}

// IsDevelopment verifica se está em ambiente de desenvolvimento
func (c *ServerConfig) IsDevelopment() bool {
	return c.Environment == "development"
}

// IsProduction verifica se está em ambiente de produção
func (c *ServerConfig) IsProduction() bool {
	return c.Environment == "production"
}
