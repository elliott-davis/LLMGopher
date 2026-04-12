package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// Config holds all gateway configuration.
// Precedence: CLI flags > environment variables > config file > defaults.
type Config struct {
	Server    ServerConfig
	Postgres  PostgresConfig
	Redis     RedisConfig
	Auth      AuthConfig
	Gateway   GatewayConfig
	Providers ProvidersConfig
	Security  SecurityConfig
}

type SecurityConfig struct {
	ProviderCredentialsKey string
}

type ProvidersConfig struct {
	OpenAI    ProviderEndpoint
	Anthropic ProviderEndpoint
	Vertex    VertexConfig
	Bedrock   BedrockConfig
}

type ProviderEndpoint struct {
	APIKey  string
	BaseURL string
}

type VertexConfig struct {
	ProjectID string
	Region    string
}

type BedrockConfig struct {
	Region          string `mapstructure:"region"`
	AccessKeyID     string `mapstructure:"access_key_id"`
	SecretAccessKey string `mapstructure:"secret_access_key"`
}

type ServerConfig struct {
	Addr            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
}

type PostgresConfig struct {
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

type RedisConfig struct {
	Enabled  bool
	Addr     string
	Password string
	DB       int
}

type AuthConfig struct {
	APIKeys map[string]string // key -> key_id
}

type GatewayConfig struct {
	GuardrailEndpoint    string
	GuardrailTimeout     time.Duration
	RateLimitRPS         int
	RateLimitBurst       int
	PricingSyncEnabled   bool
	PricingSyncInterval  time.Duration
	PricingSyncURL       string
	PricingCacheInterval time.Duration
}

// BindFlags registers CLI flags on the given flag set. Call this before parsing.
func BindFlags(fs *pflag.FlagSet) {
	fs.String("config", "", "Path to config file (yaml, toml, json)")

	fs.String("listen-addr", ":8080", "Server listen address")
	fs.Duration("server-read-timeout", 30*time.Second, "HTTP read timeout")
	fs.Duration("server-write-timeout", 120*time.Second, "HTTP write timeout")
	fs.Duration("server-idle-timeout", 60*time.Second, "HTTP idle timeout")
	fs.Duration("server-shutdown-timeout", 15*time.Second, "Graceful shutdown timeout")

	fs.String("postgres-dsn", "", "PostgreSQL connection string")
	fs.Int("postgres-max-open-conns", 25, "Max open database connections")
	fs.Int("postgres-max-idle-conns", 5, "Max idle database connections")
	fs.Duration("postgres-conn-max-lifetime", 5*time.Minute, "Max connection lifetime")

	fs.Bool("redis-enabled", false, "Enable Redis-backed rate limiting")
	fs.String("redis-addr", "localhost:6379", "Redis address")
	fs.String("redis-password", "", "Redis password")
	fs.Int("redis-db", 0, "Redis database number")

	fs.String("api-keys", "", "Comma-separated KEY:ID pairs for auth")

	fs.String("guardrail-endpoint", "", "External guardrail service URL")
	fs.Duration("guardrail-timeout", 5*time.Second, "Guardrail request timeout")
	fs.Int("rate-limit-rps", 60, "Requests per second per API key")
	fs.Int("rate-limit-burst", 10, "Token bucket burst size")

	fs.String("openai-api-key", "", "OpenAI API key")
	fs.String("openai-base-url", "https://api.openai.com", "OpenAI base URL")
	fs.String("anthropic-api-key", "", "Anthropic API key")
	fs.String("anthropic-base-url", "https://api.anthropic.com", "Anthropic base URL")
	fs.String("vertex-project-id", "", "Google Cloud project ID for Vertex AI OpenAI-compatible endpoint")
	fs.String("vertex-region", "us-central1", "Google Cloud region for Vertex AI OpenAI-compatible endpoint")
	fs.String("bedrock-region", "", "AWS region for Bedrock runtime requests")
	fs.String("bedrock-access-key-id", "", "AWS access key ID for Bedrock (optional; default credential chain is used when unset)")
	fs.String("bedrock-secret-access-key", "", "AWS secret access key for Bedrock (required when access key ID is set)")

	fs.Bool("pricing-sync-enabled", true, "Enable periodic pricing sync from upstream")
	fs.Duration("pricing-sync-interval", 6*time.Hour, "How often to sync pricing from upstream")
	fs.String("pricing-sync-url", "", "Custom upstream pricing JSON URL (defaults to LiteLLM)")
	fs.Duration("pricing-cache-interval", 1*time.Minute, "How often to refresh the in-memory pricing cache from the database")
	fs.String("provider-credentials-key", "", "Base64-encoded 32-byte key for provider credential encryption at rest")
}

// Load reads configuration with 12-factor precedence:
// CLI flags > environment variables > config file > defaults.
func Load(fs *pflag.FlagSet) (*Config, error) {
	v := viper.New()

	// Map environment variables: LLMGOPHER_LISTEN_ADDR -> listen-addr
	v.SetEnvPrefix("LLMGOPHER")
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	v.AutomaticEnv()

	// Also support un-prefixed env vars for backward compatibility and simplicity.
	// Viper's AutomaticEnv with prefix handles LLMGOPHER_*, but we also want
	// to accept POSTGRES_DSN, REDIS_ADDR, etc. directly.
	bindUnprefixedEnv(v)

	if err := v.BindPFlags(fs); err != nil {
		return nil, fmt.Errorf("bind flags: %w", err)
	}

	// Load config file if specified via --config flag or LLMGOPHER_CONFIG env var.
	if cfgFile := v.GetString("config"); cfgFile != "" {
		v.SetConfigFile(cfgFile)
		if err := v.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("read config file %s: %w", cfgFile, err)
		}
	} else {
		// Search default locations.
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath(".")
		v.AddConfigPath("/etc/llmgopher")
		// Silently ignore missing default config file.
		_ = v.ReadInConfig()
	}

	cfg := &Config{
		Server: ServerConfig{
			Addr:            v.GetString("listen-addr"),
			ReadTimeout:     v.GetDuration("server-read-timeout"),
			WriteTimeout:    v.GetDuration("server-write-timeout"),
			IdleTimeout:     v.GetDuration("server-idle-timeout"),
			ShutdownTimeout: v.GetDuration("server-shutdown-timeout"),
		},
		Postgres: PostgresConfig{
			DSN:             v.GetString("postgres-dsn"),
			MaxOpenConns:    v.GetInt("postgres-max-open-conns"),
			MaxIdleConns:    v.GetInt("postgres-max-idle-conns"),
			ConnMaxLifetime: v.GetDuration("postgres-conn-max-lifetime"),
		},
		Redis: RedisConfig{
			Enabled:  v.GetBool("redis-enabled"),
			Addr:     v.GetString("redis-addr"),
			Password: v.GetString("redis-password"),
			DB:       v.GetInt("redis-db"),
		},
		Auth: AuthConfig{
			APIKeys: parseAPIKeys(v.GetString("api-keys")),
		},
		Gateway: GatewayConfig{
			GuardrailEndpoint:    v.GetString("guardrail-endpoint"),
			GuardrailTimeout:     v.GetDuration("guardrail-timeout"),
			RateLimitRPS:         v.GetInt("rate-limit-rps"),
			RateLimitBurst:       v.GetInt("rate-limit-burst"),
			PricingSyncEnabled:   v.GetBool("pricing-sync-enabled"),
			PricingSyncInterval:  v.GetDuration("pricing-sync-interval"),
			PricingSyncURL:       v.GetString("pricing-sync-url"),
			PricingCacheInterval: v.GetDuration("pricing-cache-interval"),
		},
		Providers: ProvidersConfig{
			OpenAI: ProviderEndpoint{
				APIKey:  v.GetString("openai-api-key"),
				BaseURL: v.GetString("openai-base-url"),
			},
			Anthropic: ProviderEndpoint{
				APIKey:  v.GetString("anthropic-api-key"),
				BaseURL: v.GetString("anthropic-base-url"),
			},
			Vertex: VertexConfig{
				ProjectID: v.GetString("vertex-project-id"),
				Region:    v.GetString("vertex-region"),
			},
			Bedrock: BedrockConfig{
				Region:          v.GetString("bedrock-region"),
				AccessKeyID:     v.GetString("bedrock-access-key-id"),
				SecretAccessKey: v.GetString("bedrock-secret-access-key"),
			},
		},
		Security: SecurityConfig{
			ProviderCredentialsKey: v.GetString("provider-credentials-key"),
		},
	}

	if cfg.Postgres.DSN == "" {
		return nil, fmt.Errorf("postgres-dsn is required (set via flag, env var LLMGOPHER_POSTGRES_DSN, or config file)")
	}

	return cfg, nil
}

