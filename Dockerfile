FROM golang:1.26-alpine AS builder

ARG SERVICE_NAME

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /server ./services/${SERVICE_NAME}/cmd/server

FROM alpine:3.20

ARG SERVICE_NAME

RUN apk --no-cache add ca-certificates

COPY --from=builder /server /server
COPY --from=builder /app/services/${SERVICE_NAME}/migrations /migrations

EXPOSE 8080

CMD ["/server"]
