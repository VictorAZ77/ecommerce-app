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

// LoadTemplates carga todas las plantillas HTML al inicio de la aplicación.
func LoadTemplates() {
	basePath := filepath.Join("internal", "web", "templates")

	parseTemplate := func(templateNameInSet string, files ...string) *template.Template {
		paths := make([]string, len(files))
		for i, file := range files {
			paths[i] = filepath.Join(basePath, file)
		}
		// Asegúrate de que el nombre del template sea el mismo que se usará en tmpl.Execute
		return template.Must(template.New(templateNameInSet).Funcs(customFuncs).ParseFiles(paths...))
	}

	// Plantilla de inicio (se mantiene para compatibilidad si hay otras rutas que la usan)
	templates["home"] = parseTemplate("home.html", "home.html")
	// ¡Nueva plantilla para la página principal tipo e-commerce!
	templates["index"] = parseTemplate("index.html", "index.html") // Asegúrate que el nombre "index" coincida con el archivo base

	// Cargar plantillas de clientes
	templates["client-list"] = parseTemplate("list.html", filepath.Join("clients", "list.html"))
	templates["client-create-form"] = parseTemplate("create_form.html", filepath.Join("clients", "create_form.html"))
	templates["client-edit-form"] = parseTemplate("edit_client.html", filepath.Join("clients", "edit_client.html"))

	// Cargar plantillas de productos
	templates["product-list"] = parseTemplate("list.html", filepath.Join("products", "list.html"))
	templates["product-create-form"] = parseTemplate("create_form.html", filepath.Join("products", "create_form.html"))
	templates["product-edit-form"] = parseTemplate("edit_product.html", filepath.Join("products", "edit_product.html"))

	// Cargar plantillas de pedidos
	templates["order-list"] = parseTemplate("list.html", filepath.Join("orders", "list.html"))
	templates["order-details"] = parseTemplate("details.html", filepath.Join("orders", "details.html"))
	templates["order-create-form"] = parseTemplate("create_form.html", filepath.Join("orders", "create_form.html"))
	templates["user-orders"] = parseTemplate("user_orders.html", filepath.Join("orders", "user_orders.html"))

	// Cargar plantillas de Autenticación (en su subcarpeta 'auth')
	// El nombre del template cargado es "login" y "register"
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
		Secure:   false,                // ¡IMPORTANT! SET TO 'true' IN PRODUCTION IF USING HTTPS
		SameSite: http.SameSiteLaxMode, // Protege contra ataques CSRF básicos
	}
	log.Println("Tienda de sesiones configurada.")
}

// GetSessionStore devuelve la instancia del CookieStore.
// Necesario para que otros paquetes (como middleware) puedan acceder al store de sesiones.
func GetSessionStore() *sessions.CookieStore {
	return store
}

type CreateClientFormData struct {
	Title          string
	ErrorMessage   string
	SuccessMessage string        // Nuevo campo para el mensaje de éxito
	Client         models.Client // Para rellenar el formulario en caso de error de validación o para edición
}

