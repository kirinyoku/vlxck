package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kirinyoku/vlxck/internal/utils"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
)

// Config represents the synchronization configuration
type Config struct {
	Sync struct {
		Provider              string `mapstructure:"provider"`
		FileID                string `mapstructure:"file_id"`
		Etag                  string `mapstructure:"etag"`
		EncryptedToken        []byte `mapstructure:"encrypted_token"`
		EncryptedClientId     []byte `mapstructure:"encrypted_client_id"`
		EncryptedClientSecret []byte `mapstructure:"encrypted_client_secret"`
	} `mapstructure:"sync"`
}

// LoadConfig loads the configuration from file
func LoadConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(filepath.Join(os.Getenv("HOME"), ".vlxck"))

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return &Config{}, nil
		}
		return nil, fmt.Errorf("failed to read config: %v", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %v", err)
	}
	return &config, nil
}

// SaveConfig saves the configuration to file
func SaveConfig(config *Config) error {
	viper.Set("sync.provider", config.Sync.Provider)
	viper.Set("sync.file_id", config.Sync.FileID)
	viper.Set("sync.etag", config.Sync.Etag)
	viper.Set("sync.encrypted_token", config.Sync.EncryptedToken)
	viper.Set("sync.encrypted_client_id", config.Sync.EncryptedClientId)
	viper.Set("sync.encrypted_client_secret", config.Sync.EncryptedClientSecret)

	configPath := filepath.Join(os.Getenv("HOME"), ".vlxck", "config.yaml")
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %v", err)
	}

	if err := viper.WriteConfigAs(configPath); err != nil {
		return fmt.Errorf("failed to write config to %s: %v", configPath, err)
	}

	if info, err := os.Stat(configPath); err != nil {
		return fmt.Errorf("failed to stat config file: %v", err)
	} else if info.Size() == 0 {
		return fmt.Errorf("config file %s is empty after writing", configPath)
	}

	return nil
}

// EncryptToken encrypts an OAuth2 token
func EncryptToken(token *oauth2.Token, password string) ([]byte, error) {
	if password == "" {
		return nil, fmt.Errorf("password cannot be empty")
	}

	data, err := json.Marshal(token)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal token: %v", err)
	}

	return utils.EncryptFile(data, password)
}

// DecryptToken decrypts an OAuth2 token
func DecryptToken(encryptedToken []byte, password string) (*oauth2.Token, error) {
	if len(encryptedToken) == 0 {
		return nil, fmt.Errorf("encrypted token is empty")
	}

	data, err := utils.DecryptFile(encryptedToken, password)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt token: %v", err)
	}

	var token oauth2.Token
	if err := json.Unmarshal(data, &token); err != nil {
		return nil, fmt.Errorf("failed to unmarshal token: %v", err)
	}
	return &token, nil
}

// EncryptClientCredentials encrypts ClientID and ClientSecret
func EncryptClientCredentials(clientID, clientSecret, password string) ([]byte, []byte, error) {
	if password == "" {
		return nil, nil, fmt.Errorf("password cannot be empty")
	}

	encryptedClientID, err := utils.EncryptFile([]byte(clientID), password)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to encrypt client ID: %v", err)
	}

	encryptedClientSecret, err := utils.EncryptFile([]byte(clientSecret), password)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to encrypt client secret: %v", err)
	}

	return encryptedClientID, encryptedClientSecret, nil
}

// DecryptClientCredentials decrypts ClientID and ClientSecret
func DecryptClientCredentials(encryptedClientID, encryptedClientSecret []byte, password string) (string, string, error) {
	if len(encryptedClientID) == 0 || len(encryptedClientSecret) == 0 {
		return "", "", fmt.Errorf("encrypted client credentials are empty")
	}

	clientID, err := utils.DecryptFile(encryptedClientID, password)
	if err != nil {
		return "", "", fmt.Errorf("failed to decrypt client ID: %v", err)
	}

	clientSecret, err := utils.DecryptFile(encryptedClientSecret, password)
	if err != nil {
		return "", "", fmt.Errorf("failed to decrypt client secret: %v", err)
	}

	return string(clientID), string(clientSecret), nil
}
