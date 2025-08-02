package queryservice

import (
	"log"
	"net/http"

	"github.com/andrescris/firestore/lib/firebase" // Asumiendo que aquí está la definición de QueryOptions
	"github.com/gin-gonic/gin"
)

// ConditionalSubdomainFilterMiddleware se encarga de modificar y validar la consulta
// antes de que llegue al handler final.
func ConditionalSubdomainFilterMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Println("--- DEBUG: Entrando a ConditionalSubdomainFilterMiddleware ---")

		var options firebase.QueryOptions
		if err := c.ShouldBindJSON(&options); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Formato de consulta JSON inválido", "details": err.Error()})
			return
		}

		// Revisar si la clave tiene permisos de administrador para saltar el filtro
		bypassFilter := false
		if perms, exists := c.Get("permissions"); exists {
			permissions, _ := perms.([]interface{})
			log.Printf("DEBUG: Permisos encontrados en el contexto: %v", permissions)
			for _, p := range permissions {
				if p.(string) == "read:all_subdomains" {
					bypassFilter = true
					log.Println("DEBUG: ¡Permiso de admin encontrado! Se omitirá el filtro de subdominio.")
					break
				}
			}
		} else {
			log.Println("DEBUG: No se encontraron permisos en el contexto. Se aplicará el filtro por defecto.")
		}

		log.Printf("DEBUG: Valor de bypassFilter: %t", bypassFilter)

		// Aplicar el filtro de subdominio si NO se debe saltar
		if !bypassFilter {
			userSubdomain := c.GetHeader("X-Client-Subdomain")
			log.Printf("DEBUG: Subdominio leído de la cabecera: '%s'", userSubdomain)

			if userSubdomain == "" {
				log.Println("ERROR: El encabezado X-Client-Subdomain está vacío.")
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "El encabezado X-Client-Subdomain es requerido para esta consulta."})
				return
			}
			
			subdomainFilter := firebase.QueryFilter{
				Field:    "subdomain",
				Operator: "==",
				Value:    userSubdomain,
			}
			options.Filters = append(options.Filters, subdomainFilter)
			log.Printf("DEBUG: Filtro de subdominio AÑADIDO: {Field: subdomain, Operator: ==, Value: %s}", userSubdomain)
		}

		// Validar que la consulta siempre incluya un project_id (sin cambios)
		hasProjectFilter := false
		for _, filter := range options.Filters {
			if filter.Field == "project_id" && filter.Operator == "==" {
				hasProjectFilter = true
				break
			}
		}
		if !hasProjectFilter {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "La consulta debe incluir un filtro por 'project_id' con el operador '=='"})
			return
		}

		c.Set("secure_query_options", options)
		log.Println("--- DEBUG: Saliendo de ConditionalSubdomainFilterMiddleware ---")
		c.Next()
	}
}