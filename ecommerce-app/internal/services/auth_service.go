package services

import (
	"backend/internal/models"
	"backend/internal/repositories"
	"fmt"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"log"
	"time"
)

// AuthService maneja la lógica de autenticación y autorización de usuarios.
type AuthService struct {
	userRepo repositories.UserRepository
}

// NewAuthService crea una nueva instancia de AuthService.
func NewAuthService(ur repositories.UserRepository) *AuthService {
	return &AuthService{userRepo: ur}
}

// RegisterUser registra un nuevo usuario en el sistema, incluyendo su correo electrónico.
func (s *AuthService) RegisterUser(username, password, email, role string) (*models.User, error) {
	if username == "" || password == "" || email == "" || role == "" {
		return nil, fmt.Errorf("username, password, email, and role cannot be empty")
	}
	_, err := s.userRepo.GetByUsername(username)
	if err == nil {
		return nil, fmt.Errorf("username '%s' already exists. Please choose another", username)
	}
	if err.Error() != fmt.Sprintf("usuario con nombre de usuario '%s' no encontrado", username) { 
		return nil, fmt.Errorf("error checking for existing username: %w", err)
	}

	_, err = s.userRepo.GetByEmail(email)
	if err == nil {
		return nil, fmt.Errorf("email '%s' is already in use. Please use another or log in", email)
	}

	if err.Error() != fmt.Sprintf("usuario con email '%s' no encontrado", email) { 
		return nil, fmt.Errorf("error checking for existing email: %w", err)
	}

	// Hash the password before storing it
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create the User object with all fields, including Email
	user := models.User{
		ID:        uuid.New().String(), // Generate a UUID for the user ID
		Username:  username,
		Password:  string(hashedPassword), // Store the hash
		Email:     email,                  
		Role:      role,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Call the repository to create the user
	createdUser, err := s.userRepo.Create(user)
	if err != nil {
		return nil, fmt.Errorf("failed to create user in repository: %w", err)
	}
	log.Printf("User '%s' (ID: %s, Email: %s, Role: %s) successfully registered.", createdUser.Username, createdUser.ID, createdUser.Email, createdUser.Role)
	return createdUser, nil
}

// AuthenticateUser verifies a user's credentials.
func (s *AuthService) AuthenticateUser(username, password string) (*models.User, error) {
	if username == "" || password == "" {
		return nil, fmt.Errorf("username and password cannot be empty")
	}

	user, err := s.userRepo.GetByUsername(username)
	if err != nil {
		log.Printf("Error getting user '%s' for authentication: %v", username, err)
		return nil, fmt.Errorf("invalid credentials")
	}

	// Compare the provided password with the stored hash
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return nil, fmt.Errorf("invalid credentials")
		}
		log.Printf("Error comparing passwords for user '%s': %v", username, err)
		return nil, fmt.Errorf("internal authentication error") 
	}

	log.Printf("User '%s' successfully authenticated (Role: %s).", user.Username, user.Role)
	return user, nil
}
