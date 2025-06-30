package services

import (
	"backend/internal/models"
	"backend/internal/repositories"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ProductService define la lógica de negocio para los productos.
type ProductService struct {
	repo repositories.ProductRepository
}

// NewProductService crea una nueva instancia de ProductService.
func NewProductService(repo repositories.ProductRepository) *ProductService {
	return &ProductService{repo: repo}
}

// CreateProduct crea un nuevo producto, realizando validaciones si es necesario.
func (s *ProductService) CreateProduct(name, description string, price float64, stock int, imageURL string) (*models.Product, error) {
	if name == "" {
		return nil, fmt.Errorf("el nombre del producto es obligatorio")
	}
	if price <= 0 {
		return nil, fmt.Errorf("el precio debe ser mayor que cero")
	}
	if stock < 0 {
		return nil, fmt.Errorf("el stock no puede ser negativo")
	}

	newProduct := models.Product{
		ID:          uuid.New().String(),
		Name:        name,
		Description: description,
		Price:       price,
		Stock:       stock,
		ImageURL:    imageURL,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.repo.Create(newProduct); err != nil {
		return nil, fmt.Errorf("falló la creación del producto en la DB: %w", err)
	}
	return &newProduct, nil
}

// GetProductByID obtiene un producto por su ID.
func (s *ProductService) GetProductByID(id string) (*models.Product, error) {
	if id == "" {
		return nil, fmt.Errorf("el ID del producto no puede estar vacío")
	}
	product, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("no se pudo obtener el producto: %w", err)
	}
	return product, nil
}

// GetAllProducts obtiene todos los productos.
func (s *ProductService) GetAllProducts() ([]models.Product, error) {
	products, err := s.repo.GetAll()
	if err != nil {
		return nil, fmt.Errorf("no se pudieron obtener los productos: %w", err)
	}
	return products, nil
}

// UpdateProduct actualiza un producto existente.
func (s *ProductService) UpdateProduct(id, name, description string, price float64, stock int, imageURL string) (*models.Product, error) {
	existingProduct, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("producto a actualizar no encontrado: %w", err)
	}

	if name != "" {
		existingProduct.Name = name
	}
	if description != "" {
		existingProduct.Description = description
	}
	if price > 0 {
		existingProduct.Price = price
	}
	if stock >= 0 {
		existingProduct.Stock = stock
	}
	if imageURL != "" {
		existingProduct.ImageURL = imageURL
	}
	existingProduct.UpdatedAt = time.Now()

	if err := s.repo.Update(*existingProduct); err != nil {
		return nil, fmt.Errorf("falló la actualización del producto en la DB: %w", err)
	}
	return existingProduct, nil
}

// DeleteProduct elimina un producto por su ID.
func (s *ProductService) DeleteProduct(id string) error {
	if id == "" {
		return fmt.Errorf("el ID del producto no puede estar vacío")
	}
	if err := s.repo.Delete(id); err != nil {
		return fmt.Errorf("falló la eliminación del producto: %w", err)
	}
	return nil
}

// IncreaseProductStock aumenta el stock de un producto específico.
func (s *ProductService) IncreaseProductStock(productID string, quantity int) error {
	if productID == "" {
		return fmt.Errorf("ID del producto no puede estar vacío")
	}
	if quantity <= 0 {
		return fmt.Errorf("la cantidad para aumentar stock debe ser positiva")
	}

	err := s.repo.UpdateProductStock(productID, quantity)
	if err != nil {
		return fmt.Errorf("falló al aumentar el stock del producto %s en %d unidades: %w", productID, quantity, err)
	}
	return nil
}
