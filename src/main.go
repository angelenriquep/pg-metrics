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
	queriesExecuted = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "pg_stat_statements_queries_executed_total",
			Help: "Total number of executed queries from pg_stat_statements",
		},
		[]string{"query"},
	)
)

func init() {
	prometheus.MustRegister(prometheus.NewBuildInfoCollector())
}

func main() {
	// Replace with your PostgreSQL connection details
	//connStr := "postgres://myuser:mypassword@127.0.0.1:5432/lcp?sslmode=disable"
	connStr := "user=myuser dbname=mydatabase password=mypassword host=localhost sslmode=disable"

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

	// Start a goroutine to scrape pg_stat_statements data every 5 seconds
	go func() {
		for {
			scrapePgStatStatements(db)
			time.Sleep(5 * time.Second)
		}
	}()

	// Expose the metrics using a custom handler
	http.Handle("/pg_stat_metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))

	fmt.Println("Server started at :8080")
	http.ListenAndServe(":8080", nil)
}

func scrapePgStatStatements(db *sql.DB) {
	rows, err := db.Query("SELECT query, calls FROM pg_stat_statements")
	if err != nil {
		fmt.Println("Error querying pg_stat_statements:", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var query string
		var calls int
		err := rows.Scan(&query, &calls)
		if err != nil {
			fmt.Println("Error scanning row:", err)
			continue
		}

		fmt.Println("query")
		// Increment the queriesExecuted counter for each query
		queriesExecuted.WithLabelValues(query).Inc()
	}
}
