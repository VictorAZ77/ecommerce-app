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

// RegisterUser registra un nuevo usuario en el sistema.
func (s *AuthService) RegisterUser(username, password, role string) (*models.User, error) {
	if username == "" || password == "" || role == "" {
		return nil, fmt.Errorf("nombre de usuario, contraseña y rol no pueden estar vacíos")
	}

	// Verificar si el nombre de usuario ya existe
	_, err := s.userRepo.GetByUsername(username)
	if err == nil {
		return nil, fmt.Errorf("el nombre de usuario '%s' ya existe", username)
	}
	// Solo si el error es 'sql.ErrNoRows' significa que no existe y podemos continuar
	if err != nil && err.Error() != fmt.Sprintf("usuario con nombre de usuario '%s' no encontrado", username) {
		return nil, fmt.Errorf("error al verificar existencia de usuario: %w", err)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("falló al hashear la contraseña: %w", err)
	}

	user := models.User{
		ID:        uuid.New().String(),
		Username:  username,
		Password:  string(hashedPassword),
		Role:      role,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	createdUser, err := s.userRepo.Create(user)
	if err != nil {
		return nil, fmt.Errorf("falló al crear usuario en el repositorio: %w", err)
	}
	log.Printf("Usuario '%s' (ID: %s, Rol: %s) registrado exitosamente.", createdUser.Username, createdUser.ID, createdUser.Role)
	return createdUser, nil
}

// AuthenticateUser verifica las credenciales de un usuario.
func (s *AuthService) AuthenticateUser(username, password string) (*models.User, error) {
	if username == "" || password == "" {
		return nil, fmt.Errorf("nombre de usuario y contraseña no pueden estar vacíos")
	}

	user, err := s.userRepo.GetByUsername(username)
	if err != nil {
		return nil, fmt.Errorf("error al obtener usuario: %w", err)
	}

	// Compara la contraseña proporcionada con el hash almacenado
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return nil, fmt.Errorf("credenciales inválidas")
		}
		return nil, fmt.Errorf("error al comparar contraseñas: %w", err)
	}

	log.Printf("Usuario '%s' autenticado exitosamente (Rol: %s).", user.Username, user.Role)
	return user, nil
}
