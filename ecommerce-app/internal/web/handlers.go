package web

import (
	"backend/internal/models"
	"backend/internal/services"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
)

var templates = make(map[string]*template.Template)

var customFuncs = template.FuncMap{
	"mul": func(a, b float64) float64 {
		return a * b
	},
	"toFloat64": func(v interface{}) float64 {
		switch f := v.(type) {
		case int:
			return float64(f)
		case int8:
			return float64(f)
		case int16:
			return float64(f)
		case int32:
			return float64(f)
		case int64:
			return float64(f)
		case float32:
			return float64(f)
		case float64:
			return f
		case string:
			parsedF, err := strconv.ParseFloat(f, 64)
			if err != nil {
				log.Printf("Advertencia: toFloat64 no pudo parsear string a float64: '%s', error: %v", f, err)
				return 0.0
			}
			return parsedF
		default:
			log.Printf("Advertencia: toFloat64 recibió un tipo inesperado: %T", v)
			return 0.0
		}
	},
}

// carga todas las plantillas HTML al inicio de la aplicación.
func LoadTemplates() {
	basePath := filepath.Join("internal", "web", "templates")

	parseTemplate := func(templateNameInSet string, files ...string) *template.Template {
		paths := make([]string, len(files))
		for i, file := range files {
			paths[i] = filepath.Join(basePath, file)
		}
		return template.Must(template.New(templateNameInSet).Funcs(customFuncs).ParseFiles(paths...))
	}

	// Plantilla de inicio
	templates["home"] = parseTemplate("home.html", "home.html")

	// Cargar plantillas de clientes
	templates["client-list"] = parseTemplate("list.html", filepath.Join("clients", "list.html"))
	templates["client-create-form"] = parseTemplate("create_form.html", filepath.Join("clients", "create_form.html"))

	// Cargar plantillas de productos
	templates["product-list"] = parseTemplate("list.html", filepath.Join("products", "list.html"))
	templates["product-create-form"] = parseTemplate("create_form.html", filepath.Join("products", "create_form.html"))

	// Cargar plantillas de pedidos
	templates["order-list"] = parseTemplate("list.html", filepath.Join("orders", "list.html"))
	templates["order-details"] = parseTemplate("details.html", filepath.Join("orders", "details.html"))
	templates["order-create-form"] = parseTemplate("create_form.html", filepath.Join("orders", "create_form.html"))

	// Cargar plantillas de Autenticación
	templates["login"] = parseTemplate("login.html", filepath.Join("auth", "login.html"))
	templates["register"] = parseTemplate("register.html", filepath.Join("auth", "register.html"))

	log.Println("Plantillas HTML cargadas exitosamente.")
}

var store *sessions.CookieStore

// SetSessionStore inicializa la tienda de sesiones. Esta función debe ser llamada desde main.go.
func SetSessionStore(key []byte) {
	store = sessions.NewCookieStore(key)
	store.Options = &sessions.Options{
		Path:     "/",                  // La cookie es válida para toda la aplicación
		MaxAge:   86400 * 7,            // 7 días (en segundos). Define cuánto tiempo durará la sesión.
		HttpOnly: true,                 // Previene el acceso a la cookie desde JavaScript
		Secure:   false,                // ¡IMPORTANTE! PONER A 'true' EN PRODUCCIÓN SI USAS HTTPS
		SameSite: http.SameSiteLaxMode, // Protege contra ataques CSRF básicos
	}
	log.Println("Tienda de sesiones configurada.")
}

// ClientHandlers agrupa los manejadores HTTP relacionados con los clientes.
type ClientHandlers struct {
	clientService *services.ClientService
}

// NewClientHandlers crea una nueva instancia de ClientHandlers.
func NewClientHandlers(cs *services.ClientService) *ClientHandlers {
	return &ClientHandlers{clientService: cs}
}

