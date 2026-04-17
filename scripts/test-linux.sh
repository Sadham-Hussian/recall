#!/usr/bin/env bash
set -e

docker run -it --rm --platform linux/amd64 \
  -v "$PWD":/workspace -w /workspace \
  --name recall-test \
  ubuntu:22.04 bash -c '
    set -e
    apt-get update -qq
    apt-get install -y -qq curl gcc make bash-completion >/dev/null

    curl -fsSL https://go.dev/dl/go1.24.1.linux-amd64.tar.gz | tar -C /usr/local -xz
    export PATH=/usr/local/go/bin:$PATH

    echo "--- go version ---"
    go version

    echo "--- building recall ---"
    CGO_ENABLED=1 make build
    mv recall /usr/local/bin/

    echo "--- recall version ---"
    recall version

    echo "--- dropping into interactive bash ---"
    exec bash
  '
