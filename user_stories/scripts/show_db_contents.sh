#!/usr/bin/env bash
#
# Pretty database snapshot for the Hypertube Postgres database.
#
# Usage:
#   ./user_stories/scripts/show_db_contents.sh
#
# Optional environment variables:
#   DB_BACKEND=auto|docker|psql   Connection mode. Default: auto
#   DATABASE_URL                  Used with DB_BACKEND=psql
#   POSTGRES_USER                 Used with DB_BACKEND=docker. Default: hypertube
#   POSTGRES_DB                   Used with DB_BACKEND=docker. Default: hypertube
#   ROW_LIMIT=0                   Rows per table. 0 means all rows. Default: 0
#   FULL_DUMP=1                   Include full JSON rows for every table. Default: 1
#
# Examples:
#   ./user_stories/scripts/show_db_contents.sh
#   ROW_LIMIT=25 ./user_stories/scripts/show_db_contents.sh
#   FULL_DUMP=0 ./user_stories/scripts/show_db_contents.sh
#   DB_BACKEND=psql DATABASE_URL='postgres://hypertube:changeme@localhost:5432/hypertube?sslmode=disable' ./user_stories/scripts/show_db_contents.sh

set -o pipefail

DB_BACKEND="${DB_BACKEND:-auto}"
ROW_LIMIT="${ROW_LIMIT:-0}"
FULL_DUMP="${FULL_DUMP:-1}"

if [[ -t 1 ]] && command -v tput >/dev/null 2>&1 && [[ "$(tput colors 2>/dev/null || echo 0)" -ge 8 ]]; then
  BOLD="$(tput bold)"
  DIM="$(tput dim)"
  RED="$(tput setaf 1)"
  GREEN="$(tput setaf 2)"
  CYAN="$(tput setaf 6)"
  RESET="$(tput sgr0)"
else
  BOLD=""
  DIM=""
  RED=""
  GREEN=""
  CYAN=""
  RESET=""
fi

usage() {
  sed -n '2,20p' "$0" | sed 's/^# \{0,1\}//'
}

if [[ "${1:-}" == "-h" || "${1:-}" == "--help" ]]; then
  usage
  exit 0
fi

fail() {
  printf '%sError:%s %s\n' "$RED" "$RESET" "$1" >&2
  exit 1
}

env_file_value() {
  local key="$1"
  local line
  local value

  [[ -f .env ]] || return 0
  line="$(grep -E "^[[:space:]]*(export[[:space:]]+)?${key}=" .env | tail -n 1 || true)"
  [[ -n "$line" ]] || return 0

  value="${line#*=}"
  value="${value%$'\r'}"
  value="${value#\"}"
  value="${value%\"}"
  value="${value#\'}"
  value="${value%\'}"
  printf '%s' "$value"
}

if ! [[ "$ROW_LIMIT" =~ ^[0-9]+$ ]]; then
  fail "ROW_LIMIT must be a non-negative integer."
fi

case "$FULL_DUMP" in
  0|1|true|false|on|off|yes|no) ;;
  *) fail "FULL_DUMP must be 0/1, true/false, on/off, or yes/no." ;;
esac

DATABASE_URL="${DATABASE_URL:-$(env_file_value DATABASE_URL)}"
POSTGRES_USER="${POSTGRES_USER:-$(env_file_value POSTGRES_USER)}"
POSTGRES_DB="${POSTGRES_DB:-$(env_file_value POSTGRES_DB)}"
POSTGRES_USER="${POSTGRES_USER:-hypertube}"
POSTGRES_DB="${POSTGRES_DB:-hypertube}"

if [[ "$ROW_LIMIT" == "0" ]]; then
  LIMIT_CLAUSE=""
  ROW_LIMIT_LABEL="all rows"
else
  LIMIT_CLAUSE="LIMIT $ROW_LIMIT"
  ROW_LIMIT_LABEL="$ROW_LIMIT rows per table"
fi

