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
    CREATE USER moodie WITH PASSWORD 'moodie';
    CREATE DATABASE $database ENCODING 'UTF-8' LC_COLLATE='en_US.UTF-8' LC_CTYPE='en_US.UTF-8' TEMPLATE template0 OWNER moodie;
EOF

  cat <<EOF | psql -h127.0.0.1 -U postgres $database
    GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO moodie;
    GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO moodie;
    ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL PRIVILEGES ON TABLES TO moodie;
    ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL PRIVILEGES ON SEQUENCES TO moodie;
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

function migrate() {
  if [[ "$database" =~ "test" ]]; then
    ./bin/goose -config=./etc/api.test.conf up
  else
    ./bin/goose -config=./etc/api.conf up
  fi
}

if [ $# -lt 2 ]; then
  usage
fi

cd $(dirname "$0")/../..
pwd

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
    migrate
    ;;
  *)
    echo "no such operation"
    usage
    ;;
esac

