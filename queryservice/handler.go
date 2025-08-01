package queryservice

import (
	"errors"
	"log"
	"net/http"

	"cloud.google.com/go/firestore"
	"github.com/andrescris/firestore/lib/firebase" // Aún la usamos para el tipo QueryOptions
	"github.com/gin-gonic/gin"
	"google.golang.org/api/iterator"
)

// QueryHandler ejecuta una consulta utilizando el cliente y las opciones del contexto.
func QueryHandler(c *gin.Context) {
	collection := c.Param("collection")
	log.Printf("🔍 DEBUG QueryHandler: Starting query for collection: %s", collection)

	// --- INICIO DE LA CORRECCIÓN CON DEBUG ---

	// 1. OBTENER EL CLIENTE DEL CONTEXTO
	log.Println("🔍 DEBUG QueryHandler: Attempting to get firestoreClient from context...")
	clientValue, exists := c.Get("firestoreClient")
	log.Printf("🔍 DEBUG QueryHandler: Client from context - exists: %t, value: %v", exists, clientValue)
	
	if !exists {
		log.Println("❌ DEBUG QueryHandler: firestoreClient not found in context")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "CRITICAL: Firestore client not found in context."})
		return
	}
	
	log.Printf("🔍 DEBUG QueryHandler: Attempting to cast clientValue to *firestore.Client...")
	client, ok := clientValue.(*firestore.Client)
	log.Printf("🔍 DEBUG QueryHandler: Cast result - ok: %t, client: %v, client != nil: %t", ok, client, client != nil)
	
	if !ok {
		log.Println("❌ DEBUG QueryHandler: Failed to cast to *firestore.Client")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "CRITICAL: Failed to cast Firestore client from context."})
		return
	}
	
	if client == nil {
		log.Println("❌ DEBUG QueryHandler: Client is nil after casting")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "CRITICAL: Firestore client is nil in context."})
		return
	}

	// 2. OBTENER LAS OPCIONES DE CONSULTA DEL CONTEXTO
	log.Println("🔍 DEBUG QueryHandler: Getting query options from context...")
	optionsValue, exists := c.Get("secure_query_options")
	if !exists {
		log.Println("❌ DEBUG QueryHandler: secure_query_options not found in context")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Secure query options not found in context."})
		return
	}
	options := optionsValue.(firebase.QueryOptions)
	log.Printf("🔍 DEBUG QueryHandler: Query options: %+v", options)

	// 3. EJECUTAR LA CONSULTA USANDO EL CLIENTE INYECTADO
	log.Println("🔍 DEBUG QueryHandler: About to create query...")
	log.Printf("🔍 DEBUG QueryHandler: Client before Collection call: %v", client)
	
	// Esta es la línea que está fallando - añadimos más debug
	log.Printf("🔍 DEBUG QueryHandler: Calling client.Collection(%s)...", collection)
	collectionRef := client.Collection(collection)
	log.Printf("🔍 DEBUG QueryHandler: Collection ref: %v", collectionRef)
	
	log.Println("🔍 DEBUG QueryHandler: Getting Query from collection ref...")
	query := collectionRef.Query
	log.Printf("🔍 DEBUG QueryHandler: Query: %v", query)
	
	// Aplicar filtros
	for i, filter := range options.Filters {
		log.Printf("🔍 DEBUG QueryHandler: Applying filter %d: %+v", i, filter)
		query = query.Where(filter.Field, filter.Operator, filter.Value)
	}

	log.Println("🔍 DEBUG QueryHandler: Creating documents iterator...")
	iter := query.Documents(c.Request.Context())
	defer iter.Stop()

	var results []map[string]interface{}
	log.Println("🔍 DEBUG QueryHandler: Starting iteration...")
	for {
		doc, err := iter.Next()
		if errors.Is(err, iterator.Done) {
			log.Println("🔍 DEBUG QueryHandler: Iterator done")
			break // Se acabaron los resultados
		}
		if err != nil {
			log.Printf("❌ DEBUG QueryHandler: Iterator error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to iterate documents", "details": err.Error()})
			return
		}
		log.Printf("🔍 DEBUG QueryHandler: Got document: %s", doc.Ref.ID)
		results = append(results, doc.Data())
	}

	log.Printf("🔍 DEBUG QueryHandler: Query completed successfully with %d results", len(results))

	// --- FIN DE LA CORRECCIÓN ---

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"documents":  results,
		"count":      len(results),
		"collection": collection,
		"query":      options,
	})
}