docker_postgres_running() {
  command -v docker >/dev/null 2>&1 && [[ -n "$(docker compose ps -q postgres 2>/dev/null)" ]]
}

select_backend() {
  case "$DB_BACKEND" in
    auto)
      if docker_postgres_running; then
        DB_BACKEND="docker"
      elif command -v psql >/dev/null 2>&1 && [[ -n "$DATABASE_URL" ]]; then
        DB_BACKEND="psql"
      else
        fail "No database connection found. Start Docker Compose postgres, or install psql and set DATABASE_URL."
      fi
      ;;
    docker)
      docker_postgres_running || fail "Docker Compose postgres service is not running."
      ;;
    psql)
      command -v psql >/dev/null 2>&1 || fail "psql is not installed."
      [[ -n "$DATABASE_URL" ]] || fail "DATABASE_URL is required for DB_BACKEND=psql."
      ;;
    *)
      fail "DB_BACKEND must be auto, docker, or psql."
      ;;
  esac
}

print_banner() {
  printf '%s\n' "================================================================================"
  printf '%sHypertube Database Snapshot%s\n' "$BOLD$CYAN" "$RESET"
  printf '%s\n' "================================================================================"
  printf 'Backend: %s\n' "$DB_BACKEND"
  printf 'Rows:    %s\n' "$ROW_LIMIT_LABEL"
  printf 'Mode:    %s\n' "$([[ "$FULL_DUMP" == "0" || "$FULL_DUMP" == "false" || "$FULL_DUMP" == "off" || "$FULL_DUMP" == "no" ]] && printf 'summary only' || printf 'summary plus full JSON dump')"
  printf '%s\n\n' "================================================================================"
}

