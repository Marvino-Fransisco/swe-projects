package lib

import (
	"bufio"
	"log"
	"os"
	"strconv"
	"strings"
)

type EnvStruct struct {
	Port        string
	Environment string
}

var Env EnvStruct

func init() {
	file, err := os.Open(".env")
	if err != nil {
		Env = EnvStruct{
			Port:        getEnv("PORT", "6969"),
			Environment: getEnv("ENVIRONMENT", "development"),
		}
		return
	}
	defer file.Close()

	envMap := make(map[string]string)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		envMap[key] = value
	}

	if err := scanner.Err(); err != nil {
		log.Fatal("Failed to parse .env:", err)
	}

	Env = EnvStruct{
		Port:        envMap["PORT"],
		Environment: envMap["ENVIRONMENT"],
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}
