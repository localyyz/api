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
    CREATE EXTENSION postgis;
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
    DROP DATABASE $database;
EOF
}

function loadprod() {
  echo "LOADING DATABASE FROM PROD BACKUP";
  scp root@localyyz:/root/backups/backup-`date +%F`.gz tmp/.
  gunzip -c tmp/backup-`date +%F`.gz | psql $database
  rm tmp/backup-`date +%F`.gz
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
  "loadprod")
    drop
    create
    loadprod
    ;;
  *)
    echo "no such operation"
    usage
    ;;
esac

