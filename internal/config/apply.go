package config

import (
	"__MODULE__/internal/utils"
	"reflect"
)

var validate = utils.GetValidator()

// Apply applies the configuration to the provided config instance. If the ByPass flag is set, it skips
// the Vault loading and uses the environment variables directly. The function also sets any default
// values for the configuration.
func Apply(conf *App) error {
	conf.DatabaseConfig = DatabaseConfig{Username: "evolenta", Password: "salam", Host: "localhost", Port: "5432", Database: "evolenta"}

	return nil
}

// setDefaults sets the default values for the fields in the provided App configuration.
// If the configuration is nil, this function does nothing.
// It recursively sets the default values for all fields in the configuration struct.
func setDefaults(conf *App) {
	if conf == nil {
		return
	}

	setDefaultsRecursive(reflect.TypeOf(conf).Elem(), reflect.ValueOf(conf).Elem())
}
