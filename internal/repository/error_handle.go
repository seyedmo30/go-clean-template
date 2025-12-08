package repository

import (
	"__MODULE__/pkg"
	"context"
	"database/sql"
	"errors"
	"runtime/debug"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
	// <- REPLACE with your AppError package path
)

// NewAppErrorFromDBErr converts common DB/GORM/pgx errors into your AppError type
// It uses pkg.CustomAppError(...) and pkg.AddMeta(...) to populate metadata.
func NewAppErrorFromDBErr(err error) *pkg.AppError {
	if err == nil {
		return nil
	}

	// Default (fallback) app error
	app := pkg.CustomAppError(500, 1500, "internal server error", err.Error())
	app.AddMeta("error", err.Error())
	app.AddMeta("stack", string(debug.Stack()))

	// 1) GORM not found / sql no rows
	if errors.Is(err, gorm.ErrRecordNotFound) || errors.Is(err, sql.ErrNoRows) {
		return pkg.CustomAppError(404, 1001, "resource not found", err.Error()).AddMeta("stack", string(debug.Stack()))
	}

	// 2) context
	if errors.Is(err, context.DeadlineExceeded) {
		return pkg.CustomAppError(504, 1503, "request timed out", err.Error()).AddMeta("stack", string(debug.Stack()))
	}
	if errors.Is(err, context.Canceled) {
		return pkg.CustomAppError(499, 1500, "request cancelled", err.Error()).AddMeta("stack", string(debug.Stack()))
	}

	// 3) pgconn.PgError (Postgres SQLSTATE)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		// create base app error using PG error string as description
		app = pkg.CustomAppError(500, 1500, "database error", pgErr.Error())

		// attach PG diagnostics to meta
		app.AddMeta("pg_code", pgErr.Code).
			AddMeta("pg_message", pgErr.Message).
			AddMeta("pg_detail", pgErr.Detail).
			AddMeta("pg_hint", pgErr.Hint).
			AddMeta("pg_table", pgErr.TableName).
			AddMeta("pg_column", pgErr.ColumnName).
			AddMeta("pg_constraint", pgErr.ConstraintName).
			AddMeta("stack", string(debug.Stack()))

		// map common SQLSTATE codes to friendly messages / external codes
		switch pgErr.Code {
		case "23505": // unique_violation
			app = pkg.CustomAppError(409, 1002, "duplicate resource", pgErr.Error()).
				AddMeta("constraint", pgErr.ConstraintName).
				AddMeta("pg_code", pgErr.Code)
			return app

		case "23503": // foreign_key_violation
			app = pkg.CustomAppError(409, 1002, "invalid reference (foreign key)", pgErr.Error()).
				AddMeta("pg_code", pgErr.Code)
			return app

		case "23502": // not_null_violation
			app = pkg.CustomAppError(400, 1400, "missing required field", pgErr.Error()).
				AddMeta("column", pgErr.ColumnName).
				AddMeta("pg_code", pgErr.Code)
			return app

		case "22001": // string_data_right_truncation
			// your logged error was this one: value too long for varchar(20)
			app = pkg.CustomAppError(400, 1400, "field value too long", pgErr.Error()).
				AddMeta("column", pgErr.ColumnName).
				AddMeta("pg_code", pgErr.Code)
			return app

		case "22P02": // invalid_text_representation (bad uuid/int)
			app = pkg.CustomAppError(400, 1400, "invalid parameter format", pgErr.Error()).
				AddMeta("pg_code", pgErr.Code)
			return app

		case "23514": // check_violation
			app = pkg.CustomAppError(400, 1400, "value violates a constraint", pgErr.Error()).
				AddMeta("pg_code", pgErr.Code)
			return app

		case "40001", "40P01": // serialization_failure / deadlock
			app = pkg.CustomAppError(409, 1600, "database transactional error; retry", pgErr.Error()).
				AddMeta("pg_code", pgErr.Code)
			return app

		case "53300", "57P03": // too many connections / cannot connect now
			app = pkg.CustomAppError(503, 1501, "database unavailable", pgErr.Error()).
				AddMeta("pg_code", pgErr.Code)
			return app

		case "28P01": // invalid_password
			app = pkg.CustomAppError(500, 1401, "invalid database credentials", pgErr.Error()).
				AddMeta("pg_code", pgErr.Code)
			return app

		default:
			// fallback for unknown pg errors
			return app
		}
	}

	// 4) GORM-level errors
	if errors.Is(err, gorm.ErrInvalidTransaction) {
		return pkg.CustomAppError(500, 1500, "invalid database transaction", err.Error()).AddMeta("stack", string(debug.Stack()))
	}

	// 5) fallback generic error (already created above)
	return app
}

// Helper predicates

// IsUniqueViol
