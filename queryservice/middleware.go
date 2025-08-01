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

		// 2. Aplicar filtro de seguridad por subdominio
		claimsValue, _ := c.Get("claims")
		claims, _ := claimsValue.(map[string]interface{})
		role, _ := claims["role"].(string)

		// Solo aplicamos el filtro si el usuario NO es admin
		if role != "admin" {
			userSubdomain, exists := c.Get("subdomain")
			if !exists {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
					"error": "No se pudo determinar el subdominio del usuario para filtrar la consulta.",
				})
				return
			}
			
			// Añadimos el filtro de forma segura
			subdomainFilter := firebase.QueryFilter{
				Field:    "subdomain",
				Operator: "==",
				Value:    userSubdomain.(string),
			}
			options.Filters = append(options.Filters, subdomainFilter)
		}

		// 3. Validar que la consulta siempre incluya un project_id
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

		// 4. Guardar las opciones de consulta seguras en el contexto para el handler
		c.Set("secure_query_options", options)

		c.Next()
	}
}