// ListClientsHandler muestra la lista de clientes.
func (h *ClientHandlers) ListClientsHandler(w http.ResponseWriter, r *http.Request) {
	clients, err := h.clientService.GetAllClients()
	if err != nil {
		log.Printf("Error al obtener clientes en ListClientsHandler: %v", err)
		http.Error(w, "Error al cargar clientes", http.StatusInternalServerError)
		return
	}

	tmpl, ok := templates["client-list"]
	if !ok {
		http.Error(w, "Plantilla 'client-list' no encontrada", http.StatusInternalServerError)
		return
	}

	data := struct {
		Title   string
		Clients []models.Client
	}{
		Title:   "Listado de Clientes",
		Clients: clients,
	}
	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("Error al renderizar plantilla de clientes: %v", err)
		http.Error(w, "Error interno del servidor al renderizar plantilla", http.StatusInternalServerError)
	}
}

// CreateClientPageHandler muestra el formulario para crear un cliente (GET).
func (h *ClientHandlers) CreateClientPageHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, ok := templates["client-create-form"]
	if !ok {
		http.Error(w, "Plantilla 'client-create-form' no encontrada", http.StatusInternalServerError)
		return
	}
	data := struct{ Title string }{Title: "Crear Nuevo Cliente"}
	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("Error al renderizar formulario de creación de cliente: %v", err)
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
	}
}

// CreateClientHandler procesa la creación de un cliente (POST).
func (h *ClientHandlers) CreateClientHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}
	name := r.FormValue("name")
	email := r.FormValue("email")
	address := r.FormValue("address")

	_, err := h.clientService.CreateClient(name, email, address)
	if err != nil {
		log.Printf("Error al crear cliente: %v", err)
		http.Error(w, "Error al crear cliente: "+err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/clients", http.StatusSeeOther)
}

type ProductHandlers struct {
	productService *services.ProductService
}

// NewProductHandlers crea una nueva instancia de ProductHandlers.
func NewProductHandlers(ps *services.ProductService) *ProductHandlers {
	return &ProductHandlers{productService: ps}
}

// ListProductsHandler muestra la lista de todos los productos.
func (h *ProductHandlers) ListProductsHandler(w http.ResponseWriter, r *http.Request) {
	products, err := h.productService.GetAllProducts()
	if err != nil {
		log.Printf("Error al obtener productos: %v", err)
		http.Error(w, "Error al cargar productos", http.StatusInternalServerError)
		return
	}

	tmpl, ok := templates["product-list"]
	if !ok {
		http.Error(w, "Plantilla 'product-list' no encontrada", http.StatusInternalServerError)
		return
	}

	data := struct {
		Title    string
		Products []models.Product
	}{
		Title:    "Listado de Productos",
		Products: products,
	}

	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("Error al renderizar plantilla de productos: %v", err)
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
	}
}

// CreateProductPageHandler muestra el formulario HTML para crear un nuevo producto.
func (h *ProductHandlers) CreateProductPageHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, ok := templates["product-create-form"]
	if !ok {
		http.Error(w, "Plantilla 'product-create-form' no encontrada", http.StatusInternalServerError)
		return
	}

	data := struct{ Title string }{Title: "Crear Nuevo Producto"}
	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("Error al renderizar formulario de creación de producto: %v", err)
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
	}
}

// CreateProductHandler procesa el envío del formulario para crear un producto (POST).
func (h *ProductHandlers) CreateProductHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	name := r.FormValue("name")
	description := r.FormValue("description")
	priceStr := r.FormValue("price")
	stockStr := r.FormValue("stock")

	price, err := strconv.ParseFloat(priceStr, 64)
	if err != nil {
		http.Error(w, "Precio inválido", http.StatusBadRequest)
		return
	}

	stock, err := strconv.Atoi(stockStr)
	if err != nil {
		http.Error(w, "Stock inválido", http.StatusBadRequest)
		return
	}

	_, err = h.productService.CreateProduct(name, description, price, stock)
	if err != nil {
		log.Printf("Error al crear producto: %v", err)
		http.Error(w, "Error al crear producto: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/products", http.StatusSeeOther)
}

// OrderHandlers agrupa los manejadores HTTP relacionados con los pedidos.
type OrderHandlers struct {
	orderService   *services.OrderService
	clientService  *services.ClientService
	productService *services.ProductService
}

// NewOrderHandlers crea una nueva instancia de OrderHandlers.
func NewOrderHandlers(
	os *services.OrderService,
	cs *services.ClientService,
	ps *services.ProductService,
) *OrderHandlers {
	return &OrderHandlers{
		orderService:   os,
		clientService:  cs,
		productService: ps,
	}
}

