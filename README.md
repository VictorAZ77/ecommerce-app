E-commerce - Go
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
