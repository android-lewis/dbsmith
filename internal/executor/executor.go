package executor

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/android-lewis/dbsmith/internal/constants"
	"github.com/android-lewis/dbsmith/internal/db"
	"github.com/android-lewis/dbsmith/internal/logging"
	"github.com/android-lewis/dbsmith/internal/models"
)

type QueryExecutor struct {
	driver     db.Driver
	maxResults int
	timeout    time.Duration
}

func NewQueryExecutor(driver db.Driver) *QueryExecutor {
	return &QueryExecutor{
		driver:     driver,
		maxResults: constants.DefaultMaxResults,
		timeout:    constants.DefaultTimeout,
	}
}

func (qe *QueryExecutor) GetDriver() db.Driver {
	return qe.driver
}

func (qe *QueryExecutor) ExecuteQuery(ctx context.Context, sql string) (*models.QueryResult, error) {
	if !qe.driver.IsConnected() {
		logging.Error().Msg("Attempted to execute query on disconnected database")
		return nil, constants.ErrNotConnected
	}

	logging.Debug().Str("sql", sql).Msg("Executing query")

	ctx, cancel := context.WithTimeout(ctx, qe.timeout)
	defer cancel()

	start := time.Now()
	result, err := qe.driver.ExecuteQuery(ctx, sql)
	duration := time.Since(start)

	if err != nil {
		if errors.Is(err, context.Canceled) {
			logging.Info().Msg("Query cancelled by user")
			return nil, constants.ErrQueryCancelled
		}
		if errors.Is(err, context.DeadlineExceeded) {
			logging.Warn().
				Dur("duration", duration).
				Dur("timeout", qe.timeout).
				Msg("Query execution timeout")
			return nil, fmt.Errorf("%w: query took longer than %v", constants.ErrQueryTimeout, qe.timeout)
		}
		logging.Error().Err(err).Str("sql", sql).Msg("Query execution failed")
		return nil, fmt.Errorf("query execution failed: %w", err)
	}

	result.ExecutionMs = duration.Milliseconds()
	logging.Info().
		Int64("execution_ms", result.ExecutionMs).
		Msg("Query executed successfully")

	return result, nil
}

func (qe *QueryExecutor) GetQueryExecutionPlan(ctx context.Context, sql string) (*models.QueryResult, error) {
	if !qe.driver.IsConnected() {
		return nil, constants.ErrNotConnected
	}

	ctx, cancel := context.WithTimeout(ctx, qe.timeout)
	defer cancel()

	return qe.driver.GetQueryExecutionPlan(ctx, sql)
}

func (qe *QueryExecutor) ExecuteNonQuery(ctx context.Context, sql string) (int64, error) {
	if !qe.driver.IsConnected() {
		return 0, constants.ErrNotConnected
	}

	ctx, cancel := context.WithTimeout(ctx, qe.timeout)
	defer cancel()

	rowsAffected, err := qe.driver.ExecuteNonQuery(ctx, sql)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return 0, fmt.Errorf("%w: query took longer than %v", constants.ErrQueryTimeout, qe.timeout)
		}
		return 0, fmt.Errorf("non-query execution failed: %w", err)
	}

	return rowsAffected, nil
}

func (qe *QueryExecutor) ExecuteTransaction(ctx context.Context, queries []string) error {
	if !qe.driver.IsConnected() {
		logging.Error().Msg("Attempted to execute transaction on disconnected database")
		return constants.ErrNotConnected
	}

	if len(queries) == 0 {
		logging.Warn().Msg("Attempted to execute empty transaction")
		return errors.New("no queries provided for transaction")
	}

	logging.Info().Int("query_count", len(queries)).Msg("Executing transaction")

	ctx, cancel := context.WithTimeout(ctx, qe.timeout)
	defer cancel()

	start := time.Now()
	if err := qe.driver.ExecuteTransaction(ctx, queries); err != nil {
		duration := time.Since(start)
		if errors.Is(err, context.DeadlineExceeded) {
			logging.Warn().
				Dur("duration", duration).
				Dur("timeout", qe.timeout).
				Int("query_count", len(queries)).
				Msg("Transaction timeout")
			return fmt.Errorf("%w: transaction took longer than %v", constants.ErrQueryTimeout, qe.timeout)
		}
		logging.Error().Err(err).Int("query_count", len(queries)).Msg("Transaction failed")
		return fmt.Errorf("transaction failed: %w", err)
	}

	logging.Info().
		Int("query_count", len(queries)).
		Dur("duration", time.Since(start)).
		Msg("Transaction executed successfully")

	return nil
}

func (qe *QueryExecutor) Ping(ctx context.Context) error {
	if !qe.driver.IsConnected() {
		return constants.ErrNotConnected
	}

	ctx, cancel := context.WithTimeout(ctx, constants.PingTimeout)
	defer cancel()

	return qe.driver.Ping(ctx)
}

func (qe *QueryExecutor) SetMaxResults(max int) {
	qe.maxResults = max
}

func (qe *QueryExecutor) SetTimeout(timeout time.Duration) {
	qe.timeout = timeout
}

func (qe *QueryExecutor) GetTimeout() time.Duration {
	return qe.timeout
}

func (qe *QueryExecutor) IsConnected() bool {
	return qe.driver.IsConnected()
}

func (qe *QueryExecutor) Close(ctx context.Context) error {
	return qe.driver.Disconnect(ctx)
}