// HomeHandler maneja la ruta principal ("/") y muestra el estado de la sesión.
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "session-name")
	if err != nil {
		log.Printf("Error al obtener sesión en HomeHandler: %v", err)
	}

	data := struct {
		Title           string
		IsAuthenticated bool
		Username        string
		Role            string
	}{
		Title:           "Bienvenido a tu Tienda",
		IsAuthenticated: false,
		Username:        "",
		Role:            "",
	}

	if auth, ok := session.Values["authenticated"].(bool); ok && auth {
		data.IsAuthenticated = true
		if username, ok := session.Values["username"].(string); ok {
			data.Username = username
		}
		if role, ok := session.Values["userRole"].(string); ok {
			data.Role = role
		}
	}

	tmpl, ok := templates["home"]
	if !ok {
		log.Printf("Plantilla 'home' no encontrada en el mapa de templates.")
		http.Error(w, "Plantilla de inicio no encontrada", http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("Error al renderizar plantilla de inicio: %v", err)
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}
}

// ListOrdersHandler muestra la lista de todos los pedidos.
func (h *OrderHandlers) ListOrdersHandler(w http.ResponseWriter, r *http.Request) {
	orders, err := h.orderService.GetAllOrders()
	if err != nil {
		log.Printf("Error al obtener pedidos: %v", err)
		http.Error(w, "Error al cargar pedidos", http.StatusInternalServerError)
		return
	}

	tmpl, ok := templates["order-list"]
	if !ok {
		http.Error(w, "Plantilla 'order-list' no encontrada", http.StatusInternalServerError)
		return
	}

	data := struct {
		Title  string
		Orders []models.Order
	}{
		Title:  "Listado de Pedidos",
		Orders: orders,
	}

	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("Error al renderizar plantilla de pedidos: %v", err)
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
	}
}

// GetOrderDetailsHandler muestra los detalles de un pedido específico.
func (h *OrderHandlers) GetOrderDetailsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderID := vars["id"]
	log.Printf("DEBUG: Intentando obtener detalles para OrderID: %s", orderID)
	if orderID == "" {
		http.Error(w, "ID de pedido no proporcionado", http.StatusBadRequest)
		return
	}

	order, items, err := h.orderService.GetOrderWithDetailsByID(orderID)
	if err != nil {
		log.Printf("Error al obtener detalles del pedido %s: %v", orderID, err)
		http.Error(w, "Pedido no encontrado o error al cargar detalles", http.StatusInternalServerError)
		return
	}

	tmpl, ok := templates["order-details"]
	if !ok {
		log.Printf("Error: Plantilla 'order-details' no encontrada en el mapa de templates.")
		http.Error(w, "Plantilla 'order-details' no encontrada", http.StatusInternalServerError)
		return
	}

	data := struct {
		Title string
		Order *models.Order
		Items []models.OrderItem
	}{
		Title: fmt.Sprintf("Detalles del Pedido #%s", order.ID),
		Order: order,
		Items: items,
	}

	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("Error al renderizar plantilla de detalles de pedido: %v", err)
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
	}
}

// CreateOrderPageHandler muestra el formulario HTML para crear un nuevo pedido (GET).
func (h *OrderHandlers) CreateOrderPageHandler(w http.ResponseWriter, r *http.Request) {
	clients, err := h.clientService.GetAllClients()
	if err != nil {
		log.Printf("Error al cargar clientes para formulario de pedido: %v", err)
		http.Error(w, "Error al cargar clientes", http.StatusInternalServerError)
		return
	}
	products, err := h.productService.GetAllProducts()
	if err != nil {
		log.Printf("Error al cargar productos para formulario de pedido: %v", err)
		http.Error(w, "Error al cargar productos", http.StatusInternalServerError)
		return
	}

	tmpl, ok := templates["order-create-form"]
	if !ok {
		http.Error(w, "Plantilla 'order-create-form' no encontrada", http.StatusInternalServerError)
		return
	}

	data := struct {
		Title    string
		Clients  []models.Client
		Products []models.Product
	}{
		Title:    "Crear Nuevo Pedido",
		Clients:  clients,
		Products: products,
	}

	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("Error al renderizar formulario de creación de pedido: %v", err)
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
	}
}

