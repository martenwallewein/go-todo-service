FROM golang:1.18-alpine
WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 go build

FROM alpine
RUN apk add --no-cache ca-certificates
COPY --from=0 /src /bin/todo-service

RUN ls /bin/todo-service
ENTRYPOINT ["/bin/todo-service/todo-service"] 