// bindUnprefixedEnv allows common env vars without the LLMGOPHER_ prefix.
func bindUnprefixedEnv(v *viper.Viper) {
	envMap := map[string]string{
		"listen-addr":                "LISTEN_ADDR",
		"postgres-dsn":               "POSTGRES_DSN",
		"postgres-max-open-conns":    "POSTGRES_MAX_OPEN_CONNS",
		"postgres-max-idle-conns":    "POSTGRES_MAX_IDLE_CONNS",
		"postgres-conn-max-lifetime": "POSTGRES_CONN_MAX_LIFETIME",
		"redis-enabled":              "REDIS_ENABLED",
		"redis-addr":                 "REDIS_ADDR",
		"redis-password":             "REDIS_PASSWORD",
		"redis-db":                   "REDIS_DB",
		"api-keys":                   "API_KEYS",
		"guardrail-endpoint":         "GUARDRAIL_ENDPOINT",
		"guardrail-timeout":          "GUARDRAIL_TIMEOUT",
		"rate-limit-rps":             "RATE_LIMIT_RPS",
		"rate-limit-burst":           "RATE_LIMIT_BURST",
		"openai-api-key":             "OPENAI_API_KEY",
		"openai-base-url":            "OPENAI_BASE_URL",
		"anthropic-api-key":          "ANTHROPIC_API_KEY",
		"anthropic-base-url":         "ANTHROPIC_BASE_URL",
		"vertex-project-id":          "VERTEX_PROJECT_ID",
		"vertex-region":              "VERTEX_REGION",
		"bedrock-region":             "BEDROCK_REGION",
		"bedrock-access-key-id":      "BEDROCK_ACCESS_KEY_ID",
		"bedrock-secret-access-key":  "BEDROCK_SECRET_ACCESS_KEY",
		"server-read-timeout":        "SERVER_READ_TIMEOUT",
		"server-write-timeout":       "SERVER_WRITE_TIMEOUT",
		"server-idle-timeout":        "SERVER_IDLE_TIMEOUT",
		"server-shutdown-timeout":    "SERVER_SHUTDOWN_TIMEOUT",
		"pricing-sync-enabled":       "PRICING_SYNC_ENABLED",
		"pricing-sync-interval":      "PRICING_SYNC_INTERVAL",
		"pricing-sync-url":           "PRICING_SYNC_URL",
		"pricing-cache-interval":     "PRICING_CACHE_INTERVAL",
		"provider-credentials-key":   "PROVIDER_CREDENTIALS_KEY",
	}
	for key, env := range envMap {
		_ = v.BindEnv(key, env)
	}
}

func parseAPIKeys(raw string) map[string]string {
	keys := make(map[string]string)
	if raw == "" {
		return keys
	}
	for _, pair := range strings.Split(raw, ",") {
		parts := strings.SplitN(strings.TrimSpace(pair), ":", 2)
		if len(parts) == 2 {
			keys[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}
	return keys
}
