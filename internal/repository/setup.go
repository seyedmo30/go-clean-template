package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"__MODULE__/internal/config"
	"__MODULE__/internal/interfaces"
	"__MODULE__/pkg"

	mySqlDriver "github.com/go-sql-driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Database connection pool and thread-safety mechanisms
var (
	databaseConfig config.App
	db             *gorm.DB
	reconnectMu    sync.Mutex
)

// DB ensures a singleton instance of the database connection.
func DB() *gorm.DB {

	return db
}

// serviceRepository is the structure holding DB configuration for repository operations.
type serviceRepository struct {
	config config.App
}

// Ensure serviceRepository implements the Repository interface.
var _ interfaces.Repository = (*serviceRepository)(nil)

// NewServiceRepository initializes a new serviceRepository with the provided configuration.
func NewServiceRepository(config config.App) *serviceRepository {
	databaseConfig = config // Store the config for the singleton
	var err error
	// Initialize the database only once
	db, err = SetupDB(databaseConfig)
	if err != nil {

		fmt.Printf("Error initializing DB", "error", err.Error())
	} // Ensure DB instance is initialized

	return &serviceRepository{config: config}
}

// SetupDB configures and returns a GORM DB connection using the provided database config.
func SetupDB(config config.App) (*gorm.DB, error) {
	// Construct the DSN (Data Source Name) for PostgreSQL connection
	// Example: "host=localhost user=postgres password=secret dbname=mydb port=5432 sslmode=disable TimeZone=UTC"
	dataSourceName := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
		config.Host, config.Username, config.Password, config.Database, config.Port,
	)

	// Open the underlying SQL connection
	sqlDb, err := sql.Open("pgx", dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("failed to open SQL connection: %w", err)
	}

	// Set connection pool settings
	sqlDb.SetMaxOpenConns(config.PoolSize)
	sqlDb.SetMaxIdleConns(config.MaxIdle)

	// Ping the database to verify connection
	if err := sqlDb.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Default GORM configuration
	gormConfig := &gorm.Config{
		PrepareStmt: true,
		Logger:      logger.Default.LogMode(logger.Warn),
	}

	// Environment-specific configuration
	switch config.AppEnv {
	case pkg.AppEnvDevelopment:
		baseLogger := logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags),
			logger.Config{
				SlowThreshold:             200 * time.Millisecond,
				LogLevel:                  logger.Info,
				IgnoreRecordNotFoundError: true,
				Colorful:                  true,
			},
		)

		ignoreQueries := []string{
			"payment_request_queue AS prq", // Add custom patterns to suppress logging
		}

		gormConfig = &gorm.Config{
			PrepareStmt: true,
			Logger:      NewCustomGormLogger(baseLogger, ignoreQueries),
		}

	case pkg.AppEnvProduction:
		// Keep default warning-level logger
	}

	// Initialize GORM with PostgreSQL driver
	db, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDb}), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to open GORM connection: %w", err)
	}

	return db, nil
}

// Reconnect attempts to re-establish a database connection if the current one is lost.
func Reconnect() error {
	reconnectMu.Lock()
	defer reconnectMu.Unlock()

	// If DB is nil, return immediately
	if db == nil {
		return fmt.Errorf("no active database connection to reconnect")
	}

	// Check if the connection is alive by executing a simple query
	if err := db.Exec("SELECT 1").Error; err == nil {
		// Connection is still alive
		return nil
	}

	// If connection is lost, close the current SQL connection and attempt a new connection
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying SQL DB: %w", err)
	}

	// Close the broken connection
	sqlDB.Close()

	// Attempt to reconnect by calling SetupDB
	_, err = SetupDB(databaseConfig)
	return err
}

// AutoReconnect retries database reconnection up to a given number of attempts with a delay between retries.
func AutoReconnect(retries int, delay time.Duration) error {
	var err error
	for i := 0; i < retries; i++ {
		err = Reconnect()
		if err == nil {
			return nil
		}
		i = i + 1
		// pkg.GetLogger().Error("Reconnect attempt %d/%d failed: %v", i, retries, err)
		time.Sleep(delay) // Retry after the specified delay
	}
	return fmt.Errorf("failed to reconnect after %d attempts: %v", retries, err)
}

// GetDB returns the current database connection, ensuring it is initialized.
func GetDB() (*sql.DB, error) {
	if db == nil {
		return nil, fmt.Errorf("database connection is not initialized")
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying SQL DB: %w", err)
	}

	return sqlDB, nil
}

// handleMysqlErrors is a method that handles MySQL errors and returns an appropriate ErrorWithCode.
// It checks the MySQL error number and returns a specific ErrorWithCode based on the error type.
// For example, it handles duplicate entry errors, foreign key violations, and other MySQL errors.
// If the error is not a MySQL error, it checks if the error message is "record not found" and returns a NotFoundRepositoryMessage.
// For any other errors, it returns an InternalServerErrorRepositoryMessage.
func (r *serviceRepository) handleMysqlErrors(err error) *pkg.AppError {
	if mysqlErr, ok := err.(*mySqlDriver.MySQLError); ok {
		// Handle duplicate entry error (Error 1062)
		if mysqlErr.Number == 1062 {
			// return pkg.ErrDuplicateEntry.AddStack().AddDescription(mysqlErr.Error())
			return pkg.NewAppError(pkg.ErrBadRequest)
		} else if (mysqlErr.Number == 1451) || (mysqlErr.Number == 1452) {
			// return pkg.ErrForeignKeyViolation.AddStack().AddDescription(mysqlErr.Error())
			return pkg.NewAppError(pkg.ErrBadRequest)
		} else {
			// Handle other MySQL errors
			// return pkg.ErrInternalServerError.AddStack().AddDescription(mysqlErr.Error())
			return pkg.NewAppError(pkg.ErrBadRequest)
		}
	} else if err.Error() == "record not found" {
		// return pkg.ErrRecordNotFound.AddStack().AddDescription(err.Error())
		return pkg.NewAppError(pkg.ErrBadRequest)
	} else {
		// Handle other errors
		// return pkg.ErrInternalServerError.AddStack().AddDescription(err.Error())
		return pkg.NewAppError(pkg.ErrBadRequest)
	}
}

type CustomGormLogger struct {
	logger.Interface
	IgnoreQueries []string
}

func NewCustomGormLogger(baseLogger logger.Interface, ignoreQueries []string) *CustomGormLogger {
	return &CustomGormLogger{
		Interface:     baseLogger,
		IgnoreQueries: ignoreQueries,
	}
}

func (l *CustomGormLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	sql, rows := fc()

	// Check if this query should be ignored
	for _, ignoreQuery := range l.IgnoreQueries {
		if strings.Contains(sql, ignoreQuery) {
			return
		}
	}

	// If not ignored, pass to the underlying logger
	l.Interface.Trace(ctx, begin, func() (string, int64) {
		return sql, rows
	}, err)
}

func (r *serviceRepository) handleDBErrors(err error) *pkg.AppError {
	if err == nil {
		return nil
	} else if err.Error() == "record not found" {
		// return pkg.ErrRecordNotFound.AddStack().AddDescription(err.Error())
		return pkg.NewAppError(pkg.ErrBadRequest)
	} else {
		// return pkg.ErrInternalServerError.AddStack().AddDescription(err.Error())
		return pkg.NewAppError(pkg.ErrBadRequest)

	}
}
