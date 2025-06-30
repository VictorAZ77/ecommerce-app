package main

import (
	"backend/internal/config"
	"backend/internal/database"
	"backend/internal/repositories"
	"backend/internal/services"
	"backend/internal/web"
	"backend/internal/web/middleware"
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
	// Configura el SessionStore también para el paquete de middlewares
	middleware.SetSessionStoreMiddleware(web.GetSessionStore())

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
	// CAMBIO AQUÍ: Pasa authService y clientService a NewAuthHandlers
	authHandlers := web.NewAuthHandlers(authService, clientService)
	indexHandlers := web.NewIndexHandlers(productService)

	// Inicializar gorilla router
	router := mux.NewRouter()

	router.HandleFunc("/", indexHandlers.IndexPageHandler).Methods("GET")

	// Rutas de autenticación
	router.HandleFunc("/login", authHandlers.LoginPageHandler).Methods("GET")
	router.HandleFunc("/login", authHandlers.LoginHandler).Methods("POST")
	router.HandleFunc("/register", authHandlers.RegisterPageHandler).Methods("GET")
	router.HandleFunc("/register", authHandlers.RegisterHandler).Methods("POST")
	router.HandleFunc("/logout", authHandlers.LogoutHandler).Methods("GET", "POST") // Logout puede ser GET o POST

	// --- Rutas de páginas principales para usuarios normales ---
	// La página /index es para usuarios normales (e-commerce principal)
	router.HandleFunc("/index", indexHandlers.IndexPageHandler).Methods("GET")

	// --- Servir archivos estáticos (CSS, JS, imágenes) ---
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./internal/web/static"))))

	// --- Subrouter para la página de Administrador (/home) ---
	adminHomeRouter := router.PathPrefix("/home").Subrouter()
	adminHomeRouter.Use(middleware.AuthRequired)
	adminHomeRouter.Use(middleware.RoleRequired("admin"))
	adminHomeRouter.HandleFunc("", indexHandlers.AdminHomeHandler).Methods("GET")

	// --- Subrouter para rutas que requieren autenticación general ---
	authenticatedRouter := router.PathPrefix("/").Subrouter()
	authenticatedRouter.Use(middleware.AuthRequired)

	// Rutas de Clientes (accesibles si está autenticado)
	authenticatedRouter.HandleFunc("/clients", clientHandlers.ListClientsHandler).Methods("GET")
	authenticatedRouter.HandleFunc("/clients/new", clientHandlers.CreateClientPageHandler).Methods("GET")
	authenticatedRouter.HandleFunc("/clients/create", clientHandlers.CreateClientHandler).Methods("POST")
	authenticatedRouter.HandleFunc("/clients/edit/{id}", clientHandlers.EditClientPageHandler).Methods("GET")
	authenticatedRouter.HandleFunc("/clients/update/{id}", clientHandlers.UpdateClientHandler).Methods("POST")
	authenticatedRouter.HandleFunc("/clients/delete/{id}", clientHandlers.DeleteClientHandler).Methods("POST", "DELETE")
	// Rutas de Productos (accesibles si está autenticado)
	authenticatedRouter.HandleFunc("/products", productHandlers.ListProductsHandler).Methods("GET")
	authenticatedRouter.HandleFunc("/products/new", productHandlers.CreateProductPageHandler).Methods("GET")
	authenticatedRouter.HandleFunc("/products/create", productHandlers.CreateProductHandler).Methods("POST")
	authenticatedRouter.HandleFunc("/products/edit/{id}", productHandlers.EditProductPageHandler).Methods("GET")
	authenticatedRouter.HandleFunc("/products/update/{id}", productHandlers.UpdateProductHandler).Methods("POST")
	authenticatedRouter.HandleFunc("/products/delete/{id}", productHandlers.DeleteProductHandler).Methods("POST", "DELETE")

	// Rutas de Pedidos (listado, detalles, creación)
	authenticatedRouter.HandleFunc("/orders", orderHandlers.ListOrdersHandler).Methods("GET")
	authenticatedRouter.HandleFunc("/orders/new", orderHandlers.CreateOrderPageHandler).Methods("GET")
	authenticatedRouter.HandleFunc("/orders/{id}", orderHandlers.GetOrderDetailsHandler).Methods("GET")
	authenticatedRouter.HandleFunc("/orders/create", orderHandlers.CreateOrderHandler).Methods("POST")
	authenticatedRouter.HandleFunc("/orders/{id}", orderHandlers.DeleteOrderHandler).Methods("POST", "DELETE")

	authenticatedRouter.HandleFunc("/my-orders", orderHandlers.UserOrdersHandler).Methods("GET")
	// --- Subrouter para rutas que requieren el rol de 'admin' ---
	adminRouter := router.PathPrefix("/admin").Subrouter()
	adminRouter.Use(middleware.AuthRequired)
	adminRouter.Use(middleware.RoleRequired("admin"))

	// Ejemplo de ruta de administrador: Solo admins pueden eliminar pedidos
	adminRouter.HandleFunc("/orders/{id}", orderHandlers.DeleteOrderHandler).Methods("POST", "DELETE")

	log.Printf("Servidor escuchando en el puerto %s...", cfg.ServerPort)
	log.Fatal(http.ListenAndServe(":"+cfg.ServerPort, router))
}
