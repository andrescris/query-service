package queryservice

import (
	"net/http"

	"github.com/andrescris/firestore/lib/firebase"
	"github.com/andrescris/firestore/lib/firebase/firestore" // Importar firestore
	"github.com/gin-gonic/gin"
)

// QueryHandler ejecuta una consulta utilizando las opciones seguras
// que el middleware preparó.
func QueryHandler(c *gin.Context) {
	collection := c.Param("collection")
	
	// Obtenemos las opciones seguras del contexto
	optionsValue, exists := c.Get("secure_query_options")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "No se pudieron procesar las opciones de la consulta.",
		})
		return
	}
	
	options := optionsValue.(firebase.QueryOptions)

	// Ejecutar la consulta
	docs, err := firestore.QueryDocuments(c.Request.Context(), collection, options)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Falló la ejecución de la consulta en la base de datos",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"documents":  docs,
		"count":      len(docs),
		"collection": collection,
		"query":      options, // Devolvemos la consulta final (con el filtro de seguridad)
	})
}