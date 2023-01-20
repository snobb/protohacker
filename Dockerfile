FROM golang:1.19-alpine as build-image
WORKDIR /build
ENV CGO_ENABLED=0
COPY go.* ./
COPY ./cmd ./cmd
COPY ./pkg ./pkg
RUN go build --ldflags "-s" -o server ./cmd/main.go

FROM scratch
WORKDIR /project
COPY --from=build-image ./build/server ./
EXPOSE 8080 5000/udp
CMD [ "./server" ]
