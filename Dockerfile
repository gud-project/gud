FROM golang:alpine

WORKDIR /go/src/gud/
COPY go.mod go.sum ./
COPY gud/go.mod gud/go.sum gud/
RUN go mod download

COPY gud/*.go gud/
COPY cmd/*.go cmd/
COPY main.go ./
RUN go install

FROM bats/bats
RUN apk add -u --no-cache grep
COPY --from=0 /go/bin/gud /usr/local/bin/gud
COPY test.bats ./

CMD ["test.bats"]
