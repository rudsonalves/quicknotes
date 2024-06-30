# Etapa de construção
FROM golang:1.22.4-alpine
WORKDIR /app
RUN apk update && apk upgrade && apk add --no-cache ca-certificates
RUN update-ca-certificates
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -v -o ./server ./cmd/http/

# Etapa final
FROM scratch
WORKDIR /bin
COPY --from=0 /app/server server
COPY --from=0 /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY .env /bin/.env
COPY cer.cer /bin/cer.cer
COPY cer.key /bin/cer.key
CMD ["/bin/server"]