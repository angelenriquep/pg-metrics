# Postgres - The Cumulative Statistics System

- Statistics Collection Configuration
- Viewing Statistics
- pg_stat_activity
- pg_stat_replication
- pg_stat_replication_slots
- pg_stat_wal_receiver
- pg_stat_recovery_prefetch
- pg_stat_subscription
- pg_stat_subscription_stats
- pg_stat_ssl
- pg_stat_gssapi
- pg_stat_archiver
- pg_stat_bgwriter
- pg_stat_wal
- pg_stat_database
- pg_stat_database_conflicts
- pg_stat_all_tables
- pg_stat_all_indexes
- pg_statio_all_tables
- pg_statio_all_indexes
- pg_statio_all_sequences
- pg_stat_user_functions
- pg_stat_slru
- Statistics Functions

PostgreSQL's cumulative statistics system supports collection and reporting of
information about server activity. Presently, accesses to tables and indexes in
both disk-block and individual-row terms are counted. The total number of rows
in each table, and information about vacuum and analyze actions for each table
are also counted. If enabled, calls to user-defined functions and the total time
spent in each one are counted as well.

## Setup

Because we are using the host and not an internal netowkr to expose the services:
host.docker.internal

The problem of this script is that beucase is pull, we may affect performance of
postgres? should this scrapper run each minute rathern than 15 secs?

```sql
psql -d mydatabase -U myuser
-- ensure the conf file is correct: sudo systemctl restart postgresql@15-main
SHOW config_file; config_file
\d schemas
\dx
-- Show activity
select count(*) from pg_stat_activity;
-- Check extension is created ok
select* from pg_extension;
-- expanded display
\x
select * from pg_settings where name = 'shared_preload_libraries'
```

SELECT query, calls FROM pg_stat_statements

Top 10 extensive calls

select userid::regrole, dbid, query
    from pg_stat_statements
    order by (blk_read_time+blk_write_time)/calls desc
    limit 10;

Top 10 time-consuming queries
    select userid::regrole, dbid, query
    from pg_stat_statements
    order by mean_exec_time desc
    limit 10;

Top 10 mem consuming
select userid::regrole, dbid, query
    from pg_stat_statements
    order by (shared_blks_hit+shared_blks_dirtied) desc
    limit 10;

Top 10 consumers of temporary
select userid::regrole, dbid, query
    from pg_stat_statements
    order by temp_blks_written desc
    limit 10;

## TODO

- Transform the metrics to make humna readable
- Failure tolerance
- Performance improv
- Monitor healthcheck, and server stats
- Check metrics from <https://github.com/cybertec-postgresql/pgwatch2/blob/master/pgwatch2/pgwatch2.go#L3600>
