package d_config

import (
	"bufio"
	"fmt"
	"net/url"
	"os"
	"strings"
)

// Config holds runtime settings loaded from the environment (and optionally .env).
type Config struct {
	HTTPAddr           string
	PostgresDSN        string
	MongoURI           string
	MongoDatabase      string
	KafkaBrokers       []string
	KafkaConsumerGroup string
	KafkaTopic         string
}

// Load fills Config from process env. It first imports .env without overriding
// variables that are already set (so Docker/K8s/CI win over the file).
func Load() Config {
	// Load .env file
	_ = loadDotEnv(".env")

	// Postgres configuration
	pgHost := getenv("POSTGRES_HOST", "localhost")
	pgPort := getenv("POSTGRES_PORT", "5432")
	pgUser := getenv("POSTGRES_USER", "product_service")
	pgPass := getenv("POSTGRES_PASSWORD", "postgres")
	pgDB := getenv("POSTGRES_DB", "command_product_db")
	pgSSL := getenv("POSTGRES_SSLMODE", "disable")

	// MongoDB configuration
	// Prefer a full URI (Atlas download / MONGODB_URI). Otherwise build from parts.
	mongoDB := getenv("MONGO_DATABASE", "query_product_db")
	mongoURIStr := firstNonEmpty(
		os.Getenv("MONGODB_URI"),
		os.Getenv("MONGO_URI"),
	)
	if mongoURIStr == "" {
		mongoHost := getenv("MONGO_HOST", "localhost")
		mongoPort := getenv("MONGO_PORT", "27017")
		mongoUser := getenv("MONGO_USER", "product_service")
		mongoPass := getenv("MONGO_PASSWORD", "mongo")
		mongoURIStr = buildMongoURI(mongoHost, mongoPort, mongoUser, mongoPass)
	}

	// Kafka configuration
	kafkaBrokers := getenv("KAFKA_BROKERS", "localhost:9092")
	kafkaConsumerGroup := getenv("KAFKA_CONSUMER_GROUP", "product-consumer-group")
	kafkaTopic := getenv("KAFKA_TOPIC", "product-events")

	return Config{
		HTTPAddr: getenv("HTTP_ADDR", ":8080"),
		PostgresDSN: fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
			pgHost, pgPort, pgUser, pgPass, pgDB, pgSSL,
		),
		MongoURI:           mongoURIStr,
		MongoDatabase:      mongoDB,
		KafkaBrokers:       strings.Split(kafkaBrokers, ","),
		KafkaConsumerGroup: kafkaConsumerGroup,
		KafkaTopic:         kafkaTopic,
	}
}

// buildMongoURI constructs a MongoDB URI with authentication.
// Atlas hosts (*.mongodb.net) use mongodb+srv (TLS, no port). Local Docker uses mongodb://host:port.
func buildMongoURI(host, port, user, pass string) string {
	query := url.Values{}
	query.Set("authSource", "admin")

	if strings.Contains(host, "mongodb.net") {
		query.Set("retryWrites", "true")
		query.Set("w", "majority")
		u := &url.URL{
			Scheme:   "mongodb+srv",
			User:     url.UserPassword(user, pass),
			Host:     host,
			Path:     "/",
			RawQuery: query.Encode(),
		}
		return u.String()
	}

	u := &url.URL{
		Scheme:   "mongodb",
		User:     url.UserPassword(user, pass),
		Host:     fmt.Sprintf("%s:%s", host, port),
		Path:     "/",
		RawQuery: query.Encode(),
	}
	return u.String()
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

// getenv gets an environment variable with a fallback.
func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// loadDotEnv sets vars from a KEY=VALUE file only if they are not already in the environment.
func loadDotEnv(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.Trim(strings.TrimSpace(value), `"'`)
		if os.Getenv(key) == "" {
			_ = os.Setenv(key, value)
		}
	}
	return scanner.Err()
}
