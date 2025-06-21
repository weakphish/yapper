package config

import (
	"fmt"

	"github.com/spf13/viper"
)

var YapConfig *viper.Viper

const DbPath = "DbPath"

// TODO: define configuration defaults
func InitConfig() {
	YapConfig = viper.New()

	YapConfig.SetDefault(DbPath, "$HOME/.yapper/yapper.db")

	// Set up config file
	YapConfig.SetConfigName("config") // name of config file (without extension)
	YapConfig.SetConfigType("toml")   // REQUIRED if the config file does not have the extension in the name
	viper.AddConfigPath("$HOME/.yapper")
	viper.AddConfigPath(".")    // optionally look for config in the working directory
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		panic(fmt.Errorf("fatal error config file: %w", err))
	}
}
