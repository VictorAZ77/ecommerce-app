E-commerce - Go

Este proyecto implementa el backend de un sistema de gestión integral para un negocio de e-commerce, desarrollado en Go. Su objetivo principal es proporcionar una API robusta y segura para la administración de usuarios, clientes, productos y pedidos, facilitando las operaciones comerciales y el soporte a una posible interfaz de usuario (frontend).

La arquitectura se enfoca en la separación de responsabilidades, utilizando capas de repositorio para la persistencia de datos (SQL Server), servicios para la lógica de negocio y handlers web para la exposición de la API y la gestión de las interacciones HTTP.

Principales Funcionalidades del Código
El sistema está estructurado en módulos que cubren las áreas clave de un e-commerce:

1. Gestión de Usuarios y Autenticación
Modelos (internal/models/user.go): Define la estructura de datos para los usuarios, incluyendo campos como ID, Email, Username, Password (hash), Role, y timestamps.

Repositorio (internal/repositories/user_repo.go): Provee métodos para interactuar con la base de datos SQL Server para operaciones CRUD (Crear, Leer, Actualizar, Borrar) de usuarios, incluyendo búsqueda por ID, username y email.

Servicio de Autenticación (internal/services/auth_service.go):

Manejo del registro de nuevos usuarios, incluyendo el hashing seguro de contraseñas (bcrypt) y la validación de unicidad de username y email.

Autenticación de usuarios y generación de sesiones.

Integración con un servicio de notificaciones para enviar correos de bienvenida tras el registro.

Handlers Web (web/handlers.go): Puntos de entrada para el registro de usuarios, inicio y cierre de sesión, y la gestión de sesiones.

2. Gestión de Clientes
Modelos (internal/models/client.go): Define la estructura de datos para los clientes.

Repositorio (internal/repositories/client_repo.go): Provee métodos CRUD para la persistencia de datos de clientes en SQL Server, incluyendo búsqueda por ID y email.

Servicio de Clientes (internal/services/client_service.go):

Lógica de negocio para la creación, actualización, eliminación y recuperación de clientes.

Validación de datos de cliente y asegura la unicidad del email.

Handlers Web (web/handlers.go): Puntos de entrada para la administración de clientes (ej. crear, ver, actualizar clientes).

3. Gestión de Productos e Inventario
Modelos (internal/models/product.go): Define la estructura de datos para los productos, incluyendo nombre, descripción, precio, stock y URL de imagen.

Repositorio (No proporcionado directamente, pero implícito): Se esperaría un product_repo.go para la persistencia de productos.

Servicio de Inventario de Productos (internal/services/product_inventory_service.go - asumiendo su existencia): Lógica para gestionar el stock, precios y detalles de los productos.

Handlers Web (web/handlers.go): Puntos de entrada para la administración de productos.

4. Gestión de Pedidos y Detalles de Pedido
Modelos (internal/models/order.go): Define las estructuras para Order (pedido principal) y OrderItem (ítems individuales dentro de un pedido).

Repositorio (No proporcionado directamente, pero implícito): Se esperaría un order_repo.go para la persistencia de pedidos y sus ítems.

Servicio de Administración de Pedidos (internal/services/order_admin_service.go - asumiendo su existencia): Lógica para la creación, actualización de estado, y gestión de los pedidos y sus componentes.

Handlers Web (web/handlers.go): Puntos de entrada para la administración y consulta de pedidos.

5. Servicio de Notificaciones
Servicio de Notificaciones (internal/services/notification_service.go - asumiendo su existencia): Abstrae la lógica de envío de notificaciones

Contiene las implementaciones concretas para enviar notificaciones .

6. Arquitectura General
El proyecto sigue una arquitectura en capas:

web/: Capa de presentación (handlers HTTP).

internal/services/: Capa de lógica de negocio.

internal/repositories/: Capa de persistencia de datos.

internal/models/: Definiciones de estructuras de datos (entidades).

Sistema de e-commerce desarrollado en Go con arquitectura limpia y SQL Server.
🚀 Características

✅ Autenticación de usuarios
👥 Gestión de clientes
📦 Catálogo de productos (vista admin) (vista Usuarios en Desarrollo)
🛒 Sistema de órdenes
🏗️ Arquitectura limpia 
🗄️ SQL Server como base de datos

📋 Requisitos

Go 1.19+
SQL Server
Variables de entorno configuradas

⚡ Instalación Rápida
bash# Clonar repositorio
git clone <tu-repo>
cd Backend

# Instalar dependencias
go mod tidy

# Configurar variables de entorno
cp .env.example .env

# Ejecutar aplicación
go run cmd/api/main.go
🔧 Configuración
Crear archivo .env con:
envDB_HOST=localhost
DB_PORT=1433
DB_USER=tu_usuario
DB_PASSWORD=tu_password
DB_NAME=ecommerce_db
