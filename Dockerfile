FROM golang:alpine as builder

RUN apk update && \
    apk add git && \
    adduser -D -g "" appuser

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags "\
	-w -s \
	-X main.gitRepo=$(git remote get-url --push origin) \
	-X main.gitCommit=$(git rev-list -1 HEAD) \
	-X main.appVersion=$(git tag -l | tail -n1) \
	-X main.buildStamp=$(date +%Y-%m-%d)" \
	-o fakeserver .


FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY ./docker-entrypoint.sh /app/docker-entrypoint.sh
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /app/fakeserver .
EXPOSE 8888
ENV ADDR :8888
ENV TTL 120
USER appuser

ENTRYPOINT ["/app/docker-entrypoint.sh"]
