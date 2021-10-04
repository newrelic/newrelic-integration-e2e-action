ARG MODULE=validator
############################
# STEP 1 build binary
############################
FROM golang:alpine AS builder
RUN apk update && apk add --no-cache git
WORKDIR $GOPATH/src
COPY . .
RUN go build -o /go/bin/validator cmd/validator/main.go
RUN go build -o /go/bin/executor cmd/executor/main.go

############################
# STEP 2 build a small image
############################
FROM scratch
# Copy our static executable.
COPY --from=builder /go/bin/validator /validator
COPY --from=builder /go/bin/executor /executor
