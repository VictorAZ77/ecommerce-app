package services

import (
	"backend/internal/models"
	"backend/internal/repositories"
	"fmt"
	"github.com/google/uuid"
	"log"
	"time"
)

type OrderService struct {
	orderRepo      repositories.OrderRepository
	productRepo    repositories.ProductRepository
	clientRepo     repositories.ClientRepository
	productService *ProductService
}

func NewOrderService(
	orderRepo repositories.OrderRepository,
	productRepo repositories.ProductRepository,
	clientRepo repositories.ClientRepository,
	productService *ProductService,
) *OrderService {
	return &OrderService{
		orderRepo:      orderRepo,
		productRepo:    productRepo,
		clientRepo:     clientRepo,
		productService: productService,
	}
}

// crea un nuevo pedido con sus ítems, verificando la existencia del cliente y el stock de productos.
func (s *OrderService) CreateOrder(clientID string, productQuantities map[string]int) (*models.Order, error) {
	if clientID == "" {
		return nil, fmt.Errorf("el ID del cliente no puede estar vacío")
	}
	if len(productQuantities) == 0 {
		return nil, fmt.Errorf("el pedido debe contener al menos un producto")
	}

	//Verificar que el cliente existe
	_, err := s.clientRepo.GetByID(clientID)
	if err != nil {
		return nil, fmt.Errorf("cliente con ID '%s' no encontrado: %w", clientID, err)
	}

	// Recopilar detalles de los productos y verificar stock
	var orderItems []models.OrderItem
	var totalAmount float64
	productsToUpdateStock := make(map[string]int)

	for productID, quantity := range productQuantities {
		if quantity <= 0 {
			return nil, fmt.Errorf("la cantidad del producto '%s' debe ser mayor que cero", productID)
		}

		product, err := s.productRepo.GetByID(productID)
		if err != nil {
			return nil, fmt.Errorf("producto con ID '%s' no encontrado: %w", productID, err)
		}

		if product.Stock < quantity {
			return nil, fmt.Errorf("stock insuficiente para el producto '%s'. Disponible: %d, Solicitado: %d", product.Name, product.Stock, quantity)
		}

		orderItems = append(orderItems, models.OrderItem{
			ID:        uuid.New().String(),
			OrderID:   "",
			ProductID: product.ID,
			Quantity:  quantity,
			Price:     product.Price,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		})
		totalAmount += product.Price * float64(quantity)
		productsToUpdateStock[product.ID] = product.Stock - quantity
	}

	//struct del pedido principal
	newOrder := models.Order{
		ID:          uuid.New().String(),
		ClientID:    clientID,
		OrderDate:   time.Now(),
		TotalAmount: totalAmount,
		Status:      "Pendiente",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Asignar el ID del pedido a los ítems del pedido
	for i := range orderItems {
		orderItems[i].OrderID = newOrder.ID
	}

	err = s.orderRepo.CreateOrder(newOrder, orderItems)
	if err != nil {
		return nil, fmt.Errorf("falló la creación del pedido en la DB: %w", err)
	}

	for productID, newStock := range productsToUpdateStock {
		product, _ := s.productRepo.GetByID(productID)
		product.Stock = newStock
		err := s.productRepo.Update(*product)
		if err != nil {
			return nil, fmt.Errorf("advertencia: pedido creado, pero falló la actualización de stock para producto '%s': %w", productID, err)
		}
	}

	return &newOrder, nil
}

// obtiene un pedido por su ID, incluyendo sus ítems.
func (s *OrderService) GetOrderWithDetailsByID(orderID string) (*models.Order, []models.OrderItem, error) {
	if orderID == "" {
		return nil, nil, fmt.Errorf("el ID del pedido no puede estar vacío")
	}

	order, err := s.orderRepo.GetOrderByID(orderID)
	if err != nil {
		return nil, nil, fmt.Errorf("no se pudo obtener el pedido: %w", err)
	}

	items, err := s.orderRepo.GetOrderItemsByOrderID(orderID)
	if err != nil {
		return nil, nil, fmt.Errorf("no se pudieron obtener los ítems del pedido: %w", err)
	}

	return order, items, nil
}

// obtiene todos los pedidos
func (s *OrderService) GetAllOrders() ([]models.Order, error) {
	orders, err := s.orderRepo.GetAllOrders()
	if err != nil {
		return nil, fmt.Errorf("no se pudieron obtener los pedidos: %w", err)
	}
	return orders, nil
}

// actualiza el estado de un pedido.
func (s *OrderService) UpdateOrderStatus(orderID, newStatus string) (*models.Order, error) {
	if orderID == "" || newStatus == "" {
		return nil, fmt.Errorf("ID del pedido y nuevo estado son obligatorios")
	}

	order, err := s.orderRepo.GetOrderByID(orderID)
	if err != nil {
		return nil, fmt.Errorf("pedido con ID '%s' no encontrado para actualizar estado: %w", orderID, err)
	}

	order.Status = newStatus
	order.UpdatedAt = time.Now()

	if err := s.orderRepo.UpdateOrder(*order); err != nil {
		return nil, fmt.Errorf("falló la actualización del estado del pedido: %w", err)
	}
	return order, nil
}

// elimina un pedido y sus ítems asociados.
func (s *OrderService) DeleteOrder(orderID string) error {
	if orderID == "" {
		return fmt.Errorf("el ID del pedido no puede estar vacío")
	}

	orderItems, err := s.orderRepo.GetOrderItemsByOrderID(orderID)
	if err != nil {
		return fmt.Errorf("falló al obtener ítems del pedido para revertir stock: %w", err)
	}

	tx, err := s.orderRepo.GetDB().Begin()
	if err != nil {
		return fmt.Errorf("falló al iniciar la transacción para eliminar pedido: %w", err)
	}

	defer func() {
		if r := recover(); r != nil {
			log.Printf("PANIC detectado durante la eliminación de pedido. Realizando Rollback: %v", r)
			tx.Rollback()
			panic(r)
		}
	}()

	//Revertir el stock para cada ítem del pedido dentro de la transacción
	for _, item := range orderItems {
		err := s.productService.IncreaseProductStock(item.ProductID, item.Quantity)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("falló al revertir stock para producto %s en pedido %s: %w", item.ProductID, orderID, err)
		}
	}

	if err := s.orderRepo.DeleteOrder(tx, orderID); err != nil {
		tx.Rollback()
		return fmt.Errorf("falló la eliminación final del pedido transaccional: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("falló el commit de la transacción al eliminar pedido: %w", err)
	}
	return nil
}
