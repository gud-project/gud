FROM golang

WORKDIR /go/src/gitlab.com/magsh-2019/2/gud
COPY gud gud
COPY server server
COPY Makefile .

RUN make server
RUN mv server/server /usr/local/bin

FROM scratch
COPY --from=0 /usr/local/bin/server /usr/local/bin/server

EXPOSE 8080

CMD ["/usr/local/bin/server"]
