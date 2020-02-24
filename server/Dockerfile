FROM golang:1.13 AS back
WORKDIR /go/src/gitlab.com/magsh-2019/2/gud
COPY gud/ gud/
COPY server server
COPY Makefile .

RUN CGO_ENABLED=0 make back
RUN mv server/server /

FROM node AS front
COPY server/front/ server/front/
COPY Makefile .

RUN make front

FROM scratch
COPY --from=back /server /var/www/server
COPY --from=front /server/front/dist /var/www/front/dist

VOLUME /var/www/projects
EXPOSE 8080

CMD ["/var/www/server"]
