#!/bin/bash
set -o nounset -o errexit -o pipefail

cleanup() {
    if [ "${WAKAPI_TEST_EXTERNAL_SMTP4DEV:-0}" -eq 1 ]; then
        return
    fi
    echo "Stopping and removing existing smtp4dev instances ..."
    docker stop smtp4dev_wakapi &> /dev/null || true
    docker rm -f smtp4dev_wakapi &> /dev/null || true
}
trap cleanup EXIT

cleanup

if [ "${WAKAPI_TEST_EXTERNAL_SMTP4DEV:-0}" -eq 1 ]; then
    smtp4dev_host=${WAKAPI_TEST_SMTP4DEV_HOST:-localhost}
    smtp4dev_port=${WAKAPI_TEST_SMTP4DEV_PORT:-2525}
    echo "Using external smtp4dev at $smtp4dev_host:$smtp4dev_port ..."
    smtp4dev_ready=0
    for _ in $(seq 1 60); do
        if bash -c "</dev/tcp/$smtp4dev_host/$smtp4dev_port" 2> /dev/null; then
            smtp4dev_ready=1
            break
        fi
        sleep 1
    done
    if [ "$smtp4dev_ready" -ne 1 ]; then
        echo "Timed out waiting for smtp4dev at $smtp4dev_host:$smtp4dev_port"
        exit 1
    fi
else
    echo "Starting smtp4dev in Docker ..."
    docker run -d --rm -p 2525:25 -p 8080:80 --name smtp4dev_wakapi rnwood/smtp4dev
fi

echo "Running tests ..."
script_dir=$(dirname "${BASH_SOURCE[0]}")
go test -count=1 -run TestSmtpTestSuite "$script_dir/../services/mail"
