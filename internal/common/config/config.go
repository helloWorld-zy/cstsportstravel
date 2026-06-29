package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Config holds all application configuration.
type Config struct {
	Server     ServerConfig     `mapstructure:"server"`
	Database   DatabaseConfig   `mapstructure:"database"`
	Redis      RedisConfig      `mapstructure:"redis"`
	JWT        JWTConfig        `mapstructure:"jwt"`
	Payment    PaymentConfig    `mapstructure:"payment"`
	SMS        SMSConfig        `mapstructure:"sms"`
	Log        LogConfig        `mapstructure:"log"`
	Consul     ConsulConfig     `mapstructure:"consul"`
	Encryption EncryptionConfig `mapstructure:"encryption"`
	MFA        MFAConfig        `mapstructure:"mfa"`
	Signing    SigningConfig    `mapstructure:"signing"`
	Upload     UploadConfig     `mapstructure:"upload"`
}

// ServerConfig holds HTTP server settings.
type ServerConfig struct {
	Port         int    `mapstructure:"port"`
	Mode         string `mapstructure:"mode"` // debug, release, test
	ReadTimeout  int    `mapstructure:"read_timeout"`
	WriteTimeout int    `mapstructure:"write_timeout"`
	TLS          TLSConfig `mapstructure:"tls"`
}

// TLSConfig holds TLS/HTTPS settings.
type TLSConfig struct {
	Enabled  bool   `mapstructure:"enabled"`
	CertFile string `mapstructure:"cert_file"`
	KeyFile  string `mapstructure:"key_file"`
	MinVersion string `mapstructure:"min_version"` // "1.2" or "1.3"
}

