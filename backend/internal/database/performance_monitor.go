package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"
)

// PerformanceMonitor monitors database performance metrics
type PerformanceMonitor struct {
	db *gorm.DB
}

// NewPerformanceMonitor creates a new performance monitor
func NewPerformanceMonitor(db *gorm.DB) *PerformanceMonitor {
	return &PerformanceMonitor{db: db}
}

// ConnectionStats represents database connection statistics
type ConnectionStats struct {
	MaxOpenConnections int `json:"max_open_connections"`
	OpenConnections    int `json:"open_connections"`
	InUse              int `json:"in_use"`
	Idle               int `json:"idle"`
}

// QueryStats represents query performance statistics
type QueryStats struct {
	SlowQueries     int64         `json:"slow_queries"`
	AverageTime     time.Duration `json:"average_time"`
	TotalQueries    int64         `json:"total_queries"`
	CacheHitRatio   float64       `json:"cache_hit_ratio"`
}

// TableStats represents table-specific statistics
type TableStats struct {
	TableName    string `json:"table_name"`
	RowCount     int64  `json:"row_count"`
	TableSize    string `json:"table_size"`
	IndexSize    string `json:"index_size"`
	LastAnalyzed string `json:"last_analyzed"`
}

// GetConnectionStats returns current connection pool statistics
func (pm *PerformanceMonitor) GetConnectionStats() (*ConnectionStats, error) {
	sqlDB, err := pm.db.DB()
	if err != nil {
		return nil, err
	}

	stats := sqlDB.Stats()
	
	return &ConnectionStats{
		MaxOpenConnections: stats.MaxOpenConnections,
		OpenConnections:    stats.OpenConnections,
		InUse:              stats.InUse,
		Idle:               stats.Idle,
	}, nil
}

// GetSlowQueries returns queries that are running longer than threshold
func (pm *PerformanceMonitor) GetSlowQueries(thresholdSeconds int) ([]map[string]interface{}, error) {
	var slowQueries []map[string]interface{}
	
	query := `
		SELECT 
			pid,
			now() - pg_stat_activity.query_start AS duration,
			query,
			state
		FROM pg_stat_activity 
		WHERE (now() - pg_stat_activity.query_start) > interval '%d seconds'
		AND state = 'active'
		ORDER BY duration DESC;
	`
	
	rows, err := pm.db.Raw(fmt.Sprintf(query, thresholdSeconds)).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	for rows.Next() {
		var pid int
		var duration string
		var queryText string
		var state string
		
		if err := rows.Scan(&pid, &duration, &queryText, &state); err != nil {
			continue
		}
		
		slowQueries = append(slowQueries, map[string]interface{}{
			"pid":      pid,
			"duration": duration,
			"query":    queryText,
			"state":    state,
		})
	}
	
	return slowQueries, nil
}

// GetTableStats returns statistics for all tables
func (pm *PerformanceMonitor) GetTableStats() ([]TableStats, error) {
	var stats []TableStats
	
	query := `
		SELECT 
			schemaname,
			tablename,
			n_tup_ins + n_tup_upd + n_tup_del as total_operations,
			pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) as table_size,
			pg_size_pretty(pg_indexes_size(schemaname||'.'||tablename)) as index_size,
			last_analyze
		FROM pg_stat_user_tables 
		ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;
	`
	
	rows, err := pm.db.Raw(query).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	for rows.Next() {
		var schemaname, tablename, tableSize, indexSize string
		var totalOps int64
		var lastAnalyze sql.NullTime
		
		if err := rows.Scan(&schemaname, &tablename, &totalOps, &tableSize, &indexSize, &lastAnalyze); err != nil {
			continue
		}
		
		lastAnalyzed := "Never"
		if lastAnalyze.Valid {
			lastAnalyzed = lastAnalyze.Time.Format("2006-01-02 15:04:05")
		}
		
		stats = append(stats, TableStats{
			TableName:    tablename,
			RowCount:     totalOps,
			TableSize:    tableSize,
			IndexSize:    indexSize,
			LastAnalyzed: lastAnalyzed,
		})
	}
	
	return stats, nil
}

// GetCacheHitRatio returns the database cache hit ratio
func (pm *PerformanceMonitor) GetCacheHitRatio() (float64, error) {
	var cacheHitRatio float64
	
	query := `
		SELECT 
			CASE 
				WHEN (blks_hit + blks_read) = 0 THEN 0
				ELSE round(blks_hit::numeric / (blks_hit + blks_read) * 100, 2)
			END as cache_hit_ratio
		FROM pg_stat_database 
		WHERE datname = current_database();
	`
	
	err := pm.db.Raw(query).Scan(&cacheHitRatio).Error
	return cacheHitRatio, err
}

// GetIndexUsage returns index usage statistics
func (pm *PerformanceMonitor) GetIndexUsage() ([]map[string]interface{}, error) {
	var indexStats []map[string]interface{}
	
	query := `
		SELECT 
			schemaname,
			tablename,
			indexname,
			idx_scan,
			idx_tup_read,
			idx_tup_fetch
		FROM pg_stat_user_indexes 
		ORDER BY idx_scan DESC;
	`
	
	rows, err := pm.db.Raw(query).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	for rows.Next() {
		var schemaname, tablename, indexname string
		var idxScan, idxTupRead, idxTupFetch int64
		
		if err := rows.Scan(&schemaname, &tablename, &indexname, &idxScan, &idxTupRead, &idxTupFetch); err != nil {
			continue
		}
		
		indexStats = append(indexStats, map[string]interface{}{
			"schema":         schemaname,
			"table":          tablename,
			"index":          indexname,
			"scans":          idxScan,
			"tuples_read":    idxTupRead,
			"tuples_fetched": idxTupFetch,
		})
	}
	
	return indexStats, nil
}

