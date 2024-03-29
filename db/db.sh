#!/bin/bash -e

function usage() {
  echo "Usage: $0 <operation> <database>"
  echo "-"
  echo "operation : [ create, drop, reset ]"
  exit 1
}

function create() {
  echo "CREATE USER postgres SUPERUSER;" | psql -h127.0.0.1 || :
  cat <<EOF | psql -h127.0.0.1 -U postgres
    CREATE USER localyyz WITH PASSWORD 'localyyz';
    CREATE DATABASE $database ENCODING 'UTF-8' LC_COLLATE='en_US.UTF-8' LC_CTYPE='en_US.UTF-8' TEMPLATE template0 OWNER localyyz;
EOF

  cat <<EOF | psql -h127.0.0.1 -U postgres $database
    CREATE EXTENSION IF NOT EXISTS pg_trgm;
    CREATE EXTENSION IF NOT EXISTS plpgsql;
    GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO localyyz;
    GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO localyyz;
    ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL PRIVILEGES ON TABLES TO localyyz;
    ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL PRIVILEGES ON SEQUENCES TO localyyz;
EOF
}

function drop() {
  cat <<EOF | psql -h127.0.0.1 -U postgres
    SELECT pg_terminate_backend(pg_stat_activity.pid)
    FROM pg_stat_activity
    WHERE pg_stat_activity.datname = '$database' AND pid <> pg_backend_pid();
    DROP DATABASE IF EXISTS $database;
EOF
}

function loadstaging() {
  echo "LOADING DATABASE FROM STAGING BACKUP";
  if [ ! -f tmp/backup-latest.dump ]; then
    echo "DOWNLOADING BACKUP";
    scp root@localyyz.staging:/data/backups/backup-latest.dump tmp/.
  elif test `find "tmp/backup-latest.dump" -mmin +2000`; then
    echo "DOWNLOADING LATEST BACKUP";
    scp root@localyyz.staging:/data/backups/backup-latest.dump tmp/.
  fi
  pg_restore -j 4 -l tmp/backup-latest.dump | sed '/MATERIALIZED VIEW DATA/d' > tmp/restore.lst
  pg_restore -L tmp/restore.lst -d localyyz tmp/backup-latest.dump

  # some reason refreshing materialized view doesn't work here
  #pg_restore -l tmp/backup-latest.dump | grep 'MATERIALIZED VIEW DATA' > tmp/refresh.lst
  #pg_restore -L tmp/refresh.lst -d localyyz tmp/backup-latest.dump
}

if [ $# -lt 2 ]; then
  usage
fi

operation="$1"
database="$2"

case "$operation" in
  "create")
    create
    ;;
  "drop")
    drop
    ;;
  "reset")
    drop
    create
    ;;
  "loadstaging")
    drop
    create
    loadstaging
    ;;
  *)
    echo "no such operation"
    usage
    ;;
esac

