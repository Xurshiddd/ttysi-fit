# Dev uchun: Windows Application Control imzosiz exe'ni bloklasa,
# backend Docker (Linux) ichida ishlaydi — Windows exe umuman kerak emas.
FROM golang:1.25-alpine

WORKDIR /app

# Bog'liqliklar (cache qatlami)
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /bin/api ./cmd/api

EXPOSE 8090
CMD ["/bin/api"]
