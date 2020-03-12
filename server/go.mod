module gitlab.com/magsh-2019/2/gud/server

go 1.13

require (
	github.com/containerd/containerd v1.3.3 // indirect
	github.com/docker/distribution v2.7.1+incompatible // indirect
	github.com/docker/docker v1.4.2-0.20200229013735-71373c6105e3
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/docker/go-units v0.4.0 // indirect
	github.com/gogo/protobuf v1.3.1 // indirect
	github.com/gorilla/mux v1.7.3
	github.com/gorilla/sessions v1.2.0
	github.com/lib/pq v1.3.0
	github.com/opencontainers/go-digest v1.0.0-rc1 // indirect
	github.com/opencontainers/image-spec v1.0.1 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/sirupsen/logrus v1.4.2 // indirect
	gitlab.com/magsh-2019/2/gud/gud v0.0.0
	golang.org/x/crypto v0.0.0-20200109152110-61a87790db17
	google.golang.org/grpc v1.28.0 // indirect
)

replace gitlab.com/magsh-2019/2/gud/gud => ../gud
