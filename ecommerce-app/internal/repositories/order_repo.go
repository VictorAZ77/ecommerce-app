package repositories

import (
	"backend/internal/models"
	"database/sql"
	"fmt"
	"time"
)

// define las operaciones de persistencia para Order y OrderItem.
type OrderRepository interface {
	GetOrderByID(id string) (*models.Order, error)
	GetOrderItemsByOrderID(orderID string) ([]models.OrderItem, error)
	GetAllOrders() ([]models.Order, error)
	CreateOrder(order models.Order, items []models.OrderItem) error
	UpdateOrder(order models.Order) error
	DeleteOrder(tx *sql.Tx, id string) error
	GetDB() *sql.DB
}

// implementación de OrderRepository para SQL Server.
type SQLServerOrderRepository struct {
	db *sql.DB
}

// crea una nueva instancia de SQLServerOrderRepository.
func NewSQLServerOrderRepository(dbConn *sql.DB) *SQLServerOrderRepository {
	return &SQLServerOrderRepository{db: dbConn}
}

// obtiene un pedido por su ID.
func (r *SQLServerOrderRepository) GetOrderByID(id string) (*models.Order, error) {
	order := &models.Order{}
	query := `SELECT CONVERT(CHAR(255),Id) AS 'char',CONVERT(CHAR(255),ClientID) AS 'char', OrderDate, TotalAmount, Status, CreatedAt, UpdatedAt
	          FROM Orders WHERE ID = @p1`
	err := r.db.QueryRow(query, id).Scan(
		&order.ID, &order.ClientID, &order.OrderDate,
		&order.TotalAmount, &order.Status, &order.CreatedAt, &order.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("pedido con ID '%s' no encontrado", id)
	}
	if err != nil {
		return nil, fmt.Errorf("error al obtener pedido por ID: %w", err)
	}
	return order, nil
}

// implementa el método de la interfaz para devolver la conexión a la base de datos.
func (r *SQLServerOrderRepository) GetDB() *sql.DB {
	return r.db
}

// obtiene todos los ítems de un pedido específico.
func (r *SQLServerOrderRepository) GetOrderItemsByOrderID(orderID string) ([]models.OrderItem, error) {
	rows, err := r.db.Query(`SELECT CONVERT(CHAR(255),Id) AS 'char', CONVERT(CHAR(255),OrderID) AS 'char', CONVERT(CHAR(255),ProductID) AS 'char', Quantity, Price, CreatedAt, UpdatedAt
	                         FROM OrderItems WHERE OrderID = @p1`, orderID)
	if err != nil {
		return nil, fmt.Errorf("error al obtener ítems del pedido: %w", err)
	}
	defer rows.Close()

	var items []models.OrderItem
	for rows.Next() {
		var item models.OrderItem
		if err := rows.Scan(
			&item.ID, &item.OrderID, &item.ProductID, &item.Quantity,
			&item.Price, &item.CreatedAt, &item.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("error al escanear fila de ítem de pedido: %w", err)
		}
		items = append(items, item)
	}
	return items, nil
}

// obtiene todos los pedidos.
func (r *SQLServerOrderRepository) GetAllOrders() ([]models.Order, error) {
	rows, err := r.db.Query(`SELECT CONVERT(CHAR(255),Id) AS 'char', CONVERT(CHAR(255),ClientID) AS 'char', OrderDate, TotalAmount, Status, CreatedAt, UpdatedAt FROM Orders`)
	if err != nil {
		return nil, fmt.Errorf("error al obtener todos los pedidos: %w", err)
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		var o models.Order
		if err := rows.Scan(
			&o.ID, &o.ClientID, &o.OrderDate,
			&o.TotalAmount, &o.Status, &o.CreatedAt, &o.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("error al escanear fila de pedido: %w", err)
		}
		orders = append(orders, o)
	}
	return orders, nil
}

// crea un pedido y sus ítems asociados dentro de una transacción.
func (r *SQLServerOrderRepository) CreateOrder(order models.Order, items []models.OrderItem) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("error al iniciar transacción para crear pedido: %w", err)
	}
	defer tx.Rollback()

	// Insertar el pedido principal
	orderQuery := `INSERT INTO Orders (ID, ClientID, OrderDate, TotalAmount, Status, CreatedAt, UpdatedAt)
	               VALUES (@p1, @p2, @p3, @p4, @p5, @p6, @p7)`
	_, err = tx.Exec(orderQuery,
		order.ID, order.ClientID, order.OrderDate,
		order.TotalAmount, order.Status, order.CreatedAt, order.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("error al insertar pedido principal: %w", err)
	}

	// Insertar los ítems del pedido
	itemQuery := `INSERT INTO OrderItems (ID, OrderID, ProductID, Quantity, Price, CreatedAt, UpdatedAt)
	              VALUES (@p1, @p2, @p3, @p4, @p5, @p6, @p7)`
	for _, item := range items {
		_, err = tx.Exec(itemQuery,
			item.ID, item.OrderID, item.ProductID, item.Quantity,
			item.Price, item.CreatedAt, item.UpdatedAt,
		)
		if err != nil {
			return fmt.Errorf("error al insertar ítem de pedido '%s': %w", item.ID, err)
		}
	}

	return tx.Commit()
}

// actualiza un pedido existente.
func (r *SQLServerOrderRepository) UpdateOrder(order models.Order) error {
	query := `UPDATE Orders
	          SET ClientID = @p1, OrderDate = @p2, TotalAmount = @p3, Status = @p4, UpdatedAt = @p5
	          WHERE ID = @p6`
	res, err := r.db.Exec(query,
		order.ClientID, order.OrderDate, order.TotalAmount,
		order.Status, time.Now(), order.ID,
	)
	if err != nil {
		return fmt.Errorf("error al actualizar pedido: %w", err)
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("pedido con ID '%s' no encontrado para actualizar", order.ID)
	}
	return nil
}

// elimina un pedido y todos sus ítems asociados.
func (r *SQLServerOrderRepository) DeleteOrder(tx *sql.Tx, orderID string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("error al iniciar transacción para eliminar pedido: %w", err)
	}
	defer tx.Rollback()

	// Eliminar ítems del pedido primero
	_, err = tx.Exec(`DELETE FROM OrderItems WHERE OrderID = @p1`, orderID)
	if err != nil {
		return fmt.Errorf("error al eliminar ítems del pedido '%s': %w", orderID, err)
	}

	// Eliminar el pedido principal
	res, err := tx.Exec(`DELETE FROM Orders WHERE ID = @p1`, orderID)
	if err != nil {
		return fmt.Errorf("error al eliminar pedido principal '%s': %w", orderID, err)
	}

	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("pedido con ID '%s' no encontrado para eliminar", orderID)
	}

	return tx.Commit()
}