func renderTemplate(w http.ResponseWriter, tmplName string, data interface{}) {
	tmpl, ok := templates[tmplName]
	if !ok {
		log.Printf("Error: La plantilla '%s' no está cargada en el mapa 'templates'.", tmplName)
		http.Error(w, "Error interno del servidor: plantilla no encontrada.", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("Error al renderizar la plantilla '%s': %v", tmplName, err)
		http.Error(w, "Error interno del servidor al renderizar la página.", http.StatusInternalServerError)
	}
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
	data := CreateClientFormData{ // Usar la nueva estructura
		Title:          "Crear Nuevo Cliente",
		ErrorMessage:   "",
		SuccessMessage: "",
		Client:         models.Client{}, // Cliente vacío para un formulario nuevo
	}
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

type EditClientForm struct {
	Title        string
	ErrorMessage string
	Client       models.Client // Para rellenar el formulario con datos existentes o en caso de error
}

// EditClientPageHandler muestra el formulario para editar un cliente existente.
func (h *ClientHandlers) EditClientPageHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clientID := vars["id"]

	if clientID == "" {
		http.Error(w, "ID de cliente no proporcionado", http.StatusBadRequest)
		return
	}

	client, err := h.clientService.GetClientByID(clientID)
	if err != nil {
		log.Printf("Error al obtener cliente para edición %s: %v", clientID, err)
		http.Error(w, "Error al cargar los datos del cliente: "+err.Error(), http.StatusInternalServerError)
		return
	}

	data := CreateClientFormData{
		Title:          "Editar Cliente",
		ErrorMessage:   "",
		SuccessMessage: "",
		Client:         *client,
	}
	h.renderClientForm(w, "client-edit-form", data.Title, data.ErrorMessage, data.SuccessMessage, data.Client)
}

// UpdateClientHandler procesa el envío del formulario para actualizar un cliente.
func (h *ClientHandlers) UpdateClientHandler(w http.ResponseWriter, r *http.Request) {
	var err error

	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	vars := mux.Vars(r)
	clientID := vars["id"]

	if clientID == "" {
		http.Error(w, "ID de cliente no proporcionado para actualizar", http.StatusBadRequest)
		return
	}

	if err = r.ParseForm(); err != nil {
		log.Printf("Error al parsear el formulario de actualización de cliente: %v", err)
		client, _ := h.clientService.GetClientByID(clientID)
		if client == nil {
			client = &models.Client{ID: clientID}
		}
		renderTemplate(w, "client-edit-form", EditClientForm{
			Title:        "Editar Cliente",
			ErrorMessage: "Error interno del servidor al procesar el formulario.",
			Client:       *client,
		})
		return
	}

	updatedClient := &models.Client{
		ID:      clientID,
		Name:    r.FormValue("name"),
		Email:   r.FormValue("email"),
		Address: r.FormValue("address"),
	}

	// Validación básica de campos
	if updatedClient.Name == "" || updatedClient.Email == "" || updatedClient.Address == "" {
		log.Println("Error de validación: Todos los campos obligatorios deben ser rellenados.")
		renderTemplate(w, "client-edit-form", EditClientForm{
			Title:        "Editar Cliente",
			ErrorMessage: "Por favor, completa todos los campos obligatorios.",
			Client:       *updatedClient,
		})
		return
	}

	_, err = h.clientService.UpdateClient(updatedClient.ID, updatedClient.Name, updatedClient.Email, updatedClient.Address)
	if err != nil {
		log.Printf("Error al actualizar el cliente %s en el servicio: %v", clientID, err)
		renderTemplate(w, "client-edit-form", EditClientForm{
			Title:        "Editar Cliente",
			ErrorMessage: fmt.Sprintf("Error al actualizar el cliente: %v", err),
			Client:       *updatedClient,
		})
		return
	}

	log.Printf("Cliente %s actualizado con éxito.", clientID)
	http.Redirect(w, r, "/clients?message=cliente-actualizado", http.StatusSeeOther)
}
func (h *ClientHandlers) DeleteClientHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	vars := mux.Vars(r)
	clientID := vars["id"]

	if clientID == "" {
		http.Error(w, "ID de cliente no proporcionado para eliminar", http.StatusBadRequest)
		return
	}

	err := h.clientService.DeleteClient(clientID)
	if err != nil {
		log.Printf("Error al eliminar el cliente %s en el servicio: %v", clientID, err)
		// Puedes redirigir con un mensaje de error o mostrar una página de error
		http.Redirect(w, r, fmt.Sprintf("/clients?error=%s", err.Error()), http.StatusSeeOther)
		return
	}

	log.Printf("Cliente %s eliminado con éxito.", clientID)
	http.Redirect(w, r, "/clients?message=cliente-eliminado", http.StatusSeeOther)
}
func (h *ClientHandlers) renderClientForm(w http.ResponseWriter, tmplName, title, errMsg, successMsg string, clientData models.Client) {
	tmpl, ok := templates[tmplName]
	if !ok {
		log.Printf("Error: Plantilla '%s' no encontrada para renderizar.", tmplName)
		http.Error(w, "Error interno del servidor: plantilla no encontrada.", http.StatusInternalServerError)
		return
	}
	data := CreateClientFormData{
		Title:          title,
		ErrorMessage:   errMsg,
		SuccessMessage: successMsg,
		Client:         clientData,
	}
	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("Error al renderizar formulario de cliente: %v", err)
		http.Error(w, "Error interno del servidor.", http.StatusInternalServerError)
	}
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
	data := struct {
		Title        string
		ErrorMessage string
		Product      models.Product
	}{
		Title:   "Crear Nuevo Producto",
		Product: models.Product{}, //
	}
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
	imageURL := r.FormValue("imageURL")

	price, err := strconv.ParseFloat(priceStr, 64)
	if err != nil {
		// Pasa el error a la plantilla para que se muestre en el formulario
		h.renderCreateProductFormWithError(w, "Invalid price. Please enter a valid number.", models.Product{
			Name:        name,
			Description: description,
			Stock:       0,
			ImageURL:    imageURL,
		})
		return
	}

	stock, err := strconv.Atoi(stockStr)
	if err != nil {
		// Pasa el error a la plantilla para que se muestre en el formulario
		h.renderCreateProductFormWithError(w, "Invalid stock. Please enter a valid integer.", models.Product{
			Name:        name,
			Description: description,
			Price:       price,
			ImageURL:    imageURL,
		})
		return
	}

	_, err = h.productService.CreateProduct(name, description, price, stock, imageURL)
	if err != nil {
		log.Printf("Error creating product: %v", err)
		http.Error(w, "Error al crear producto: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/products", http.StatusSeeOther)
}

