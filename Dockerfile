FROM golang:1.10 AS builder
WORKDIR /go/src/github.com/matthewlujp/audiube/
COPY ./src ./src
COPY ./Gopkg.lock .
COPY ./Gopkg.toml .
RUN go get -u github.com/golang/dep/cmd/dep
RUN dep ensure --vendor-only
RUN cd src && GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -a -installsuffix cgo -o ../main .

FROM matthewishige0528/golang-ffmpeg:v1.0 AS app_runner
ARG api_key
ENV API_KEY $api_key
RUN apk add --no-cache openssh
WORKDIR /app
COPY ./static ./static
COPY ./run.sh ./run.sh
COPY --from=builder /go/src/github.com/matthewlujp/audiube/main .
ENTRYPOINT ["./run.sh"]