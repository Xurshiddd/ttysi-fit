#!/usr/bin/env bash
# TTYSI_FIT backend — Git Bash yordamchi skript (make o'rniga)
# Foydalanish: ./run.sh <buyruq>
set -e

# .env.local dan DB sozlamalarini o'qish
ENV_FILE=".env.local"
if [ -f "$ENV_FILE" ]; then
  export $(grep -v '^#' "$ENV_FILE" | grep -E '^(DB_|REDIS_)' | xargs)
fi

DB_STRING="host=${DB_HOST:-localhost} port=${DB_PORT:-5432} user=${DB_USER:-postgres} password=${DB_PASSWORD} dbname=${DB_NAME:-ttysi_fit_dev} sslmode=${DB_SSLMODE:-disable}"
GOOSE="$(go env GOPATH)/bin/goose"

# Windows Application Control %TEMP% dagi exe'ni bloklaydi — shu sababli
# `go run` o'rniga loyiha ichidagi bin/ ga build qilib ishga tushiramiz.
# Nom "ttysi_" prefiksi bilan — boshqa loyihalarning exe'lari bilan adashmasin.
run_cmd() { # run_cmd <nom> <paket> [arg...]
  local name="ttysi_$1" pkg="$2"; shift 2
  # Eski jarayon exe faylni band qilib turgan bo'lishi mumkin ("Permission denied").
  taskkill //F //IM "$name.exe" >/dev/null 2>&1 || true
  go build -o "bin/$name.exe" "$pkg"
  "./bin/$name.exe" "$@"
}

case "$1" in
  tidy)        go mod tidy ;;
  dev)         run_cmd api ./cmd/api ;;
  dev-docker)  docker compose --profile api up --build api ;;  # App Control exe'ni bloklasa
  sync)        run_cmd sync ./cmd/sync "${2:-all}" ;;   # ./run.sh sync [all|structures|groups|students|employees]
  create-admin) run_cmd createadmin ./cmd/createadmin "$2" "$3" "$4" ;;  # ./run.sh create-admin <email> <parol> [ism]
  build)       go build -o bin/ttysi_api.exe cmd/api/main.go ;;
  test)
    # -race Windows'da cgo (gcc) talab qiladi; gcc bo'lmasa -race'siz ishlaydi.
    if command -v gcc >/dev/null 2>&1; then
      CGO_ENABLED=1 go test ./... -v -race
    else
      echo "⚠️  gcc topilmadi — race detector'siz test (to'liq: gcc o'rnatib qayta ishga tushiring)"
      go test ./... -v
    fi
    ;;
  redis-up)    docker compose up -d redis ;;
  redis-down)  docker compose down ;;
  db-create)   psql -U "${DB_USER:-postgres}" -c "CREATE DATABASE ${DB_NAME:-ttysi_fit_dev};" ;;
  migrate)     "$GOOSE" -dir migrations postgres "$DB_STRING" up ;;
  migrate-down) "$GOOSE" -dir migrations postgres "$DB_STRING" down ;;
  migrate-fresh)
    "$GOOSE" -dir migrations postgres "$DB_STRING" reset
    "$GOOSE" -dir migrations postgres "$DB_STRING" up ;;
  goose-install) go install github.com/pressly/goose/v3/cmd/goose@latest ;;
  setup)
    go mod tidy
    docker compose up -d redis
    go install github.com/pressly/goose/v3/cmd/goose@latest
    echo "✅ Tayyor. Endi: ./run.sh db-create && ./run.sh migrate && ./run.sh dev"
    ;;
  *)
    echo "Buyruqlar: setup | tidy | dev | sync [all|structures|groups|students|employees] | build | test | redis-up | redis-down | db-create | goose-install | migrate | migrate-down | migrate-fresh"
    ;;
esac