// CreateOrderHandler procesa el envío del formulario para crear un pedido (POST).
func (h *OrderHandlers) CreateOrderHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	clientID := r.FormValue("clientID")
	productQuantities := make(map[string]int)

	for _, productID := range r.Form["productID"] {
		quantityStr := r.FormValue("quantity_" + productID)
		quantity, err := strconv.Atoi(quantityStr)
		if err != nil || quantity <= 0 {
			http.Error(w, fmt.Sprintf("Cantidad inválida para el producto %s", productID), http.StatusBadRequest)
			return
		}
		productQuantities[productID] = quantity
	}

	// Verificar si no se seleccionó ningún producto
	if len(productQuantities) == 0 {
		http.Error(w, "Debe seleccionar al menos un producto para el pedido.", http.StatusBadRequest)
		return
	}

	newOrder, err := h.orderService.CreateOrder(clientID, productQuantities)
	if err != nil {
		log.Printf("Error al crear pedido: %v", err)
		http.Error(w, "Error al crear pedido: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/orders/%s", newOrder.ID), http.StatusSeeOther)
}

func (h *OrderHandlers) DeleteOrderHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		if r.FormValue("_method") != "DELETE" {
			http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
			return
		}
	} else if r.Method != http.MethodDelete {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	vars := mux.Vars(r)
	orderID := vars["id"]

	if orderID == "" {
		http.Error(w, "ID de pedido no proporcionado", http.StatusBadRequest)
		return
	}

	err := h.orderService.DeleteOrder(orderID)
	if err != nil {
		log.Printf("Error al eliminar pedido %s: %v", orderID, err)
		http.Error(w, "Error al eliminar pedido: "+err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/orders", http.StatusSeeOther)
}

// AuthHandlers maneja las peticiones relacionadas con la autenticación de usuarios.
type AuthHandlers struct {
	authService *services.AuthService
}

// NewAuthHandlers crea e inicializa AuthHandlers.
func NewAuthHandlers(authService *services.AuthService) *AuthHandlers {
	return &AuthHandlers{
		authService: authService,
	}
}

// LoginPageHandler muestra el formulario de login.
func (h *AuthHandlers) LoginPageHandler(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "session-name")
	if err != nil {
		log.Printf("Error al obtener sesión en LoginPageHandler: %v", err)
		// No fatal, podemos seguir sin sesión si hay un error
	}
	if session.Values["authenticated"] == true {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	tmpl, ok := templates["login"]
	if !ok {
		log.Printf("Plantilla 'login' no encontrada en el mapa de templates.")
		http.Error(w, "Plantilla de login no encontrada", http.StatusInternalServerError)
		return
	}

	data := struct {
		ErrorMessage string
	}{
		ErrorMessage: "",
	}
	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("Error al renderizar formulario de login: %v", err)
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
	}
}

// LoginHandler procesa el envío del formulario de login.
func (h *AuthHandlers) LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	if username == "" || password == "" {
		h.renderLoginFormWithError(w, "Nombre de usuario y contraseña son obligatorios.")
		return
	}

	user, err := h.authService.AuthenticateUser(username, password)
	if err != nil {
		log.Printf("Fallo de autenticación para '%s': %v", username, err)
		h.renderLoginFormWithError(w, "Credenciales inválidas.")
		return
	}

	// Obtener la sesión y establecer los valores
	session, err := store.Get(r, "session-name")
	if err != nil {
		log.Printf("Error al obtener sesión para LoginHandler: %v", err)
		http.Error(w, "Error de sesión", http.StatusInternalServerError)
		return
	}

	session.Values["authenticated"] = true
	session.Values["userID"] = user.ID
	session.Values["username"] = user.Username
	session.Values["userRole"] = user.Role

	// Guardar la sesión para que la cookie sea enviada al cliente
	err = session.Save(r, w)
	if err != nil {
		log.Printf("Error al guardar sesión: %v", err)
		http.Error(w, "Error de sesión", http.StatusInternalServerError)
		return
	}

	log.Printf("Usuario '%s' (Rol: %s) ha iniciado sesión exitosamente.", user.Username, user.Role)
	// Redirección: ambos roles van al home como acordamos
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// renderLoginFormWithError es una función auxiliar para renderizar el formulario de login con un mensaje de error.
func (h *AuthHandlers) renderLoginFormWithError(w http.ResponseWriter, msg string) {
	tmpl, ok := templates["login"]
	if !ok {
		log.Printf("Error: Plantilla 'login' no encontrada para renderizar con error.")
		http.Error(w, "Error interno del servidor: plantilla no encontrada.", http.StatusInternalServerError)
		return
	}
	data := struct {
		ErrorMessage string
	}{
		ErrorMessage: msg,
	}
	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("Error al renderizar formulario de login con error: %v", err)
		http.Error(w, "Error interno del servidor.", http.StatusInternalServerError)
	}
}

