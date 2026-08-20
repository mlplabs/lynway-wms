FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git ca-certificates

ARG PROJECT_NAME
ARG APP_VERSION

WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build \
      -o /app-bin \
      -ldflags "-X 'main.AppVersion=${APP_VERSION}'" ./cmd/${PROJECT_NAME}


FROM alpine:3.20
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app-bin ./app

EXPOSE 8080
CMD ["./app"]
