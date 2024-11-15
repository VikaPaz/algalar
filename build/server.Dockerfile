FROM golang:1.22-alpine as builder
WORKDIR /src
COPY cmd/main.go ./main.go
COPY internal ./internal
COPY docs ./docs
COPY migrations ./migrations
COPY ["go.mod", "go.sum", "./"]
RUN go mod download
RUN go build -o bin/server ./main.go

FROM alpine:edge
COPY --from=builder src/bin/server bin/server
CMD ["/bin/server"]