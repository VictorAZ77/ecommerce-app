package repositories

import (
	"backend/internal/models"
	"database/sql"
	"fmt"
	"log"
)

// define las operaciones de persistencia para Client.
type ClientRepository interface {
	GetByID(id string) (*models.Client, error)
	GetClientByEmail(email string) (*models.Client, error)
	GetAll() ([]models.Client, error)
	Create(client models.Client) error
	Update(client models.Client) error
	Delete(id string) error
}

// implementación de ClientRepository para SQL Server.
type SQLServerClientRepository struct {
	db *sql.DB
}

// crea una nueva instancia de SQLServerClientRepository.
func NewSQLServerClientRepository(db *sql.DB) *SQLServerClientRepository {
	return &SQLServerClientRepository{db: db}
}

// Implementación de los métodos de la interfaz ClientRepository
func (r *SQLServerClientRepository) GetByID(id string) (*models.Client, error) {
	client := &models.Client{}
	query := "SELECT CONVERT(CHAR(255),Id) AS 'char', Name, Email, Address, CreatedAt, UpdatedAt FROM Clients WHERE ID = @p1"
	err := r.db.QueryRow(query, id).Scan(&client.ID, &client.Name, &client.Email, &client.Address, &client.CreatedAt, &client.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("cliente con ID %s no encontrado", id)
	}
	if err != nil {
		return nil, fmt.Errorf("error al obtener cliente por ID: %w", err)
	}
	return client, nil
}
func (r *SQLServerClientRepository) GetClientByEmail(email string) (*models.Client, error) {
	client := &models.Client{}
	query := "SELECT CONVERT(CHAR(255),Id) AS 'char', Name, Email, Address, CreatedAt, UpdatedAt FROM Clients WHERE Email = @p1"
	row := r.db.QueryRow(query, email)
	err := row.Scan(
		&client.ID,
		&client.Name,
		&client.Email,
		&client.Address,
		&client.CreatedAt,
		&client.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("cliente con Email '%s' no encontrado", email)
	}
	if err != nil {
		log.Printf("Error al escanear cliente con Email '%s': %v", email, err)
		return nil, fmt.Errorf("error al obtener cliente por Email: %w", err)
	}
	return client, nil
}

// Obtiene todos los clientes
func (r *SQLServerClientRepository) GetAll() ([]models.Client, error) {
	rows, err := r.db.Query("SELECT CONVERT(CHAR(255),Id) AS 'char', Name, Email, Address, CreatedAt, UpdatedAt FROM Clients")

	if err != nil {
		return nil, fmt.Errorf("error al obtener todos los clientes: %w", err)
	}
	defer rows.Close()

	var clients []models.Client
	for rows.Next() {
		var c models.Client
		if err := rows.Scan(&c.ID, &c.Name, &c.Email, &c.Address, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, fmt.Errorf("error al escanear fila de cliente: %w", err)
		}
		clients = append(clients, c)
	}
	return clients, nil
}

// Creacion de los clintes
func (r *SQLServerClientRepository) Create(client models.Client) error {
	_, err := r.db.Exec("INSERT INTO Clients (ID, Name, Email, Address, CreatedAt, UpdatedAt) VALUES (@p1, @p2, @p3, @p4, @p5, @p6)",
		client.ID, client.Name, client.Email, client.Address, client.CreatedAt, client.UpdatedAt)
	if err != nil {
		return fmt.Errorf("error al crear cliente: %w", err)
	}
	return nil
}

// Actualizacion de los clintes
func (r *SQLServerClientRepository) Update(client models.Client) error {
	_, err := r.db.Exec("UPDATE Clients SET Name=@p1, Email=@p2, Address=@p3, UpdatedAt=@p4 WHERE ID=@p5",
		client.Name, client.Email, client.Address, client.UpdatedAt, client.ID)
	if err != nil {
		return fmt.Errorf("error al actualizar cliente: %w", err)
	}
	return nil
}

// Eliminacion de los clintes
func (r *SQLServerClientRepository) Delete(id string) error {
	res, err := r.db.Exec("DELETE FROM Clients WHERE ID=@p1", id)
	if err != nil {
		return fmt.Errorf("error al eliminar cliente: %w", err)
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("cliente con ID %s no encontrado para eliminar", id)
	}
	return nil
}
