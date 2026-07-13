package storage

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"net/url"
	"path/filepath"
	"runtime"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func InitDB() *sql.DB {
	// DSN matches the exact credentials from docker-compose.yml

	sourceFile := ""
	if _, file, _, ok := runtime.Caller(0); ok {
		sourceFile = file
		repoRoot := filepath.Clean(filepath.Join(filepath.Dir(sourceFile), "..", "..", ".."))
		envPath := filepath.Join(repoRoot, ".env")
		if err := godotenv.Load(envPath); err != nil {
			log.Println("No .env file found, falling back to system environment variables")
		} else {
			log.Printf("Loaded environment from %s", envPath)
		}
	} else if err := godotenv.Load(".env"); err != nil {
		log.Println("No .env file found, falling back to system environment variables")
	}

	// Safely read strings out of memory instead of hardcoding them
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	dbname := os.Getenv("DB_NAME")

	//  Construct the connection string dynamically
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", 
		user, password, host, port, dbname)
	
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Could not connect to database: %v", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("Database unreachable: %v", err)
	}

	fmt.Println("Successfully connected to PostgreSQL container via Environment Variables!")

	// Run automated migrations
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Fatalf("Could not create migration driver: %v", err)
	}

	migrationsDir := filepath.Join(filepath.Dir(sourceFile), "migrations")
	sourceURL := (&url.URL{Scheme: "file", Path: filepath.ToSlash(migrationsDir)}).String()

	m, err := migrate.NewWithDatabaseInstance(
		sourceURL,
		"postgres", driver,
	)
	if err != nil {
		log.Fatalf("Migration initialization failed: %v", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	fmt.Println("Database schemas are up to date.")
	return db
}