// DatabaseConfig holds PostgreSQL connection settings.
type DatabaseConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	User         string `mapstructure:"user"`
	Password     string `mapstructure:"password"`
	DBName       string `mapstructure:"dbname"`
	SSLMode      string `mapstructure:"sslmode"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
	MaxLifetime  int    `mapstructure:"max_lifetime"` // minutes
}

// DSN returns the PostgreSQL connection string.
func (d DatabaseConfig) DSN() string {
	if d.Password == "" {
		return fmt.Sprintf(
			"host=%s port=%d user=%s dbname=%s sslmode=%s",
			d.Host, d.Port, d.User, d.DBName, d.SSLMode,
		)
	}
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.DBName, d.SSLMode,
	)
}

// RedisConfig holds Redis connection settings.
type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
	PoolSize int    `mapstructure:"pool_size"`
}

// JWTConfig holds JWT token settings.
type JWTConfig struct {
	PrivateKeyPath string `mapstructure:"private_key_path"`
	PublicKeyPath  string `mapstructure:"public_key_path"`
	AccessExpiry   int    `mapstructure:"access_expiry"`  // minutes
	RefreshExpiry  int    `mapstructure:"refresh_expiry"` // minutes
	Issuer         string `mapstructure:"issuer"`
}

// PaymentConfig holds payment channel settings.
type PaymentConfig struct {
	Alipay  AlipayConfig  `mapstructure:"alipay"`
	Wechat  WechatConfig  `mapstructure:"wechat"`
	Timeout int           `mapstructure:"timeout"` // minutes
}

// AlipayConfig holds Alipay SDK settings.
type AlipayConfig struct {
	AppID      string `mapstructure:"app_id"`
	PrivateKey string `mapstructure:"private_key"`
	PublicKey  string `mapstructure:"public_key"`
	NotifyURL  string `mapstructure:"notify_url"`
	ReturnURL  string `mapstructure:"return_url"`
}

// WechatConfig holds WeChat Pay SDK settings.
type WechatConfig struct {
	AppID     string `mapstructure:"app_id"`
	MchID     string `mapstructure:"mch_id"`
	APIKey    string `mapstructure:"api_key"`
	NotifyURL string `mapstructure:"notify_url"`
	CertPath  string `mapstructure:"cert_path"`
}

// SMSConfig holds SMS service settings.
type SMSConfig struct {
	AccessKeyID  string `mapstructure:"access_key_id"`
	AccessSecret string `mapstructure:"access_secret"`
	SignName     string `mapstructure:"sign_name"`
	TemplateCode string `mapstructure:"template_code"`
}

// LogConfig holds logging settings.
type LogConfig struct {
	Level      string `mapstructure:"level"`
	FilePath   string `mapstructure:"file_path"`
	MaxSize    int    `mapstructure:"max_size"`    // MB
	MaxBackups int    `mapstructure:"max_backups"`
	MaxAge     int    `mapstructure:"max_age"` // days
}

// ConsulConfig holds Consul connection settings.
type ConsulConfig struct {
	Addr      string `mapstructure:"addr"`
	KeyPrefix string `mapstructure:"key_prefix"`
}

// EncryptionConfig holds AES-256-GCM field encryption settings.
type EncryptionConfig struct {
	Key string `mapstructure:"key"` // 32-byte hex-encoded key
}

// MFAConfig holds TOTP MFA settings.
type MFAConfig struct {
	Issuer string `mapstructure:"issuer"` // TOTP issuer name
}

// SigningConfig holds HMAC-SHA256 request signing settings.
type SigningConfig struct {
	Secret     string `mapstructure:"secret"`      // HMAC signing secret
	Tolerance  int    `mapstructure:"tolerance"`   // timestamp tolerance in minutes
	NonceTTL   int    `mapstructure:"nonce_ttl"`   // nonce dedup TTL in minutes
}

// UploadConfig holds OSS file upload settings.
type UploadConfig struct {
	AccessKeyID     string `mapstructure:"access_key_id"`
	AccessKeySecret string `mapstructure:"access_key_secret"`
	BucketName      string `mapstructure:"bucket_name"`
	Region          string `mapstructure:"region"`
	Endpoint        string `mapstructure:"endpoint"`
	CDNDomain       string `mapstructure:"cdn_domain"`
	BasePath        string `mapstructure:"base_path"`
}

// Load reads configuration from file and environment variables.
func Load(configPath string) (*Config, error) {
	v := viper.New()

	// Set defaults
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.mode", "debug")
	v.SetDefault("server.read_timeout", 30)
	v.SetDefault("server.write_timeout", 30)
	v.SetDefault("database.host", "localhost")
	v.SetDefault("database.port", 5432)
	v.SetDefault("database.user", "postgres")
	v.SetDefault("database.password", "")
	v.SetDefault("database.dbname", "travel_booking")
	v.SetDefault("database.sslmode", "disable")
	v.SetDefault("database.max_open_conns", 25)
	v.SetDefault("database.max_idle_conns", 5)
	v.SetDefault("database.max_lifetime", 30)
	v.SetDefault("redis.addr", "localhost:6379")
	v.SetDefault("redis.password", "")
	v.SetDefault("redis.db", 0)
	v.SetDefault("redis.pool_size", 10)
	v.SetDefault("jwt.access_expiry", 15)
	v.SetDefault("jwt.refresh_expiry", 10080) // 7 days
	v.SetDefault("jwt.issuer", "travel-booking")
	v.SetDefault("payment.timeout", 30)
	v.SetDefault("log.level", "info")
	v.SetDefault("log.file_path", "logs/app.log")
	v.SetDefault("log.max_size", 100)
	v.SetDefault("log.max_backups", 7)
	v.SetDefault("log.max_age", 7)
	v.SetDefault("consul.addr", "localhost:8500")
	v.SetDefault("consul.key_prefix", "travel-booking/")
	v.SetDefault("encryption.key", "")
	v.SetDefault("mfa.issuer", "TravelBooking")
	v.SetDefault("signing.secret", "")
	v.SetDefault("signing.tolerance", 5)
	v.SetDefault("signing.nonce_ttl", 5)

	// Read config file
	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath("./configs")
		v.AddConfigPath(".")
	}

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("read config: %w", err)
		}
	}

	// Environment variable binding
	v.SetEnvPrefix("TRAVEL")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Bind specific env vars for sensitive values
	_ = v.BindEnv("database.password", "TRAVEL_DB_PASSWORD")
	_ = v.BindEnv("redis.password", "TRAVEL_REDIS_PASSWORD")
	_ = v.BindEnv("jwt.private_key_path", "TRAVEL_JWT_PRIVATE_KEY")
	_ = v.BindEnv("jwt.public_key_path", "TRAVEL_JWT_PUBLIC_KEY")
	_ = v.BindEnv("payment.alipay.private_key", "TRAVEL_ALIPAY_PRIVATE_KEY")
	_ = v.BindEnv("payment.wechat.api_key", "TRAVEL_WECHAT_API_KEY")
	_ = v.BindEnv("sms.access_key_id", "TRAVEL_SMS_ACCESS_KEY")
	_ = v.BindEnv("sms.access_secret", "TRAVEL_SMS_ACCESS_SECRET")
	_ = v.BindEnv("encryption.key", "TRAVEL_ENCRYPTION_KEY")
	_ = v.BindEnv("signing.secret", "TRAVEL_SIGNING_SECRET")

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	return &cfg, nil
}
