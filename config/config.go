package config

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
)

type Config struct {
	DBUser                string `json:"db_user"`
	DBPass                string `json:"db_pass"`
	DBPort                int    `json:"db_port"`
	DBName                string `json:"db_name"`
	ServerPort            string `json:"server_port"`
	MinimumDeposit        int    `json:"minimum_deposit"`
	MinimumBet            int    `json:"minimum_bet"`
	MinimumPasswordLength int    `json:"minimum_password_length"`
	MinimumNameLength     int    `json:"minimum_name_length"`
	MaximumNameLength     int    `json:"maximum_name_length"`
	SecretKey             string `json:"secret_key"`
	MaxTokenLifeMinutes   int    `json:"max_token_life_minutes"`
}

const configPath = "/config/config.json"

var Settings Config

// LoadConfig loads the config from a predeefined path, relative to the cwd
// if it fails, panic occurs and the application does not start
func LoadConfig() {

	executableDir, _ := os.Getwd()
	absoluteConfigDir := filepath.Join(executableDir, configPath)

	configFile, err := os.Open(absoluteConfigDir)
	if err != nil {
		panic("Config file not found")
	}
	defer configFile.Close()

	comfigFileData, err := io.ReadAll(configFile)
	if err != nil {
		panic("Error reading config file")
	}

	err = json.Unmarshal(comfigFileData, &Settings)
	if err != nil {
		panic("Config file could not be parsed")
	}

}