type EditProductForm struct {
	Title        string
	ErrorMessage string
	Product      models.Product
}

// EditProductPageHandler muestra el formulario para editar un producto existente.
func (h *ProductHandlers) EditProductPageHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	productID := vars["id"]

	if productID == "" {
		http.Error(w, "ID de producto no proporcionado", http.StatusBadRequest)
		return
	}

	product, err := h.productService.GetProductByID(productID)
	if err != nil {
		log.Printf("Error al obtener producto para edición %s: %v", productID, err)
		http.Error(w, "Error al cargar los datos del producto: "+err.Error(), http.StatusInternalServerError)
		return
	}

	data := EditProductForm{
		Title:        "Editar Producto",
		ErrorMessage: "",
		Product:      *product,
	}
	renderTemplate(w, "product-edit-form", data)
}

// UpdateProductHandler procesa el envío del formulario para actualizar un producto.
func (h *ProductHandlers) UpdateProductHandler(w http.ResponseWriter, r *http.Request) {
	var err error

	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	vars := mux.Vars(r)
	productID := vars["id"]

	if productID == "" {
		http.Error(w, "ID de producto no proporcionado para actualizar", http.StatusBadRequest)
		return
	}

	if err = r.ParseForm(); err != nil {
		log.Printf("Error al parsear el formulario de actualización de producto: %v", err)
		product, _ := h.productService.GetProductByID(productID)
		if product == nil {
			product = &models.Product{ID: productID}
		}
		renderTemplate(w, "product-edit-form", EditProductForm{
			Title:        "Editar Producto",
			ErrorMessage: "Error interno del servidor al procesar el formulario.",
			Product:      *product,
		})
		return
	}

	name := r.FormValue("name")
	description := r.FormValue("description")
	priceStr := r.FormValue("price")
	stockStr := r.FormValue("stock")
	imageURL := r.FormValue("imageURL")

	price, parseErr := strconv.ParseFloat(priceStr, 64)
	if parseErr != nil {
		log.Printf("Error al parsear precio para actualización: %v", parseErr)
		product, _ := h.productService.GetProductByID(productID)
		if product == nil {
			product = &models.Product{ID: productID, Name: name, Description: description, ImageURL: imageURL}
		}
		product.Price = price // Intentar mantener el valor si es parcial
		renderTemplate(w, "product-edit-form", EditProductForm{
			Title:        "Editar Producto",
			ErrorMessage: "El precio debe ser un número válido.",
			Product:      *product,
		})
		return
	}

	stock, parseErr := strconv.Atoi(stockStr)
	if parseErr != nil {
		log.Printf("Error al parsear stock para actualización: %v", parseErr)
		product, _ := h.productService.GetProductByID(productID)
		if product == nil {
			product = &models.Product{ID: productID, Name: name, Description: description, Price: price, ImageURL: imageURL}
		}
		product.Stock = stock
		renderTemplate(w, "product-edit-form", EditProductForm{
			Title:        "Editar Producto",
			ErrorMessage: "El stock debe ser un número entero válido.",
			Product:      *product,
		})
		return
	}

	// Llamar al servicio para actualizar el producto
	_, err = h.productService.UpdateProduct(productID, name, description, price, stock, imageURL)
	if err != nil {
		log.Printf("Error al actualizar el producto %s en el servicio: %v", productID, err)
		renderTemplate(w, "product-edit-form", EditProductForm{
			Title:        "Editar Producto",
			ErrorMessage: fmt.Sprintf("Error al actualizar el producto: %v", err),
			Product:      models.Product{ID: productID, Name: name, Description: description, Price: price, Stock: stock, ImageURL: imageURL}, // Pasa los datos ingresados
		})
		return
	}

	log.Printf("Producto %s actualizado con éxito.", productID)
	http.Redirect(w, r, "/products?message=producto-actualizado", http.StatusSeeOther)
}