// RegisterPageHandler muestra el formulario de registro.
func (h *AuthHandlers) RegisterPageHandler(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "session-name")
	if err != nil {
		log.Printf("Error al obtener sesión en RegisterPageHandler: %v", err)
	}
	if session.Values["authenticated"] == true {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	tmpl, ok := templates["register"]
	if !ok {
		log.Printf("Plantilla 'register' no encontrada en el mapa de templates.")
		http.Error(w, "Plantilla de registro no encontrada", http.StatusInternalServerError)
		return
	}
	data := struct {
		ErrorMessage string
	}{
		ErrorMessage: "",
	}
	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("Error al renderizar formulario de registro: %v", err)
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
	}
}

// RegisterHandler procesa el envío del formulario de registro.
func (h *AuthHandlers) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	if username == "" || password == "" {
		h.renderRegisterFormWithError(w, "Nombre de usuario y contraseña son obligatorios.")
		return
	}

	// Por defecto, asignar el rol "user" a los nuevos registros
	user, err := h.authService.RegisterUser(username, password, "user")
	if err != nil {
		log.Printf("Error al registrar usuario '%s': %v", username, err)
		h.renderRegisterFormWithError(w, fmt.Sprintf("Error al registrar: %v", err.Error()))
		return
	}

	// Iniciar sesión automáticamente después del registro exitoso
	session, err := store.Get(r, "session-name")
	if err != nil {
		log.Printf("Error al obtener sesión después de registro: %v", err)
		http.Error(w, "Error de sesión", http.StatusInternalServerError)
		return
	}
	session.Values["authenticated"] = true
	session.Values["userID"] = user.ID
	session.Values["username"] = user.Username
	session.Values["userRole"] = user.Role

	err = session.Save(r, w)
	if err != nil {
		log.Printf("Error al guardar sesión después de registro: %v", err)
		http.Error(w, "Error de sesión", http.StatusInternalServerError)
		return
	}

	log.Printf("Nuevo usuario '%s' registrado y logueado automáticamente.", user.Username)
	http.Redirect(w, r, "/", http.StatusSeeOther) // Redirigir al home
}

// renderRegisterFormWithError es una función auxiliar para renderizar el formulario de registro con un mensaje de error.
func (h *AuthHandlers) renderRegisterFormWithError(w http.ResponseWriter, msg string) {
	tmpl, ok := templates["register"]
	if !ok {
		log.Printf("Error: Plantilla 'register' no encontrada para renderizar con error.")
		http.Error(w, "Error interno del servidor: plantilla no encontrada.", http.StatusInternalServerError)
		return
	}
	data := struct {
		ErrorMessage string
	}{
		ErrorMessage: msg,
	}
	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("Error al renderizar formulario de registro con error: %v", err)
		http.Error(w, "Error interno del servidor.", http.StatusInternalServerError)
	}
}

// LogoutHandler maneja el cierre de sesión.
func (h *AuthHandlers) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "session-name")
	if err != nil {
		log.Printf("Error al obtener sesión para logout: %v", err)

	}

	// Borrar los valores de la sesión y establecer MaxAge a -1 para eliminar la cookie.
	session.Values["authenticated"] = false
	session.Values["userID"] = nil
	session.Values["username"] = nil
	session.Values["userRole"] = nil
	session.Options.MaxAge = -1

	err = session.Save(r, w)
	if err != nil {
		log.Printf("Error al guardar sesión para logout: %v", err)
		http.Error(w, "Error al cerrar sesión", http.StatusInternalServerError)
		return
	}

	log.Println("Usuario ha cerrado sesión.")
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
