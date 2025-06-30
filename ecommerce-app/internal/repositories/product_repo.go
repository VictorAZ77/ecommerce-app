package repositories

import (
	"backend/internal/models"
	"database/sql"
	"fmt"
	"time"
)

// define las operaciones de persistencia para el modelo Product.
type ProductRepository interface {
	GetByID(id string) (*models.Product, error)
	GetAll() ([]models.Product, error)
	Create(product models.Product) error
	Update(product models.Product) error
	UpdateProductStock(productID string, quantity int) error
	Delete(id string) error
}

// implementación de ProductRepository para SQL Server.
type SQLServerProductRepository struct {
	db *sql.DB
}

// crea una nueva instancia de SQLServerProductRepository.
func NewSQLServerProductRepository(db *sql.DB) *SQLServerProductRepository {
	return &SQLServerProductRepository{db: db}
}

// btiene un producto por su ID desde SQL Server.
func (r *SQLServerProductRepository) GetByID(id string) (*models.Product, error) {
	product := &models.Product{}
	query := `SELECT CONVERT(CHAR(255),Id) AS 'char', Name, Description, Price, Stock, ImageURL, CreatedAt, UpdatedAt
	          FROM Products WHERE ID = @p1`
	err := r.db.QueryRow(query, id).Scan(
		&product.ID, &product.Name, &product.Description,
		&product.Price, &product.Stock, &product.ImageURL, &product.CreatedAt, &product.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("producto con ID '%s' no encontrado", id)
	}
	if err != nil {
		return nil, fmt.Errorf("error al obtener producto por ID: %w", err)
	}
	return product, nil
}

// obtiene todos los productos desde SQL Server.
func (r *SQLServerProductRepository) GetAll() ([]models.Product, error) {
	rows, err := r.db.Query(`SELECT CONVERT(CHAR(255),Id) AS 'char', Name, Description, Price, Stock, ImageURL, CreatedAt, UpdatedAt FROM Products`)
	if err != nil {
		return nil, fmt.Errorf("error al obtener todos los productos: %w", err)
	}
	defer rows.Close()

	var products []models.Product
	for rows.Next() {
		var p models.Product
		if err := rows.Scan(
			&p.ID, &p.Name, &p.Description,
			&p.Price, &p.Stock, &p.ImageURL, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("error al escanear fila de producto: %w", err)
		}
		products = append(products, p)
	}
	return products, nil
}

// inserta un nuevo producto en SQL Server.
func (r *SQLServerProductRepository) Create(product models.Product) error {
	query := `INSERT INTO Products (ID, Name, Description, Price, Stock, ImageURL, CreatedAt, UpdatedAt)
	          VALUES (@p1, @p2, @p3, @p4, @p5, @p6, @p7, @p8)`
	_, err := r.db.Exec(query,
		product.ID, product.Name, product.Description,
		product.Price, product.Stock, product.ImageURL, product.CreatedAt, product.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("error al crear producto: %w", err)
	}
	return nil
}

// actualiza un producto existente en SQL Server.
func (r *SQLServerProductRepository) Update(product models.Product) error {
	query := `UPDATE Products
	          SET Name = @p1, Description = @p2, Price = @p3, Stock = @p4, ImageURL = @p5, UpdatedAt = @p6
	          WHERE ID = @p7`
	res, err := r.db.Exec(query,
		product.Name, product.Description, product.Price,
		product.Stock, product.ImageURL, time.Now(), product.ID,
	)
	if err != nil {
		return fmt.Errorf("error al actualizar producto: %w", err)
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("producto con ID '%s' no encontrado para actualizar", product.ID)
	}
	return nil
}

func (r *SQLServerProductRepository) UpdateProductStock(productID string, quantity int) error {
	query := `UPDATE Products SET Stock = Stock + @p1, UpdatedAt = @p2 WHERE ID = @p3`
	_, err := r.db.Exec(query, quantity, time.Now(), productID)
	if err != nil {
		return fmt.Errorf("error al actualizar stock del producto %s: %w", productID, err)
	}
	return nil
}

// Delete elimina un producto de SQL Server por su ID.
func (r *SQLServerProductRepository) Delete(id string) error {
	query := `DELETE FROM Products WHERE ID = @p1`
	res, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("error al eliminar producto: %w", err)
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("producto con ID '%s' no encontrado para eliminar", id)
	}
	return nil
}
