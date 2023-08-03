-- Enable pg_stat_statements
CREATE EXTENSION IF NOT EXISTS pg_stat_statements;

-- Adjust the pg_stat_statements settings
ALTER SYSTEM SET pg_stat_statements.track = 'all';

-- Enable pg_stat_bgwriter and pg_stat_wal
ALTER SYSTEM SET shared_preload_libraries = 'pg_stat_statements,pg_stat_bgwriter,pg_stat_wal';

-- Adjust the pg_stat_bgwriter settings
ALTER SYSTEM SET bgwriter_delay = 100;

-- Restart the PostgreSQL instance for changes to take effect
SELECT pg_reload_conf();

-- Create a function to expose pg_stat_statements as a view
CREATE OR REPLACE FUNCTION pg_stat_statements() RETURNS SETOF pg_stat_statements AS
$$ SELECT * FROM public.pg_stat_statements; $$
LANGUAGE SQL;

-- Expose the function as a view for Prometheus scraping
CREATE OR REPLACE VIEW pg_stat AS
SELECT * FROM pg_stat_statements();
