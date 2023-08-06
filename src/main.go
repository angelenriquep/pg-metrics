package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	_ "github.com/lib/pq"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	queriesExecuted = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "pg_stat_statements_queries_executed_total",
			Help: "Total number of executed queries from pg_stat_statements",
		},
		[]string{"query"},
	)

	top10ExtensiveCalls = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "top_10_extensive_calls",
			Help: "Top 10 extensive calls from pg_stat_statements",
		},
		[]string{"userid", "dbid", "query"},
	)

	top10TimeConsumingQueries = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "top_10_time_consuming_queries",
			Help: "Top 10 time-consuming queries from pg_stat_statements",
		},
		[]string{"userid", "dbid", "query"},
	)

	top10MemConsuming = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "top_10_mem_consuming_queries",
			Help: "Top 10 memory-consuming queries from pg_stat_statements",
		},
		[]string{"userid", "dbid", "query"},
	)

	top10TempConsumers = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "top_10_temp_consumers",
			Help: "Top 10 consumers of temporary space from pg_stat_statements",
		},
		[]string{"userid", "dbid", "query"},
	)

	bgwriterStats = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "pg_stat_bgwriter_stats",
			Help: "Statistics from pg_stat_bgwriter view",
		},
		[]string{"stat"},
	)

	replicationStats = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "pg_stat_replication_stats",
			Help: "Statistics from pg_stat_replication view",
		},
		[]string{"application_name"},
	)

	activityStats = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "pg_stat_activity_stats",
			Help: "Statistics from pg_stat_activity view",
		},
		[]string{"application_name"},
	)
)

func main() {
	// Replace with your PostgreSQL connection details
	connStr := "user=myuser dbname=mydatabase password=mypassword host=host.docker.internal sslmode=disable"

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		fmt.Println(err, db)
		panic(err)
	}

	// Create a new Prometheus registry
	reg := prometheus.NewRegistry()
	reg.MustRegister(queriesExecuted)
	reg.MustRegister(top10ExtensiveCalls)
	reg.MustRegister(top10TimeConsumingQueries)
	reg.MustRegister(top10MemConsuming)
	reg.MustRegister(top10TempConsumers)
	reg.MustRegister(bgwriterStats)
	reg.MustRegister(replicationStats)
	reg.MustRegister(activityStats)

	// Start a goroutine to scrape pg_stat_statements data every 5 seconds
	go func() {
		for {
			scrapePgStatStatements(db)
			scrapeBgwriterStats(db)
			scrapeReplicationStats(db)
			scrapeActivityStats(db)
			time.Sleep(5 * time.Second)
		}
	}()

	// Expose the metrics using a custom handler
	http.Handle("/pg_stat_metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))

	fmt.Println("Server started at :8080")
	// Blocking code
	http.ListenAndServe(":8080", nil)
}

// Run additional queries and update respective Prometheus metrics
func scrapePgStatStatements(db *sql.DB) {
	//runAndCollectTop10Metrics(db, queriesExecuted, "SELECT query, calls FROM pg_stat_statements")

	// Run additional queries and update respective Prometheus metrics
	runAndCollectTop10Metrics(db, top10ExtensiveCalls, `
        select userid::regrole, dbid, query
        from pg_stat_statements
        order by (blk_read_time+blk_write_time)/calls desc
        limit 10;
    `)

	runAndCollectTop10Metrics(db, top10TimeConsumingQueries, `
        select userid::regrole, dbid, query
        from pg_stat_statements
        order by mean_exec_time desc
        limit 10;
    `)

	runAndCollectTop10Metrics(db, top10MemConsuming, `
        select userid::regrole, dbid, query
        from pg_stat_statements
        order by (shared_blks_hit+shared_blks_dirtied) desc
        limit 10;
    `)

	runAndCollectTop10Metrics(db, top10TempConsumers, `
        select userid::regrole, dbid, query
        from pg_stat_statements
        order by temp_blks_written desc
        limit 10;
    `)
}

