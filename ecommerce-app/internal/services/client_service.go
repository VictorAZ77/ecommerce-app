package services

import (
	"backend/internal/models"
	"backend/internal/repositories"
	"fmt"
	"github.com/google/uuid"
	"time"
)

// ClientService ofrece la lógica de negocio para los clientes.
type ClientService struct {
	repo repositories.ClientRepository
}

// NewClientService crea una nueva instancia de ClientService.
func NewClientService(repo repositories.ClientRepository) *ClientService {
	return &ClientService{repo: repo}
}

// CreateClient crea un nuevo cliente con validaciones.
func (s *ClientService) CreateClient(name, email, address string) (*models.Client, error) {
	if name == "" || email == "" || address == "" {
		return nil, fmt.Errorf("name, email, and address are mandatory for the client")
	}

	_, err := s.repo.GetClientByEmail(email)
	if err == nil {
		return nil, fmt.Errorf("a client with this email already exists: %s", email)
	}

	if err.Error() != fmt.Sprintf("cliente con Email %s no encontrado", email) {
		return nil, fmt.Errorf("error checking for existing client by email: %w", err)
	}

	newClient := models.Client{
		ID:        uuid.New().String(), // Generar un UUID para el ID del cliente
		Name:      name,
		Email:     email,
		Address:   address,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.repo.Create(newClient); err != nil {
		return nil, fmt.Errorf("client creation failed in DB: %w", err)
	}
	return &newClient, nil
}

// GetClientByID obtiene un cliente por su ID.
func (s *ClientService) GetClientByID(id string) (*models.Client, error) {
	if id == "" {
		return nil, fmt.Errorf("client ID cannot be empty")
	}
	client, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("could not get client: %w", err)
	}
	return client, nil
}

func (s *ClientService) GetClientByEmail(email string) (*models.Client, error) {
	client, err := s.repo.GetClientByEmail(email)
	if err != nil {
		return nil, fmt.Errorf("fallo al obtener cliente por Email: %w", err)
	}
	return client, nil
}

// GetAllClients obtiene todos los clientes.
func (s *ClientService) GetAllClients() ([]models.Client, error) {
	clients, err := s.repo.GetAll()
	if err != nil {
		return nil, fmt.Errorf("could not get clients: %w", err)
	}
	return clients, nil
}

// UpdateClient actualiza un cliente existente.
func (s *ClientService) UpdateClient(id, name, email, address string) (*models.Client, error) {
	existingClient, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("client to update not found: %w", err)
	}

	// Se validan que todos los campos obligatorios estén presentes
	if name == "" || email == "" || address == "" {
		return nil, fmt.Errorf("name, email, and address are mandatory for update")
	}

	// Opcional: Si el email se está cambiando, verificar unicidad del nuevo email
	if email != existingClient.Email {
		_, err := s.repo.GetClientByEmail(email)
		if err == nil {
			return nil, fmt.Errorf("the new email '%s' is already associated with another client", email)
		}
		if err.Error() != fmt.Sprintf("cliente con Email %s no encontrado", email) {
			return nil, fmt.Errorf("error checking new email for existing client: %w", err)
		}
	}

	existingClient.Name = name
	existingClient.Email = email
	existingClient.Address = address
	existingClient.UpdatedAt = time.Now()

	if err := s.repo.Update(*existingClient); err != nil {
		return nil, fmt.Errorf("client update failed in DB: %w", err)
	}
	return existingClient, nil
}

// DeleteClient elimina un cliente por su ID.
func (s *ClientService) DeleteClient(id string) error {
	if id == "" {
		return fmt.Errorf("client ID cannot be empty")
	}
	if err := s.repo.Delete(id); err != nil {
		return fmt.Errorf("client deletion failed: %w", err)
	}
	return nil
}
