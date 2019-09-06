#/bin/bash
#set -o errexit -o pipefail -o nounset

cd $(dirname "$BASH_SOURCE")

# relies on entry for VM public ip in /etc/hosts
export DOCKER_HOST="tcp://docker.local:2376"
export DOCKER_TLS_VERIFY="1"
export DOCKER_CERT_PATH="$PWD/server" #certs are in server folder