// DeleteProductHandler procesa la solicitud para eliminar un producto.
func (h *ProductHandlers) DeleteProductHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	vars := mux.Vars(r)
	productID := vars["id"]

	if productID == "" {
		http.Error(w, "ID de producto no proporcionado para eliminar", http.StatusBadRequest)
		return
	}

	err := h.productService.DeleteProduct(productID)
	if err != nil {
		log.Printf("Error al eliminar el producto %s en el servicio: %v", productID, err)
		http.Redirect(w, r, fmt.Sprintf("/products?error=%s", err.Error()), http.StatusSeeOther)
		return
	}

	log.Printf("Producto %s eliminado con éxito.", productID)
	http.Redirect(w, r, "/products?message=producto-eliminado", http.StatusSeeOther)
}

// renderCreateProductFormWithError es una función auxiliar para renderizar el formulario de creación de producto con un mensaje de error.
func (h *ProductHandlers) renderCreateProductFormWithError(w http.ResponseWriter, msg string, productData models.Product) {
	tmpl, ok := templates["product-create-form"]
	if !ok {
		log.Printf("Error: 'product-create-form' template not found for error rendering.")
		http.Error(w, "Internal server error: template not found.", http.StatusInternalServerError)
		return
	}
	data := struct {
		Title        string
		ErrorMessage string
		Product      models.Product
	}{
		Title:        "Create New Product",
		ErrorMessage: msg,
		Product:      productData,
	}
	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("Error rendering product creation form with error: %v", err)
		http.Error(w, "Internal server error.", http.StatusInternalServerError)
	}
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

// IndexHandlers estructura para manejar las rutas de la página principal (e-commerce).
type IndexHandlers struct {
	productService *services.ProductService
}

// NewIndexHandlers crea una nueva instancia de IndexHandlers.
func NewIndexHandlers(ps *services.ProductService) *IndexHandlers {
	return &IndexHandlers{productService: ps}
}

