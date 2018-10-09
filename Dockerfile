FROM golang:1.10.2 AS build
COPY . /go/src/app

WORKDIR /go/src/app
RUN CGO_ENABLED=0 go build -a -ldflags '-s' -installsuffix cgo -o app .

# copy the binary from the build stage to the final stage
FROM alpine:3.7
COPY --from=build /go/src/app/app /job-blocker
ENTRYPOINT ["/job-blocker"]