FROM golang:alpine AS back
COPY server/go.mod server/go.sum server/
COPY gud/go.mod gud/go.sum gud/
RUN cd server/ && go mod download

COPY gud/*.go gud/
COPY server/*.go server/
RUN cd server/ && CGO_ENABLED=0 go build

FROM node:alpine AS front
COPY server/front/ .
RUN npm run build -- --mode development

FROM scratch
COPY --from=back /go/server /var/www/server
COPY --from=front /dist /var/www/front/dist

VOLUME /var/www/projects
EXPOSE 8080

CMD ["/var/www/server"]
