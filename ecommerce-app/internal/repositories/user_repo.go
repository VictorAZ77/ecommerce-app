package repositories

import (
	"backend/internal/models"
	"database/sql"
	"fmt"
	"time"
)

// SQLServerUserRepository implementa UserRepository para SQL Server.
type SQLServerUserRepository struct {
	db *sql.DB
}

// NewSQLServerUserRepository crea una nueva instancia de SQLServerUserRepository.
func NewSQLServerUserRepository(db *sql.DB) *SQLServerUserRepository {
	return &SQLServerUserRepository{db: db}
}

// GetDB devuelve la instancia de la base de datos.
func (r *SQLServerUserRepository) GetDB() *sql.DB {
	return r.db
}

// Create inserta un nuevo usuario en la base de datos.
func (r *SQLServerUserRepository) Create(user models.User) (*models.User, error) {
	query := `
		INSERT INTO Users (ID, Username, PasswordHash, Role, CreatedAt, UpdatedAt)
		VALUES (@p1, @p2, @p3, @p4, @p5, @p6);`
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	_, err := r.db.Exec(query,
		sql.Named("p1", user.ID),
		sql.Named("p2", user.Username),
		sql.Named("p3", user.Password),
		sql.Named("p4", user.Role),
		sql.Named("p5", user.CreatedAt),
		sql.Named("p6", user.UpdatedAt),
	)
	if err != nil {
		return nil, fmt.Errorf("falló al crear usuario en DB: %w", err)
	}
	return &user, nil
}

// GetByID busca un usuario por su ID.
func (r *SQLServerUserRepository) GetByID(id string) (*models.User, error) {
	var user models.User
	query := `
		SELECT ID, Username, PasswordHash, Role, CreatedAt, UpdatedAt
		FROM Users
		WHERE ID = @p1;`

	err := r.db.QueryRow(query, sql.Named("p1", id)).Scan(
		&user.ID,
		&user.Username,
		&user.Password,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("usuario con ID '%s' no encontrado", id)
		}
		return nil, fmt.Errorf("falló al obtener usuario por ID: %w", err)
	}
	return &user, nil
}

// GetByUsername busca un usuario por su nombre de usuario.
func (r *SQLServerUserRepository) GetByUsername(username string) (*models.User, error) {
	var user models.User
	query := `
		SELECT ID, Username, PasswordHash, Role, CreatedAt, UpdatedAt
		FROM Users
		WHERE Username = @p1;`

	err := r.db.QueryRow(query, sql.Named("p1", username)).Scan(
		&user.ID,
		&user.Username,
		&user.Password,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("usuario con nombre de usuario '%s' no encontrado", username)
		}
		return nil, fmt.Errorf("falló al obtener usuario por nombre de usuario: %w", err)
	}
	return &user, nil
}

// Update actualiza un usuario existente en la base de datos.
func (r *SQLServerUserRepository) Update(user models.User) error {
	query := `
		UPDATE Users
		SET Username = @p1, PasswordHash = @p2, Role = @p3, UpdatedAt = @p4
		WHERE ID = @p5;`

	user.UpdatedAt = time.Now()

	_, err := r.db.Exec(query,
		sql.Named("p1", user.Username),
		sql.Named("p2", user.Password),
		sql.Named("p3", user.Role),
		sql.Named("p4", user.UpdatedAt),
		sql.Named("p5", user.ID),
	)
	if err != nil {
		return fmt.Errorf("falló al actualizar usuario en DB: %w", err)
	}
	return nil
}

// Delete elimina un usuario por su ID.
func (r *SQLServerUserRepository) Delete(id string) error {
	query := `DELETE FROM Users WHERE ID = @p1;`
	res, err := r.db.Exec(query, sql.Named("p1", id))
	if err != nil {
		return fmt.Errorf("falló al eliminar usuario de DB: %w", err)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("falló al obtener filas afectadas después de eliminar usuario: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("usuario con ID '%s' no encontrado para eliminar", id)
	}
	return nil
}