snapshot_sql() {
  cat <<'SQL'
\set ON_ERROR_STOP on
\pset pager off
\pset border 2
\pset linestyle unicode
\pset null '[null]'
\pset columns 140
\x auto

\echo
\echo ============================
\echo Connection
\echo ============================
\pset title 'Database connection'
SELECT
  current_database() AS database,
  current_user AS user,
  inet_server_addr() AS server_addr,
  inet_server_port() AS server_port,
  to_char(now(), 'YYYY-MM-DD HH24:MI:SS TZ') AS snapshot_at,
  version() AS postgres_version;

\echo
\echo ============================
\echo Database Size
\echo ============================
\pset title 'Database size'
SELECT
  pg_size_pretty(pg_database_size(current_database())) AS database_size;

\echo
\echo ============================
\echo Public Tables
\echo ============================
\pset title 'Table overview'
SELECT
  n.nspname AS schema,
  c.relname AS table,
  pg_size_pretty(pg_total_relation_size(c.oid)) AS total_size,
  pg_size_pretty(pg_relation_size(c.oid)) AS table_size,
  COALESCE(s.n_live_tup, 0)::bigint AS estimated_live_rows
FROM pg_class c
JOIN pg_namespace n ON n.oid = c.relnamespace
LEFT JOIN pg_stat_user_tables s ON s.relid = c.oid
WHERE n.nspname = 'public'
  AND c.relkind = 'r'
ORDER BY c.relname;

\pset title 'Exact row counts'
SELECT COALESCE(
  string_agg(
    format('SELECT %L AS table_name, count(*)::bigint AS rows FROM %I.%I', schemaname || '.' || tablename, schemaname, tablename),
    E'\nUNION ALL\n'
    ORDER BY tablename
  ),
  'SELECT ''(no public tables)'' AS table_name, 0::bigint AS rows'
) || E'\nORDER BY table_name;' AS sql
FROM pg_tables
WHERE schemaname = 'public'
\gexec

\echo
\echo ============================
\echo Schema
\echo ============================
\pset title 'Columns'
SELECT
  table_name,
  ordinal_position AS pos,
  column_name,
  data_type,
  is_nullable,
  COALESCE(column_default, '') AS default_value
FROM information_schema.columns
WHERE table_schema = 'public'
ORDER BY table_name, ordinal_position;

\pset title 'Foreign keys'
SELECT
  tc.table_name,
  kcu.column_name,
  ccu.table_name AS references_table,
  ccu.column_name AS references_column,
  rc.update_rule,
  rc.delete_rule
FROM information_schema.table_constraints tc
JOIN information_schema.key_column_usage kcu
  ON tc.constraint_name = kcu.constraint_name
 AND tc.table_schema = kcu.table_schema
JOIN information_schema.constraint_column_usage ccu
  ON ccu.constraint_name = tc.constraint_name
 AND ccu.table_schema = tc.table_schema
JOIN information_schema.referential_constraints rc
  ON rc.constraint_name = tc.constraint_name
 AND rc.constraint_schema = tc.table_schema
WHERE tc.constraint_type = 'FOREIGN KEY'
  AND tc.table_schema = 'public'
ORDER BY tc.table_name, kcu.column_name;

SELECT
  CASE WHEN to_regclass('public.users') IS NULL THEN 'false' ELSE 'true' END AS has_users,
  CASE WHEN to_regclass('public.oauth_accounts') IS NULL THEN 'false' ELSE 'true' END AS has_oauth_accounts,
  CASE WHEN to_regclass('public.movies') IS NULL THEN 'false' ELSE 'true' END AS has_movies,
  CASE WHEN to_regclass('public.featured_movies') IS NULL THEN 'false' ELSE 'true' END AS has_featured_movies,
  CASE WHEN to_regclass('public.torrents') IS NULL THEN 'false' ELSE 'true' END AS has_torrents,
  CASE WHEN to_regclass('public.watch_history') IS NULL THEN 'false' ELSE 'true' END AS has_watch_history
\gset

\echo
\echo ============================
\echo Users
\echo ============================
\if :has_users
\pset title 'users'
SELECT
  id,
  email::text AS email,
  username,
  first_name,
  last_name,
  CASE WHEN COALESCE(password_hash, '') = '' THEN 'no' ELSE 'yes' END AS password_set,
  CASE WHEN COALESCE(password_hash, '') = '' THEN '' ELSE 'redacted (' || length(password_hash)::text || ' chars)' END AS password_hash,
  to_char(created_at, 'YYYY-MM-DD HH24:MI:SS TZ') AS created_at,
  to_char(updated_at, 'YYYY-MM-DD HH24:MI:SS TZ') AS updated_at
FROM users
ORDER BY id
:limit_clause;
\else
SELECT 'users table does not exist' AS users;
\endif

\echo
\echo ============================
\echo OAuth Accounts
\echo ============================
\if :has_oauth_accounts
\pset title 'oauth_accounts'
SELECT
  oa.id,
  oa.user_id,
  u.username,
  oa.provider,
  oa.provider_user_id,
  oa.provider_email::text AS provider_email,
  to_char(oa.created_at, 'YYYY-MM-DD HH24:MI:SS TZ') AS created_at,
  to_char(oa.updated_at, 'YYYY-MM-DD HH24:MI:SS TZ') AS updated_at
FROM oauth_accounts oa
LEFT JOIN users u ON u.id = oa.user_id
ORDER BY oa.id
:limit_clause;
\else
SELECT 'oauth_accounts table does not exist' AS oauth_accounts;
\endif

\echo
\echo ============================
\echo Movies
\echo ============================
\if :has_movies
\pset title 'movies'
SELECT
  imdbid,
  tmdbid,
  title,
  year,
  note,
  runtime_minutes AS runtime_min,
  director,
  array_to_string(genre, ', ') AS genre_ids,
  COALESCE(array_length("cast", 1), 0) AS cast_count,
  CASE WHEN length(summary) > 180 THEN left(summary, 180) || '...' ELSE summary END AS summary
FROM movies
ORDER BY title, imdbid
:limit_clause;
\else
SELECT 'movies table does not exist' AS movies;
\endif

\echo
\echo ============================
\echo Featured Movies
\echo ============================
\if :has_featured_movies
\pset title 'featured_movies'
SELECT
  f.position,
  f.imdbid,
  m.title,
  m.year,
  m.note
FROM featured_movies f
LEFT JOIN movies m ON m.imdbid = f.imdbid
ORDER BY f.position, f.imdbid
:limit_clause;
\else
SELECT 'featured_movies table does not exist' AS featured_movies;
\endif

\echo
\echo ============================
\echo Torrents
\echo ============================
\if :has_torrents
\pset title 'torrents'
SELECT
  t.id,
  t.imdbid,
  COALESCE(m.title, '') AS movie,
  t.source,
  t.quality,
  t.language,
  t.size,
  t.seeds,
  CASE WHEN length(t.title) > 90 THEN left(t.title, 90) || '...' ELSE t.title END AS torrent_title,
  CASE WHEN length(t.url) > 110 THEN left(t.url, 110) || '...' ELSE t.url END AS url
FROM torrents t
LEFT JOIN movies m ON m.imdbid = t.imdbid
ORDER BY t.id
:limit_clause;
\else
SELECT 'torrents table does not exist' AS torrents;
\endif

\echo
\echo ============================
\echo Watch History
\echo ============================
\if :has_watch_history
\pset title 'watch_history'
SELECT
  wh.id,
  wh.user_id,
  u.username,
  wh.movie_id,
  m.title AS movie_title,
  to_char(wh.watched_at, 'YYYY-MM-DD HH24:MI:SS TZ') AS watched_at
FROM watch_history wh
LEFT JOIN users u ON u.id = wh.user_id
LEFT JOIN movies m ON m.imdbid = wh.movie_id
ORDER BY wh.watched_at DESC, wh.id DESC
:limit_clause;
\else
SELECT 'watch_history table does not exist' AS watch_history;
\endif

\if :full_dump
\echo
\echo ============================
\echo Full Table Dump
\echo ============================
\echo Full JSON rows for every public table. password_hash is redacted.
\pset title 'full table dump'
SELECT COALESCE(
  string_agg(
    format(
      'SELECT %L AS table_name, jsonb_pretty(to_jsonb(t)%s) AS row_json FROM %I.%I AS t ORDER BY to_jsonb(t)::text %s;',
      schemaname || '.' || tablename,
      CASE
        WHEN EXISTS (
          SELECT 1
          FROM information_schema.columns c
          WHERE c.table_schema = pt.schemaname
            AND c.table_name = pt.tablename
            AND c.column_name = 'password_hash'
        )
        THEN ' || jsonb_build_object(''password_hash'', CASE WHEN t.password_hash IS NULL OR t.password_hash = '''' THEN '''' ELSE ''<redacted>'' END)'
        ELSE ''
      END,
      schemaname,
      tablename,
      :'limit_clause'
    ),
    E'\n'
    ORDER BY tablename
  ),
  'SELECT ''(no public tables)'' AS table_name, ''{}'' AS row_json'
) AS sql
FROM pg_tables pt
WHERE schemaname = 'public'
\gexec
\else
\echo
\echo Full table dump skipped. Run FULL_DUMP=1 ./user_stories/scripts/show_db_contents.sh to include it.
\endif

\echo
\echo ============================
\echo Done
\echo ============================
SQL
}

run_psql() {
  local common_args
  common_args=(
    -v "limit_clause=$LIMIT_CLAUSE"
    -v "row_limit_label=$ROW_LIMIT_LABEL"
    -v "full_dump=$FULL_DUMP"
  )

  case "$DB_BACKEND" in
    docker)
      snapshot_sql | docker compose exec -T -e PSQL_PAGER=cat postgres psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" "${common_args[@]}"
      ;;
    psql)
      snapshot_sql | PSQL_PAGER=cat psql "$DATABASE_URL" "${common_args[@]}"
      ;;
  esac
}

select_backend
print_banner
run_psql

printf '\n%sDone.%s\n' "$GREEN" "$RESET"
