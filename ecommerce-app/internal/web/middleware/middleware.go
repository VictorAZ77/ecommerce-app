package middleware

import (
	"github.com/gorilla/sessions"
	"log"
	"net/http"
)

var store *sessions.CookieStore

// SetSessionStoreMiddleware inicializa la tienda de sesiones para el paquete de middleware.
func SetSessionStoreMiddleware(s *sessions.CookieStore) {
	store = s
	log.Println("Middleware: Tienda de sesiones configurada.")
}

// AuthRequired es un middleware que verifica si el usuario está autenticado.
func AuthRequired(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := store.Get(r, "session-name")
		if err != nil {
			log.Printf("Middleware AuthRequired: Error al obtener sesión: %v", err)
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		if session.Values["authenticated"] != true {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Si está autenticado, pasa la petición al siguiente manejador
		next.ServeHTTP(w, r)
	})
}

// RoleRequired es un middleware para verificar si el usuario tiene un rol específico.
func RoleRequired(requiredRole string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session, err := store.Get(r, "session-name")
			if err != nil || session.Values["authenticated"] != true {
				log.Printf("Middleware RoleRequired: Error al obtener sesión o no autenticado: %v", err)
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}

			userRole, ok := session.Values["userRole"].(string)
			if !ok || userRole != requiredRole {
				log.Printf("Middleware RoleRequired: Acceso denegado para usuario con rol '%s'. Se requiere rol '%s'.", userRole, requiredRole)
				http.Error(w, "Acceso denegado: rol insuficiente", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
