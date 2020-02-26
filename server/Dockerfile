FROM golang:1.13

WORKDIR /go/src/gitlab.com/magsh-2019/2/gud
COPY gud gud
COPY server server
COPY Makefile .

RUN CGO_ENABLED=0 make server
RUN mv server/server /

FROM scratch
COPY --from=0 /server /var/www/server

VOLUME /var/www/projects
EXPOSE 8080

CMD ["/var/www/server"]
