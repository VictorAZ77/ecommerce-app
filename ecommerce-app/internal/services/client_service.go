// internal/services/client_service.go
package services

import (
	"backend/internal/models"
	"backend/internal/repositories"
	"fmt"
	"github.com/google/uuid"
	"time"
)

// ofrece la lógica de negocio para los clientes.
type ClientService struct {
	repo repositories.ClientRepository
}

// crea una nueva instancia de ClientService.
func NewClientService(repo repositories.ClientRepository) *ClientService {
	return &ClientService{repo: repo}
}

// crea un nuevo cliente con validaciones.
func (s *ClientService) CreateClient(name, email, address string) (*models.Client, error) {
	if name == "" || email == "" {
		return nil, fmt.Errorf("nombre y email son obligatorios")
	}

	newClient := models.Client{
		ID:        uuid.New().String(), // Generar un UUID
		Name:      name,
		Email:     email,
		Address:   address,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.repo.Create(newClient); err != nil {
		return nil, fmt.Errorf("falló la creación del cliente en la DB: %w", err)
	}
	return &newClient, nil
}

// obtiene un cliente por su ID.
func (s *ClientService) GetClientByID(id string) (*models.Client, error) {
	if id == "" {
		return nil, fmt.Errorf("el ID del cliente no puede estar vacío")
	}
	client, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("no se pudo obtener el cliente: %w", err)
	}
	return client, nil
}

// obtiene todos los clientes.
func (s *ClientService) GetAllClients() ([]models.Client, error) {
	clients, err := s.repo.GetAll()
	if err != nil {
		return nil, fmt.Errorf("no se pudieron obtener los clientes: %w", err)
	}
	return clients, nil
}

// actualiza un cliente existente.
func (s *ClientService) UpdateClient(id, name, email, address string) (*models.Client, error) {
	existingClient, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("cliente a actualizar no encontrado: %w", err)
	}

	if name != "" {
		existingClient.Name = name
	}
	if email != "" {
		existingClient.Email = email
	}
	if address != "" {
		existingClient.Address = address
	}
	existingClient.UpdatedAt = time.Now()

	if err := s.repo.Update(*existingClient); err != nil {
		return nil, fmt.Errorf("falló la actualización del cliente en la DB: %w", err)
	}
	return existingClient, nil
}

// elimina un cliente por su ID.
func (s *ClientService) DeleteClient(id string) error {
	if id == "" {
		return fmt.Errorf("el ID del cliente no puede estar vacío")
	}
	if err := s.repo.Delete(id); err != nil {
		return fmt.Errorf("falló la eliminación del cliente: %w", err)
	}
	return nil
}
