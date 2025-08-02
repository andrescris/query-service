package queryservice

import (
	"errors"
	"net/http"

	"cloud.google.com/go/firestore"
	"github.com/andrescris/firestore/lib/firebase" // Aún la usamos para el tipo QueryOptions
	"github.com/gin-gonic/gin"
	"google.golang.org/api/iterator"
)

// QueryHandler ejecuta una consulta utilizando el cliente y las opciones del contexto.
func QueryHandler(c *gin.Context) {
	collection := c.Param("collection")

	// --- INICIO DE LA CORRECCIÓN ---

	// 1. OBTENER EL CLIENTE DEL CONTEXTO
	// En lugar de confiar en una variable global, usamos el cliente
	// que el servicio 'products' nos inyectó.
	clientValue, exists := c.Get("firestoreClient")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "CRITICAL: Firestore client not found in context."})
		return
	}
	client, ok := clientValue.(*firestore.Client)
	if !ok || client == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "CRITICAL: Invalid or nil Firestore client in context."})
		return
	}

	// 2. OBTENER LAS OPCIONES DE CONSULTA DEL CONTEXTO
	optionsValue, exists := c.Get("secure_query_options")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Secure query options not found in context."})
		return
	}
	options := optionsValue.(firebase.QueryOptions)

	// 3. EJECUTAR LA CONSULTA USANDO EL CLIENTE INYECTADO
	// Ya no llamamos a la función problemática `firestore.QueryDocuments`.
	// Construimos y ejecutamos la consulta aquí mismo.
	query := client.Collection(collection).Query
	for _, filter := range options.Filters {
		query = query.Where(filter.Field, filter.Operator, filter.Value)
	}

	iter := query.Documents(c.Request.Context())
	defer iter.Stop()

	var results []map[string]interface{}
	for {
		doc, err := iter.Next()
		if errors.Is(err, iterator.Done) {
			break // Se acabaron los resultados
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to iterate documents", "details": err.Error()})
			return
		}
		results = append(results, doc.Data())
	}

	// --- FIN DE LA CORRECCIÓN ---

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"documents":  results,
		"count":      len(results),
		"collection": collection,
		"query":      options,
	})
}