// GetUnusedIndexes returns indexes that are not being used
func (pm *PerformanceMonitor) GetUnusedIndexes() ([]map[string]interface{}, error) {
	var unusedIndexes []map[string]interface{}
	
	query := `
		SELECT 
			schemaname,
			tablename,
			indexname,
			pg_size_pretty(pg_relation_size(indexrelid)) as index_size
		FROM pg_stat_user_indexes 
		WHERE idx_scan = 0
		AND indexname NOT LIKE '%_pkey'
		ORDER BY pg_relation_size(indexrelid) DESC;
	`
	
	rows, err := pm.db.Raw(query).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	for rows.Next() {
		var schemaname, tablename, indexname, indexSize string
		
		if err := rows.Scan(&schemaname, &tablename, &indexname, &indexSize); err != nil {
			continue
		}
		
		unusedIndexes = append(unusedIndexes, map[string]interface{}{
			"schema":     schemaname,
			"table":      tablename,
			"index":      indexname,
			"index_size": indexSize,
		})
	}
	
	return unusedIndexes, nil
}

// AnalyzeTables runs ANALYZE on all tables to update statistics
func (pm *PerformanceMonitor) AnalyzeTables() error {
	return pm.db.Exec("ANALYZE").Error
}

// VacuumTables runs VACUUM on all tables to reclaim space
func (pm *PerformanceMonitor) VacuumTables() error {
	// Note: VACUUM cannot be run inside a transaction
	sqlDB, err := pm.db.DB()
	if err != nil {
		return err
	}
	
	_, err = sqlDB.Exec("VACUUM")
	return err
}

// GetLockingQueries returns queries that are causing locks
func (pm *PerformanceMonitor) GetLockingQueries() ([]map[string]interface{}, error) {
	var lockingQueries []map[string]interface{}
	
	query := `
		SELECT 
			blocked_locks.pid AS blocked_pid,
			blocked_activity.usename AS blocked_user,
			blocking_locks.pid AS blocking_pid,
			blocking_activity.usename AS blocking_user,
			blocked_activity.query AS blocked_statement,
			blocking_activity.query AS blocking_statement
		FROM pg_catalog.pg_locks blocked_locks
		JOIN pg_catalog.pg_stat_activity blocked_activity ON blocked_activity.pid = blocked_locks.pid
		JOIN pg_catalog.pg_locks blocking_locks ON blocking_locks.locktype = blocked_locks.locktype
			AND blocking_locks.DATABASE IS NOT DISTINCT FROM blocked_locks.DATABASE
			AND blocking_locks.relation IS NOT DISTINCT FROM blocked_locks.relation
			AND blocking_locks.page IS NOT DISTINCT FROM blocked_locks.page
			AND blocking_locks.tuple IS NOT DISTINCT FROM blocked_locks.tuple
			AND blocking_locks.virtualxid IS NOT DISTINCT FROM blocked_locks.virtualxid
			AND blocking_locks.transactionid IS NOT DISTINCT FROM blocked_locks.transactionid
			AND blocking_locks.classid IS NOT DISTINCT FROM blocked_locks.classid
			AND blocking_locks.objid IS NOT DISTINCT FROM blocked_locks.objid
			AND blocking_locks.objsubid IS NOT DISTINCT FROM blocked_locks.objsubid
			AND blocking_locks.pid != blocked_locks.pid
		JOIN pg_catalog.pg_stat_activity blocking_activity ON blocking_activity.pid = blocking_locks.pid
		WHERE NOT blocked_locks.GRANTED;
	`
	
	rows, err := pm.db.Raw(query).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	for rows.Next() {
		var blockedPid, blockingPid int
		var blockedUser, blockingUser, blockedStatement, blockingStatement string
		
		if err := rows.Scan(&blockedPid, &blockedUser, &blockingPid, &blockingUser, &blockedStatement, &blockingStatement); err != nil {
			continue
		}
		
		lockingQueries = append(lockingQueries, map[string]interface{}{
			"blocked_pid":        blockedPid,
			"blocked_user":       blockedUser,
			"blocking_pid":       blockingPid,
			"blocking_user":      blockingUser,
			"blocked_statement":  blockedStatement,
			"blocking_statement": blockingStatement,
		})
	}
	
	return lockingQueries, nil
}

// StartPerformanceMonitoring starts a background goroutine to monitor performance
func (pm *PerformanceMonitor) StartPerformanceMonitoring(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			pm.logPerformanceMetrics()
		}
	}
}

// logPerformanceMetrics logs current performance metrics
func (pm *PerformanceMonitor) logPerformanceMetrics() {
	// Connection stats
	connStats, err := pm.GetConnectionStats()
	if err == nil {
		log.Printf("DB Connections - Open: %d, InUse: %d, Idle: %d", 
			connStats.OpenConnections, connStats.InUse, connStats.Idle)
	}
	
	// Cache hit ratio
	cacheHitRatio, err := pm.GetCacheHitRatio()
	if err == nil {
		log.Printf("DB Cache Hit Ratio: %.2f%%", cacheHitRatio)
		
		// Alert if cache hit ratio is low
		if cacheHitRatio < 90.0 {
			log.Printf("WARNING: Low cache hit ratio detected: %.2f%%", cacheHitRatio)
		}
	}
	
	// Check for slow queries
	slowQueries, err := pm.GetSlowQueries(5) // queries running longer than 5 seconds
	if err == nil && len(slowQueries) > 0 {
		log.Printf("WARNING: %d slow queries detected", len(slowQueries))
	}
	
	// Check for locking queries
	lockingQueries, err := pm.GetLockingQueries()
	if err == nil && len(lockingQueries) > 0 {
		log.Printf("WARNING: %d locking queries detected", len(lockingQueries))
	}
}