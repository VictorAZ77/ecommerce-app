package config

import (
	"github.com/joho/godotenv"
	"log"
	"os"
)

type AppConfig struct {
	DatabaseURL string
	ServerPort  string
	SessionKey  []byte
}

// carga la configuración desde variables de entorno
func LoadConfig() *AppConfig {
	err := godotenv.Load()
	if err != nil && !os.IsNotExist(err) {
		log.Printf("Advertencia: No se pudo cargar el archivo .env: %v. Usando variables de entorno existentes.", err)
	}

	// carga la variables BD
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL no está configurada. Verifica variables de entorno o .env.")
	}
	// carga el puerto defindo para la esucha local
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}
	// carga llave de la sesion definida
	sessionKey := os.Getenv("SESSION_KEY")
	if sessionKey == "" {
		log.Println("SESSION_KEY no establecida, usando valor por defecto.")
		sessionKey = "super-secret-key-for-development-only-1234567890"
	}
	return &AppConfig{DatabaseURL: dbURL, ServerPort: port, SessionKey: []byte(sessionKey)}
}