// IndexPageHandler renderiza la página principal con la lista de productos
func (h *IndexHandlers) IndexPageHandler(w http.ResponseWriter, r *http.Request) {
	// Obtener todos los productos desde el servicio
	products, err := h.productService.GetAllProducts()
	if err != nil {
		log.Printf("Error al obtener productos para la página index: %v", err)
		http.Error(w, "Error interno del servidor al cargar productos.", http.StatusInternalServerError)
		return
	}

	// Recuperar información de la sesión (ej. si el usuario está logueado)
	session, err := store.Get(r, "session-name")
	if err != nil {
		log.Printf("Error al obtener sesión en IndexPageHandler: %v", err)

	}

	isAuthenticated := session.Values["authenticated"] == true
	username := ""
	userRole := "" // Obtener el rol del usuario de la sesión
	if isAuthenticated {
		if u, ok := session.Values["username"].(string); ok {
			username = u
		}
		if r, ok := session.Values["userRole"].(string); ok { // Obtener el rol
			userRole = r
		}
	}

	// Redirigir a /home si el usuario es un administrador y está en la página de inicio (/)

	if isAuthenticated && userRole == "admin" && r.URL.Path == "/" {
		http.Redirect(w, r, "/home", http.StatusSeeOther)
		return
	}

	// Datos a pasar a la plantilla
	data := struct {
		Title           string
		Products        []models.Product
		IsAuthenticated bool
		Username        string
		UserRole        string // Pasar el rol a la plantilla

	}{
		Title:           "Productos de la Tienda Hípica",
		Products:        products,
		IsAuthenticated: isAuthenticated,
		Username:        username,
		UserRole:        userRole, // Asignar el rol
	}

	// Renderizar la plantilla index.html
	tmpl, ok := templates["index"]
	if !ok {
		log.Printf("Error: Plantilla 'index' no encontrada en el mapa de plantillas.")
		http.Error(w, "Error interno del servidor: plantilla no encontrada.", http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("Error al renderizar la plantilla index: %v", err)
		http.Error(w, "Error interno del servidor al renderizar la página.", http.StatusInternalServerError)
	}
}

// HomeHandler (Para administradores)
func (h *IndexHandlers) AdminHomeHandler(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "session-name")
	if err != nil {
		log.Printf("Error al obtener sesión en AdminHomeHandler: %v", err)
		http.Redirect(w, r, "/login", http.StatusSeeOther) // Si no hay sesión, al login
		return
	}

	isAuthenticated := session.Values["authenticated"] == true
	userRole := ""
	if r, ok := session.Values["userRole"].(string); ok {
		userRole = r
	}

	// Si no está autenticado o no es admin, redirigir
	if !isAuthenticated || userRole != "admin" {
		http.Redirect(w, r, "/login", http.StatusSeeOther) // O a una página de "Acceso Denegado"
		return
	}

	// Datos a pasar a la plantilla home.html (puedes personalizar esto)
	data := struct {
		Title           string
		Username        string
		IsAuthenticated bool
		UserRole        string
	}{
		Title:           "Panel de Administración",
		Username:        session.Values["username"].(string),
		IsAuthenticated: isAuthenticated,
		UserRole:        userRole,
	}

	tmpl, ok := templates["home"]
	if !ok {
		log.Printf("Error: Plantilla 'home' no encontrada en el mapa de plantillas.")
		http.Error(w, "Error interno del servidor: plantilla de administración no encontrada.", http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("Error al renderizar la plantilla home para admin: %v", err)
		http.Error(w, "Error interno del servidor al renderizar la página de administración.", http.StatusInternalServerError)
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
func (h *OrderHandlers) UserOrdersHandler(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "session-name")
	if err != nil {
		log.Printf("UserOrdersHandler: Error al obtener sesión: %v", err)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Recuperar el username de la sesión (asumiendo que es el email para los usuarios normales)
	userEmail, ok := session.Values["username"].(string)
	if !ok || userEmail == "" {
		log.Println("UserOrdersHandler: Email de usuario no encontrado en la sesión o es inválido.")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Buscar el cliente asociado a este email
	client, err := h.clientService.GetClientByEmail(userEmail)
	if err != nil {
		log.Printf("UserOrdersHandler: No se encontró cliente para el email '%s': %v", userEmail, err)

		tmpl, ok := templates["user-orders"]
		if !ok {
			http.Error(w, "Plantilla 'user-orders' no encontrada", http.StatusInternalServerError)
			return
		}
		data := struct {
			Title   string
			Orders  []models.Order
			Message string
		}{
			Title:   "Mis Pedidos",
			Orders:  []models.Order{},
			Message: "No se encontró un perfil de cliente asociado a tu cuenta. Asegúrate de que tu email esté registrado como cliente.",
		}
		if err := tmpl.Execute(w, data); err != nil {
			log.Printf("Error al renderizar plantilla de pedidos de usuario (cliente no encontrado): %v", err)
			http.Error(w, "Error interno del servidor al renderizar la página.", http.StatusInternalServerError)
		}
		return
	}
	// Si se encuentra el cliente, obtener sus pedidos por el ClientID
	orders, err := h.orderService.GetOrdersByClientID(client.ID)
	if err != nil {
		log.Printf("Error al obtener pedidos para el cliente %s (email: %s): %v", client.ID, userEmail, err)
		http.Error(w, "Error al cargar tus pedidos.", http.StatusInternalServerError)
		return
	}

	tmpl, ok := templates["user-orders"]
	if !ok {
		log.Printf("Error: Plantilla 'user-orders' no encontrada en el mapa de plantillas.")
		http.Error(w, "Error interno del servidor: plantilla de pedidos de usuario no encontrada.", http.StatusInternalServerError)
		return
	}

	data := struct {
		Title   string
		Orders  []models.Order
		Message string
	}{
		Title:   "Mis Pedidos",
		Orders:  orders,
		Message: "",
	}

	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("Error al renderizar plantilla de pedidos de usuario: %v", err)
		http.Error(w, "Error interno del servidor al renderizar la página.", http.StatusInternalServerError)
	}
}

// AuthHandlers maneja las peticiones relacionadas con la autenticación de usuarios.
type AuthHandlers struct {
	authService   *services.AuthService
	clientService *services.ClientService
}

// NewAuthHandlers crea e inicializa AuthHandlers.
// Recibe también ClientService.
func NewAuthHandlers(authService *services.AuthService, clientService *services.ClientService) *AuthHandlers {
	return &AuthHandlers{
		authService:   authService,
		clientService: clientService, // Asignar ClientService
	}
}

// LoginPageHandler muestra el formulario de login.
func (h *AuthHandlers) LoginPageHandler(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "session-name")
	if err != nil {
		log.Printf("Error al obtener sesión en LoginPageHandler: %v", err)

	}
	// Si el usuario ya está autenticado, redirigir según su rol
	if session.Values["authenticated"] == true {
		if userRole, ok := session.Values["userRole"].(string); ok {
			if userRole == "admin" {
				http.Redirect(w, r, "/home", http.StatusSeeOther)
			} else {
				http.Redirect(w, r, "/index", http.StatusSeeOther)
			}
		} else {
			// Si el rol no está en sesión por alguna razón, redirigir a un default
			http.Redirect(w, r, "/index", http.StatusSeeOther)
		}
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
	// Aquí se comprueba si el template fue renderizado con un error
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
	session.Values["username"] = user.Email
	session.Values["userRole"] = user.Role // Guardando el rol del usuario

	// Guardar la sesión para que la cookie sea enviada al cliente
	err = session.Save(r, w)
	if err != nil {
		log.Printf("Error al guardar sesión: %v", err)
		http.Error(w, "Error de sesión", http.StatusInternalServerError)
		return
	}

	log.Printf("Usuario '%s' (Rol: %s) ha iniciado sesión exitosamente.", user.Username, user.Role)

	// Redirección basada en el rol del usuario
	if user.Role == "admin" {
		http.Redirect(w, r, "/home", http.StatusSeeOther) // Redirige a /home para administradores
	} else {
		http.Redirect(w, r, "/index", http.StatusSeeOther) // Redirige a /index para usuarios normales
	}
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
	// Si el usuario ya está autenticado, redirigir según su rol
	if session.Values["authenticated"] == true {
		if userRole, ok := session.Values["userRole"].(string); ok {
			if userRole == "admin" {
				http.Redirect(w, r, "/home", http.StatusSeeOther)
			} else {
				http.Redirect(w, r, "/index", http.StatusSeeOther)
			}
		} else {
			// Si el rol no está en sesión, default a index
			http.Redirect(w, r, "/index", http.StatusSeeOther)
		}
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
	email := r.FormValue("email")

	if username == "" || password == "" || email == "" {
		h.renderRegisterFormWithError(w, "Nombre de usuario, email y contraseña son obligatorios.")
		return
	}

	user, err := h.authService.RegisterUser(username, password, email, "user")
	if err != nil {
		log.Printf("Error al registrar usuario '%s' (Email: %s): %v", username, email, err)
		h.renderRegisterFormWithError(w, fmt.Sprintf("Error al registrar: %v", err.Error()))
		return
	}

	// --- Lógica para buscar o crear un cliente asociado ---

	client, err := h.clientService.GetClientByEmail(email)
	if err != nil {
		log.Printf("Cliente con email '%s' no encontrado. Creando nuevo cliente asociado.", email)
		client, err = h.clientService.CreateClient(username, email, "") // Nombre del cliente puede ser el username, dirección vacía
		if err != nil {
			log.Printf("Error crítico al crear cliente asociado para usuario '%s' (Email: %s): %v", username, email, err)
		} else {
			log.Printf("Cliente '%s' (ID: %s, Email: %s) creado exitosamente y asociado.", client.Name, client.ID, client.Email)
		}
	} else {
		log.Printf("Cliente existente '%s' (ID: %s, Email: %s) asociado con nuevo usuario.", client.Name, client.ID, client.Email)
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
	session.Values["username"] = user.Email
	session.Values["userRole"] = user.Role

	if client != nil {
		session.Values["clientID"] = client.ID
	}

	err = session.Save(r, w)
	if err != nil {
		log.Printf("Error al guardar sesión después de registro: %v", err)
		http.Error(w, "Error de sesión", http.StatusInternalServerError)
		return
	}

	log.Printf("Nuevo usuario '%s' registrado y logueado automáticamente.", user.Username)
	http.Redirect(w, r, "/index", http.StatusSeeOther)
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
	session.Values["clientID"] = nil // También borrar el ID del cliente
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
