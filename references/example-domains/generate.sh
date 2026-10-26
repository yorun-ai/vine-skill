#!/usr/bin/env bash
set -euo pipefail
cd -- "$(dirname -- "${BASH_SOURCE[0]}")"
export GOWORK=off

skelc check --skel-in ./skel/catalog
skelc check --skel-in ./skel/orders

# The public Skel view is the input another domain resolves.
skelc gen skel --pub --skel-in ./skel/catalog --skel-out ./skeled/skel/catalog
skelc gen go-module --skel-in ./skel/catalog \
  --go-out ./skeled/golang/catalog --go-module example.com/shop/skeled/catalog \
  --go-pub-out ./skeled/golang/catalogpub --go-pub-module example.com/shop/skeled/catalogpub \
  --go-vine-version v0.32.5

skelc gen skel --pub --skel-in ./skel/orders --skel-out ./skeled/skel/orders \
  --skel-import shop.catalog=./skeled/skel/catalog
skelc gen go-module --skel-in ./skel/orders \
  --skel-import shop.catalog=./skeled/skel/catalog \
  --go-import shop.catalog=example.com/shop/skeled/catalogpub@v0.0.0 \
  --go-out ./skeled/golang/orders --go-module example.com/shop/skeled/orders \
  --go-pub-out ./skeled/golang/orderspub --go-pub-module example.com/shop/skeled/orderspub \
  --go-vine-version v0.32.5

# Go resolves local modules from the module currently being built. Dependency
# modules' replace directives do not propagate to the server module.
for target in catalogpub catalog orderspub orders; do
  (
    cd "skeled/golang/$target"
    if [[ "$target" != catalogpub ]]; then
      go mod edit -replace=example.com/shop/skeled/catalogpub=../catalogpub
    fi
    if [[ "$target" == orders ]]; then
      go mod edit -replace=example.com/shop/skeled/orderspub=../orderspub
    fi
    go mod tidy
  )
done

(
  cd src/server
  if [[ ! -f go.mod ]]; then
    go mod init example.com/shop/src/server
  fi
  go mod edit -go=1.27.0
  for dependency in catalogpub catalog orderspub orders; do
    go mod edit "-require=example.com/shop/skeled/$dependency@v0.0.0" \
      "-replace=example.com/shop/skeled/$dependency=../../skeled/golang/$dependency"
  done
  go mod tidy
)
