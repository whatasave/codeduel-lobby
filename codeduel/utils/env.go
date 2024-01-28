package utils

import (
	"log"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

func LoadingEnvVars() {
  isProduction := os.Getenv("GO_ENV") == "production"
  if isProduction { return }
  
  path_dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
  if err != nil { log.Printf("[MAIN] Error getting absolute path: %v", err) }
  // log.Default().SetPrefix("[MAIN] ")
  log.Printf("[MAIN] Info: Loading .env file from %s", path_dir)
  env_path := filepath.Join(path_dir, ".env")
  if _, err := os.Stat(env_path); os.IsNotExist(err) {
    log.Printf("[MAIN] Error: .env file not found in %s", path_dir)
    return
  }
  err = godotenv.Load(env_path)
  if err != nil { log.Printf("[MAIN] Error loading .env file") }
}

func WarnUndefinedEnvVars() {
  envVars := []string{
    "HOST",
    "PORT",
  }
  
  log.Printf("")
  for _, envVar := range envVars {
    test, exists := os.LookupEnv(envVar)
    if !exists {
      log.Printf("[MAIN] Warning: %s not defined in .env file", envVar)
      continue
    }
    if test == "" { log.Printf("[MAIN] Warning: %s is empty", envVar) }
    log.Printf("[MAIN] %s: %s", envVar, test)
  }
  log.Printf("")
}