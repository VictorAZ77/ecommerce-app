package main

import (
	"backend/internal/config"
	"backend/internal/database"
	"backend/internal/repositories"
	"backend/internal/services"
	"backend/internal/web"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

func main() {
	// Cargar configuración
	cfg := config.LoadConfig()

	// Inicializar la base de datos
	db, err := database.InitDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Error crítico: no se pudo inicializar la base de datos: %v", err)
	}
	defer db.Close()

	// Cargar plantillas HTML al inicio
	web.LoadTemplates()

	// Pasa la clave de sesión cargada desde la configuración a la función en el paquete 'web'.
	web.SetSessionStore(cfg.SessionKey)

	// Inicializar repositorios e inyectar la conexión a DB
	clientRepo := repositories.NewSQLServerClientRepository(db)
	productRepo := repositories.NewSQLServerProductRepository(db)
	orderRepo := repositories.NewSQLServerOrderRepository(db)
	userRepo := repositories.NewSQLServerUserRepository(db)

	// Inicializar servicios e inyectar los repositorios
	clientService := services.NewClientService(clientRepo)
	productService := services.NewProductService(productRepo)
	orderService := services.NewOrderService(orderRepo, productRepo, clientRepo, productService)
	authService := services.NewAuthService(userRepo)

	// Inicializar manejadores web e inyectar los servicios
	clientHandlers := web.NewClientHandlers(clientService)
	productHandlers := web.NewProductHandlers(productService)
	orderHandlers := web.NewOrderHandlers(orderService, clientService, productService)
	authHandlers := web.NewAuthHandlers(authService)

	// Inicializar gorilla
	router := mux.NewRouter()

	// Ruta de Inicio
	router.HandleFunc("/", web.HomeHandler).Methods("GET")

	// Rutas de Clientes
	router.HandleFunc("/clients", clientHandlers.ListClientsHandler).Methods("GET")
	router.HandleFunc("/clients/new", clientHandlers.CreateClientPageHandler).Methods("GET")
	router.HandleFunc("/clients/create", clientHandlers.CreateClientHandler).Methods("POST")

	// Rutas de Productos
	router.HandleFunc("/products", productHandlers.ListProductsHandler).Methods("GET")
	router.HandleFunc("/products/new", productHandlers.CreateProductPageHandler).Methods("GET")
	router.HandleFunc("/products/create", productHandlers.CreateProductHandler).Methods("POST")

	// Rutas de Pedidos
	router.HandleFunc("/orders", orderHandlers.ListOrdersHandler).Methods("GET")
	router.HandleFunc("/orders/new", orderHandlers.CreateOrderPageHandler).Methods("GET")
	router.HandleFunc("/orders/{id}", orderHandlers.GetOrderDetailsHandler).Methods("GET")
	router.HandleFunc("/orders/create", orderHandlers.CreateOrderHandler).Methods("POST")
	router.HandleFunc("/orders/{id}", orderHandlers.DeleteOrderHandler).Methods("POST", "DELETE")

	// Rutas de Auntenticacion
	router.HandleFunc("/login", authHandlers.LoginPageHandler).Methods("GET")
	router.HandleFunc("/login", authHandlers.LoginHandler).Methods("POST")
	router.HandleFunc("/logout", authHandlers.LogoutHandler).Methods("GET", "POST")

	// Rutas de Registro de Usuario ---
	router.HandleFunc("/register", authHandlers.RegisterPageHandler).Methods("GET")
	router.HandleFunc("/register", authHandlers.RegisterHandler).Methods("POST")

	log.Printf("Servidor escuchando en el puerto %s...", cfg.ServerPort)
	log.Fatal(http.ListenAndServe(":"+cfg.ServerPort, router))
}
