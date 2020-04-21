FROM golang:alpine AS back
WORKDIR /go/src/
RUN apk add -u --no-cache upx

COPY server/go.mod server/go.sum server/
COPY gud/go.mod gud/go.sum gud/
RUN cd server/ && go mod download

COPY gud/*.go gud/
COPY server/*.go server/
RUN cd server/ && CGO_ENABLED=0 go build && upx server


FROM golang:alpine AS cli
WORKDIR /go/src/

COPY go.mod go.sum ./
COPY gud/go.mod gud/go.sum gud/
RUN go mod download

COPY gud/*.go gud/
COPY cmd/*.go cmd/
COPY main.go ./
RUN mkdir release && \
	for target in 'windows' 'darwin' 'linux'; do \
		GOOS="$target" GOARCH=amd64 go build \
			-ldflags '-s -w' -o "/$target/gud$([ $target = 'windows' ] && echo -z '.exe')"; \
	done


FROM node:alpine AS front
COPY server/front/package.json server/front/package-lock.json ./
RUN npm install

COPY server/front/ ./
RUN npm run build

FROM scratch
WORKDIR /var/www/
COPY --from=back /go/src/server/server server
COPY --from=front /dist/ front/dist/
COPY --from=cli /go/src/release front/dist/release/

VOLUME projects/
EXPOSE 8080

CMD ["./server"]
