package database

import (
	"database/sql"
	"fmt"
	_ "github.com/denisenkom/go-mssqldb"
	"log"
)

// inicializa la conexión a SQL Server.
func InitDB(connStr string) (*sql.DB, error) {
	db, err := sql.Open("sqlserver", connStr)
	if err != nil {
		return nil, fmt.Errorf("error al abrir DB: %w", err)
	}

	//Configurar pool de conexiones
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("error al conectar con DB: %w", err)
	}

	log.Println("Conexión a SQL Server exitosa.")

	return db, nil
}
