# Setup

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