func scrapeBgwriterStats(db *sql.DB) {
	rows, err := db.Query("SELECT * FROM pg_stat_bgwriter LIMIT 1")
	if err != nil {
		fmt.Println("Error querying pg_stat_bgwriter:", err)
		return
	}
	defer rows.Close()

	if rows.Next() {
		var checkpointsTimed, checkpointsReq, checkpointWriteTime,
			checkpointSyncTime, buffersAlloc, buffersBackend, buffersBackendFsync,
			buffersAllocated, buffersHit, buffersRead, buffersWritten sql.NullInt64

		err := rows.Scan(&checkpointsTimed, &checkpointsReq, &checkpointWriteTime,
			&checkpointSyncTime, &buffersAlloc, &buffersBackend, &buffersBackendFsync,
			&buffersAllocated, &buffersHit, &buffersRead, &buffersWritten)
		if err != nil {
			fmt.Println("Error scanning pg_stat_bgwriter row:", err)
			return
		}

		// Set the metrics for bgwriterStats
		bgwriterStats.WithLabelValues("checkpoints_timed").Set(float64(checkpointsTimed.Int64))
		bgwriterStats.WithLabelValues("checkpoints_req").Set(float64(checkpointsReq.Int64))
		bgwriterStats.WithLabelValues("checkpoint_write_time").Set(float64(checkpointWriteTime.Int64))
		bgwriterStats.WithLabelValues("checkpoint_sync_time").Set(float64(checkpointSyncTime.Int64))
		bgwriterStats.WithLabelValues("buffers_checkpoint").Set(float64(buffersAlloc.Int64))
		bgwriterStats.WithLabelValues("buffers_clean").Set(float64(buffersBackend.Int64))
		bgwriterStats.WithLabelValues("maxwritten_clean").Set(float64(buffersBackendFsync.Int64))
		bgwriterStats.WithLabelValues("buffers_backend").Set(float64(buffersAllocated.Int64))
		bgwriterStats.WithLabelValues("buffers_backend_fsync").Set(float64(buffersHit.Int64))
		bgwriterStats.WithLabelValues("buffers_alloc").Set(float64(buffersRead.Int64))
		// bgwriterStats.WithLabelValues("stats_reset").Set(float64(buffersWritten.Int64))
	}
}

func scrapeReplicationStats(db *sql.DB) {
	rows, err := db.Query("SELECT application_name, state, sync_state FROM pg_stat_replication")
	if err != nil {
		fmt.Println("Error querying pg_stat_replication:", err)
		return
	}
	defer rows.Close()

	// Reset the replicationStats metric before updating
	replicationStats.Reset()

	for rows.Next() {
		var applicationName, state, syncState string
		err := rows.Scan(&applicationName, &state, &syncState)
		if err != nil {
			fmt.Println("Error scanning pg_stat_replication row:", err)
			continue
		}

		// Set the metrics for replicationStats
		replicationStats.WithLabelValues(applicationName).Set(1)
	}
}

func scrapeActivityStats(db *sql.DB) {
	rows, err := db.Query("SELECT application_name FROM pg_stat_activity WHERE state = 'active'")
	if err != nil {
		fmt.Println("Error querying pg_stat_activity:", err)
		return
	}
	defer rows.Close()

	// Reset the activityStats metric before updating
	activityStats.Reset()

	for rows.Next() {
		var applicationName string
		err := rows.Scan(&applicationName)
		if err != nil {
			fmt.Println("Error scanning pg_stat_activity row:", err)
			continue
		}

		// Set the metrics for activityStats
		activityStats.WithLabelValues(applicationName).Set(1)
	}
}

func runAndCollectTop10Metrics(db *sql.DB, metric *prometheus.GaugeVec, query string) {
	rows, err := db.Query(query)
	if err != nil {
		fmt.Println("Error querying:", err)
		return
	}
	defer rows.Close()

	// Reset the metric before updating
	metric.Reset()

	for rows.Next() {
		var userid, dbid, queryText string
		err := rows.Scan(&userid, &dbid, &queryText)
		if err != nil {
			fmt.Println("Error scanning row:", err)
			continue
		}

		// Set the metric value with labels
		metric.WithLabelValues(userid, dbid, queryText).Set(1)
	}
}
