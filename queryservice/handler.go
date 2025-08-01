package queryservice

import (
	"context"
	"errors"
	"net/http"

	"cloud.google.com/go/firestore"
	"github.com/andrescris/firestore/lib/firebase" // Aún la usamos para el tipo QueryOptions
	"github.com/gin-gonic/gin"
	"google.golang.org/api/iterator"
)

// QueryHandler ejecuta una consulta utilizando las opciones seguras y el cliente
// que el middleware y el inyector prepararon.
func QueryHandler(c *gin.Context) {
	collection := c.Param("collection")

	// 1. Obtiene el cliente de Firestore desde el contexto (inyectado en main.go).
	clientValue, exists := c.Get("firestoreClient")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Firestore client not found in context."})
		return
	}
	client, ok := clientValue.(*firestore.Client)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid Firestore client type in context."})
		return
	}

	// 2. Obtiene las opciones de consulta seguras del contexto.
	optionsValue, exists := c.Get("secure_query_options")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Secure query options not found in context."})
		return
	}
	options := optionsValue.(firebase.QueryOptions)

	// 3. Ejecuta la consulta usando el cliente y las opciones.
	docs, err := executeQuery(c.Request.Context(), client, collection, options)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to execute query",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"documents":  docs,
		"count":      len(docs),
		"collection": collection,
		"query":      options,
	})
}

// executeQuery construye y ejecuta la consulta de Firestore usando el cliente proporcionado.
func executeQuery(ctx context.Context, client *firestore.Client, collection string, options firebase.QueryOptions) ([]map[string]interface{}, error) {
	query := client.Collection(collection).Query
	for _, filter := range options.Filters {
		query = query.Where(filter.Field, filter.Operator, filter.Value)
	}

	// Aquí podrías añadir más lógica para otros campos de QueryOptions si los tuvieras
	// Por ejemplo:
	// if options.Limit > 0 {
	//     query = query.Limit(options.Limit)
	// }
	// if options.OrderBy != "" {
	//     query = query.OrderBy(options.OrderBy, firestore.Asc)
	// }

	iter := query.Documents(ctx)
	defer iter.Stop()

	var results []map[string]interface{}
	for {
		doc, err := iter.Next()
		if errors.Is(err, iterator.Done) {
			break // Se terminaron los documentos, salimos del bucle.
		}
		if err != nil {
			return nil, err // Ocurrió un error real.
		}
		results = append(results, doc.Data())
	}
	return results, nil
}