package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"time"

	"product-microservice/internal/b_application/service"
	"product-microservice/internal/c_adapter/in/a_web/command"
	"product-microservice/internal/c_adapter/in/a_web/health"
	"product-microservice/internal/c_adapter/in/a_web/query"
	"product-microservice/internal/c_adapter/in/b_kafka"

	"product-microservice/internal/c_adapter/out/command_db/postgres/outbox"
	"product-microservice/internal/c_adapter/out/command_db/postgres/product"
	"product-microservice/internal/c_adapter/out/kafka"

	"product-microservice/internal/c_adapter/out/query_db/mongorepo"
	"product-microservice/internal/d_config"

	"os"
	"os/signal"
	"syscall"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	_ "github.com/lib/pq" // registra el driver de Postgres (side-effect import)
)

func main() {
	// Load reads .env (if present) and builds connection settings.
	// main never hardcodes hosts/passwords; cfg is the single source for this process.
	cfg := d_config.Load()

	// --- Infrastructure (adapters out) ---
	// Postgres
	postgresDb, err := sql.Open("postgres", cfg.PostgresDSN)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer postgresDb.Close()

	// MongoDB
	mongoClient, err := mongo.Connect(
		context.Background(),
		options.Client().ApplyURI(cfg.MongoURI).SetServerSelectionTimeout(5*time.Second),
	)
	if err != nil {
		log.Fatalf("failed to connect to mongo: %v", err)
	}
	defer mongoClient.Disconnect(context.Background())

	healthHandler := health.NewHealthHandler(postgresDb, mongoClient)

	mongoDb := mongoClient.Database(cfg.MongoDatabase)

	// Repositories
	pgProductRep := product.NewProductRepository(postgresDb)
	pgOutboxRep := outbox.NewOutboxRepository(postgresDb)
	mongoProductRep := mongorepo.NewProductRepository(mongoDb)
	productEventPublisher, err := kafka.NewProductEventPublisher(cfg.KafkaBrokers, cfg.KafkaTopic)
	if err != nil {
		log.Fatalf("failed to create product event publisher: %v", err)
	}
	defer productEventPublisher.Close() // Close the producer when the application exits

	// --- Application (use cases) ---
	// Commands (Exist in Postgres = source of truth of writing)
	createProductUC := service.NewCreateProduct(pgProductRep, pgOutboxRep, postgresDb)
	updateProductUC := service.NewUpdateProduct(pgProductRep, pgProductRep, pgProductRep, pgOutboxRep, postgresDb)
	deleteProductUC := service.NewDeleteProduct(pgProductRep, pgProductRep, pgOutboxRep, postgresDb)
	publishOutboxUC := service.NewPublishOutbox(pgOutboxRep, pgOutboxRep, productEventPublisher)

	// Queries (Exist/Load in Mongo = read model)
	findUC := service.NewGetProductById(mongoProductRep, mongoProductRep)
	searchUC := service.NewSearchProduct(mongoProductRep)
	byCategoryUC := service.NewGetProductsByCategory(mongoProductRep)
	consumeProductUC := service.NewConsumeProduct(mongoProductRep, mongoProductRep) // mongoRepository is both saver and deleter

	// --- Adapters in (HTTP) Handlers ---
	// Commands
	createProductH := command.NewCreateProductHandler(createProductUC)
	updateProductH := command.NewUpdateProductHandler(updateProductUC)
	deleteProductH := command.NewDeleteProductHandler(deleteProductUC)

	// consumer of kafka events
	consumeProductH, err := b_kafka.NewProductEventConsumer(cfg.KafkaBrokers, cfg.KafkaConsumerGroup, cfg.KafkaTopic, consumeProductUC)
	if err != nil {
		log.Fatalf("failed to create product event consumer: %v", err)
	}
	defer consumeProductH.Close()

	appCtx, stopApp := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stopApp()

	// consume the events
	go consumeProductH.ConsumeProductEvent(appCtx)

	// publish the outbox events (short timeout so a missing Kafka does not hang the relay forever)
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-appCtx.Done():
				return
			case <-ticker.C:
				relayCtx, cancel := context.WithTimeout(appCtx, 10*time.Second)
				err := publishOutboxUC.Execute(relayCtx, 50)
				cancel()
				if err != nil {
					log.Printf("outbox relay: tick failed: %v", err)
				}
			}
		}
	}()

	// Queries
	findH := query.NewFindProductHandler(findUC)
	searchH := query.NewSearchProductHandler(searchUC)
	byCategoryH := query.NewCategoryProductListerHandler(byCategoryUC)
	// --- Router (HTTP) ---
	mux := http.NewServeMux()

	// The path /health to check the health of the service
	mux.HandleFunc("/health", healthHandler.Handle)

	// The path /products to create products
	mux.HandleFunc("/products", createProductH.Handle)

	// The same path /products/{id}, different HTTP method → one handler that dispatches
	mux.HandleFunc("/products/{id}", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			findH.Handle(w, r)
		case http.MethodDelete:
			deleteProductH.Handle(w, r)
		case http.MethodPatch:
			updateProductH.Handle(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Your handlers read query params (?query=, ?category_id=), no path params
	mux.HandleFunc("/products/search", searchH.Handle)       // GET /products/search?query=...
	mux.HandleFunc("/products/category", byCategoryH.Handle) // GET /products/category?category_id=...

	srv := &http.Server{
		Addr:    cfg.HTTPAddr,
		Handler: mux,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("failed to listen and serve: %v", err)
		}
	}()

	<-appCtx.Done()

	ctxServer, stopHttp := context.WithTimeout(context.Background(), 5*time.Second)
	defer stopHttp()

	srv.Shutdown(ctxServer)
}
