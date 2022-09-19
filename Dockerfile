FROM golang:1.19-alpine as build-image
WORKDIR /build
COPY go.* ./
COPY ./cmd ./cmd
COPY ./pkg ./pkg
RUN go build --ldflags "-s" -o server ./cmd/main.go

FROM alpine:latest
WORKDIR /project
COPY --from=build-image ./build/server ./
EXPOSE 8080 5000/udp
CMD [ "./server" ]
