FROM golang:latest as build
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
	-o main .


FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=build /app/main .
EXPOSE 8888
CMD ["./main"]
