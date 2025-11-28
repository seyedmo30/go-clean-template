/*
Package config contains the configuration for the application.
*/
package config

// use github.com/caarlos0/env/v11 to parse these tags
type App struct {
	KafkaConfig
	DatabaseConfig
	UsecaseConfig
	ProviderConfig
	WorkerConfig
	AppConfig
}

type AppConfig struct {
	// required environment variable
	AppEnv string `env:"APP_ENV"  envDefault:"development"`
	// default value 30s if not set
	SettingsTTL string `envDefault:"30s" env:"SETTINGS_TTL"`
}

type KafkaConfig struct {
	// add kafka fields here when needed, e.g.
	// Brokers []string `env:"KAFKA_BROKERS" envSeparator:","`
}

type DatabaseConfig struct {
	Username string `env:"DB_USERNAME" envDefault:"postgres"` // add ,required if needed: env:"DB_USERNAME,required"
	Password string `env:"DB_PASSWORD" envDefault:"salam"`    // probably required, add option if so
	Host     string `env:"DB_HOST" envDefault:"localhost"`
	Port     string `env:"DB_PORT" envDefault:"5432"`
	Database string `env:"DB_DATABASE" envDefault:"clean_template"`
	PoolSize int    `env:"DB_POOL_SIZE" envDefault:"10"`
	MaxIdle  int    `env:"DB_MAX_IDLE" envDefault:"5"`
}

type UsecaseConfig struct {
	// add fields as needed
}

type ProviderConfig struct {
	ReqresConfig
	JsonplaceholderConfig
}

type ReqresConfig struct {
	BaseUrl              string `env:"REQRES_BASE_URL"  envDefault:"30"`
	AuthenticationUrl    string `env:"REQRES_AUTHENTICATION_URL"  envDefault:""`
	Timeout              int    `env:"REQRES_TIMEOUT" envDefault:"30"`
	RetryCount           int    `env:"REQRES_RETRY_COUNT" envDefault:"0"`
	RefreshTokenInterval int    `env:"REQRES_REFRESH_TOKEN_INTERVAL" envDefault:"600"`
	GrantType            string `env:"REQRES_GRANT_TYPE" envDefault:"client_credentials"`
}

type JsonplaceholderConfig struct {
	Timeout              int    `env:"JSONPLACEHOLDER_TIMEOUT" envDefault:"30"`
	RetryCount           int    `env:"JSONPLACEHOLDER_RETRY_COUNT" envDefault:"0"`
	RefreshTokenInterval int    `env:"JSONPLACEHOLDER_REFRESH_TOKEN_INTERVAL" envDefault:"600"`
	GrantType            string `env:"JSONPLACEHOLDER_GRANT_TYPE" envDefault:"client_credentials"`
}

type WorkerConfig struct {
	ExpirePendingEndOfDay string `env:"EXPIRE_PENDING_END_OF_DAY" envDefault:"0 0 * * *"`
}
