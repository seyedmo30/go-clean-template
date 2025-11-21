package config

import (
	"fmt"

	"github.com/caarlos0/env/v11"
)

// Apply applies the configuration to the provided config instance. If the ByPass flag is set, it skips
// the Vault loading and uses the environment variables directly. The function also sets any default
// values for the configuration.
func Apply(conf *App) error {
	if conf == nil {
		return fmt.Errorf("config: nil App passed to Apply")
	}

	// Parse environment variables into conf.
	// env will respect `envDefault` tags and required settings declared as `env:"NAME,required"`.
	if err := env.Parse(conf); err != nil {
		return fmt.Errorf("failed to parse env into config: %w", err)
	}

	return nil
}
