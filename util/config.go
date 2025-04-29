package util

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

type Config struct {
	AllowedOrigins            []string `mapstructure:"ALLOWED_ORIGINS"`
	AmadeusAPIKey             string   `mapstructure:"AMADEUS_API_KEY"`
	AmadeusAPISecret          string   `mapstructure:"AMADEUS_API_SECRET"`
	AmadeusAPIBaseURL         string   `mapstructure:"AMADEUS_API_BASE_URL"`
	DefaultUserName           string   `mapstructure:"DEFAULT_USER_NAME"`
	DefaultUserPassword       string   `mapstructure:"DEFAULT_USER_PASS"`
	Environment               string   `mapstructure:"ENVIRONMENT"`
	FlightAPIKey              string   `mapstructure:"FLIGHT_API_KEY"`
	FlightAPIURL              string   `mapstructure:"FLIGHT_API_URL"`
	Host                      string   `mapstructure:"HTTP_HOST"`
	JwtKey                    string   `mapstructure:"JWT_KEY"`
	Port                      string   `mapstructure:"HTTP_PORT"`
	SkyScannerRapidAPIBaseURL string   `mapstructure:"SKY_SCANNER_RAPID_API_BASE_URL"`
	SkyScannerRapidAPIHost    string   `mapstructure:"SKY_SCANNER_RAPID_API_HOST"`
	SkyScannerRapidAPIKey     string   `mapstructure:"SKY_SCANNER_RAPID_API_KEY"`
}

func (c Config) Validate() error {
	if len(c.AmadeusAPIKey) == 0 {
		return ErrMissingEnvironmentVariable("ALLOWED_ORIGINS")
	}

	if len(c.AmadeusAPIKey) == 0 {
		return ErrMissingEnvironmentVariable("AMADEUS_API_KEY")
	}

	if len(c.AmadeusAPISecret) == 0 {
		return ErrMissingEnvironmentVariable("AMADEUS_API_SECRET")
	}

	if len(c.AmadeusAPIBaseURL) == 0 {
		return ErrMissingEnvironmentVariable("AMADEUS_API_BASE_URL")
	}

	if len(c.DefaultUserName) == 0 {
		return ErrMissingEnvironmentVariable("DEFAULT_USER_NAME")
	}

	if len(c.DefaultUserPassword) == 0 {
		return ErrMissingEnvironmentVariable("DEFAULT_USER_PASS")
	}

	if len(c.FlightAPIKey) == 0 {
		return ErrMissingEnvironmentVariable("FLIGHT_API_KEY")
	}

	if len(c.FlightAPIURL) == 0 {
		return ErrMissingEnvironmentVariable("FLIGHT_API_URL")
	}

	if len(c.Host) == 0 {
		return ErrMissingEnvironmentVariable("HTTP_HOST")
	}

	if len(c.JwtKey) == 0 {
		return ErrMissingEnvironmentVariable("JWT_KEY")
	}

	if len(c.Port) == 0 {
		return ErrMissingEnvironmentVariable("HTTP_PORT")
	}

	if len(c.SkyScannerRapidAPIBaseURL) == 0 {
		return ErrMissingEnvironmentVariable("SKY_SCANNER_RAPID_API_BASE_URL")
	}

	if len(c.SkyScannerRapidAPIHost) == 0 {
		return ErrMissingEnvironmentVariable("SKY_SCANNER_RAPID_API_HOST")
	}

	if len(c.SkyScannerRapidAPIKey) == 0 {
		return ErrMissingEnvironmentVariable("SKY_SCANNER_RAPID_API_KEY")
	}

	return nil
}

func ErrMissingEnvironmentVariable(v string) error {
	return fmt.Errorf("missing %s environment variable", v)
}

func LoadConfig(path string) (Config, error) {
	viper.AddConfigPath(path)

	// for debugging
	pwd, _ := os.Getwd()
	viper.AddConfigPath(pwd + "../../")

	viper.SetConfigName("app")
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {

		// we can ignore this error if values were loaded from secrets or docker env
		envEmpty := !viper.IsSet("HTTP_PORT")
		if _, ok := err.(viper.ConfigFileNotFoundError); ok && envEmpty {
			return Config{}, err
		}
	}

	var config Config

	if err := viper.Unmarshal(&config); err != nil {
		return Config{}, err
	}

	if err := config.Validate(); err != nil {
		return config, err
	}

	return config, nil
}
