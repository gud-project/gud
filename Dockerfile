FROM golang

# testing framework
WORKDIR /bats
RUN git clone --depth=1 https://github.com/bats-core/bats-core.git && \
	cd bats-core && \
	./install.sh /usr/local

WORKDIR /go/src/gitlab.com/magsh-2019/2/gud
COPY . .
RUN make cli

CMD ["/usr/local/bin/bats", "test.bats"]
