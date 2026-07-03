#!/bin/bash
set -o nounset -o pipefail -o errexit

DB_TYPE=${1-sqlite}
USE_EXTERNAL_DB=${WAKAPI_TEST_EXTERNAL_DB:-0}

if ! command -v bru &> /dev/null; then
    echo "Bruno CLI could not be found. Run 'npm install -g @usebruno/cli' first."
    exit 1
fi

for i in "$@"; do
    case $i in
        --migration)
            MIGRATION=1
            ;;
    esac
done

script_path=$(realpath "${BASH_SOURCE[0]}")
script_dir=$(dirname "$script_path")

export TZ=${TZ:-Europe/Berlin}

echo "Compiling."
(cd "$script_dir/.." || exit 1; CGO_ENABLED=0 go build -o wakapi)

cd "$script_dir" || exit 1

# Download previous release (when upgrade testing)
if [ "${MIGRATION-0}" -eq 1 ]; then
    if [ ! -f wakapi_linux_amd64.zip ]; then
        echo "Downloading latest release"
        curl https://github.com/muety/wakapi/releases/latest/download/wakapi_linux_amd64.zip -O -L
    fi
    unzip -o wakapi_linux_amd64.zip
    echo "Running tests with release version"
fi

cleanup() {
    if [ -n "$pid" ] && ps -p "$pid" > /dev/null; then
        kill -TERM "$pid"
    fi
    if [ "${docker_down-0}" -eq 1 ]; then
        docker compose -f "$script_dir/compose.yml" down
    fi
}
trap cleanup EXIT

wait_for_tcp() {
    local host=$1
    local port=$2

    echo "Waiting for $host:$port ..."
    for _ in $(seq 1 60); do
        if bash -c "</dev/tcp/$host/$port" 2> /dev/null; then
            return 0
        fi
        sleep 1
    done

    echo "Timed out waiting for $host:$port"
    return 1
}

# Initialise test data
case $DB_TYPE in
    postgres|mysql|mariadb|cockroach)
    config="config.$DB_TYPE.yml"
    if [ "$DB_TYPE" == "mariadb" ]; then
        config="config.mysql.yml"
    fi

    if [ "$USE_EXTERNAL_DB" -eq 1 ]; then
        db_host=${WAKAPI_DB_HOST:-$DB_TYPE}
        if [ "$DB_TYPE" == "postgres" ]; then
            db_port=${WAKAPI_DB_PORT:-5432}
        elif [ "$DB_TYPE" == "cockroach" ]; then
            db_port=${WAKAPI_DB_PORT:-26257}
        else
            db_port=${WAKAPI_DB_PORT:-3306}
        fi
        export WAKAPI_DB_HOST=$db_host
        export WAKAPI_DB_PORT=$db_port
    else
        docker compose -f "$script_dir/compose.yml" down

        docker_down=1
        docker compose -f "$script_dir/compose.yml" up --wait --detach "$DB_TYPE"

        db_host=127.0.0.1
        if [ "$DB_TYPE" == "postgres" ]; then
            db_port=55432
        elif [ "$DB_TYPE" == "cockroach" ]; then
            db_port=56257
        else
            db_port=53306
        fi
    fi

    wait_for_tcp "$db_host" "$db_port"
    ;;

    sqlite|*)
    config="config.sqlite.yml"
    rm -f wakapi_testing.db
    ;;
esac

wait_for_wakapi () {
    counter=0
    echo "Waiting for Wakapi to come up ..."
    until curl --output /dev/null --silent --get --fail http://localhost:3000/api/health; do
        if [ "$counter" -ge 30 ]; then
            echo "Waited for 30s, but Wakapi failed to come up ..."
            exit 1
        fi

        printf '.'
        sleep 1
        counter=$((counter+1))
    done
    sleep 1
    printf "\n"
}

start_wakapi_background() {
    path=$1
    config=$2

    "$path" -config "$config" &
    pid=$!
    wait_for_wakapi
}

kill_wakapi() {
    echo "Shutting down Wakapi ..."
    kill -TERM $pid || true
}

# Run original wakapi
echo "Configuration file: $config"
if [ "${MIGRATION-0}" -eq 1 ]; then
    echo "Running last release ..."
    start_wakapi_background "./wakapi" "$config"
    kill_wakapi
fi

echo "Running current build ..."
start_wakapi_background "../wakapi" "$config"
kill_wakapi
rm -f wakapi_testing.db

# Only sqlite has data
if [ "$DB_TYPE" == "sqlite" ]; then
    echo "Creating database and schema ..."
    sqlite3 wakapi_testing.db < schema.sql
    echo "Importing seed data ..."
    sqlite3 wakapi_testing.db < data.sql

    start_wakapi_background "../wakapi" "$config"
    echo "Running test collection ..."
    if ! (cd "wakapi_api_tests"; bru run); then
        echo "bruno cli failed"
        exit 1
    fi

    kill_wakapi
fi
