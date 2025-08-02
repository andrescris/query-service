package queryservice

import (
	"net/http"

	"github.com/andrescris/firestore/lib/firebase" // Asumiendo que aquí está la definición de QueryOptions
	"github.com/gin-gonic/gin"
)

// ConditionalSubdomainFilterMiddleware se encarga de modificar y validar la consulta
// antes de que llegue al handler final.
func ConditionalSubdomainFilterMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		var options firebase.QueryOptions

		// 1. Leer y validar el JSON de la consulta
		if err := c.ShouldBindJSON(&options); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error":   "Formato de consulta JSON inválido",
				"details": err.Error(),
			})
			return
		}

		// --- LÓGICA CORREGIDA ---
		// 2. Revisar si la clave tiene permisos de administrador para saltar el filtro.
		//    Definimos un permiso especial, por ejemplo "read:all_subdomains".
		bypassFilter := false
		if perms, exists := c.Get("permissions"); exists {
			permissions, _ := perms.([]interface{})
			for _, p := range permissions {
				if p.(string) == "read:all_subdomains" {
					bypassFilter = true
					break
				}
			}
		}

		// 3. Aplicar el filtro de subdominio si NO se debe saltar.
		if !bypassFilter {
			// Leemos el subdominio de la cabecera X-Client-Subdomain
			userSubdomain := c.GetHeader("X-Client-Subdomain")
			if userSubdomain == "" {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
					"error": "El encabezado X-Client-Subdomain es requerido para esta consulta.",
				})
				return
			}
			
			// Añadimos el filtro de forma segura
			subdomainFilter := firebase.QueryFilter{
				Field:    "subdomain",
				Operator: "==",
				Value:    userSubdomain,
			}
			options.Filters = append(options.Filters, subdomainFilter)
		}

		// 4. Validar que la consulta siempre incluya un project_id (sin cambios)
		hasProjectFilter := false
		for _, filter := range options.Filters {
			if filter.Field == "project_id" && filter.Operator == "==" {
				hasProjectFilter = true
				break
			}
		}
		if !hasProjectFilter {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": "La consulta debe incluir un filtro por 'project_id' con el operador '=='",
			})
			return
		}

		// 5. Guardar las opciones de consulta seguras en el contexto (sin cambios)
		c.Set("secure_query_options", options)

		c.Next()
	}
}