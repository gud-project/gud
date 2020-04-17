module gitlab.com/magsh-2019/2/gud/server

go 1.14

require (
	github.com/Azure/go-ansiterm v0.0.0-20170929234023-d6e3b3328b78 // indirect
	github.com/Microsoft/go-winio v0.4.14 // indirect
	github.com/containerd/containerd v1.3.4 // indirect
	github.com/docker/distribution v2.7.1+incompatible // indirect
	github.com/docker/docker v17.12.0-ce-rc1.0.20200309214505-aa6a9891b09c+incompatible
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/docker/go-units v0.4.0 // indirect
	github.com/gogo/protobuf v1.3.1 // indirect
	github.com/gorilla/mux v1.7.4
	github.com/gorilla/sessions v1.2.0
	github.com/lib/pq v1.3.0
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/opencontainers/go-digest v1.0.0-rc1 // indirect
	github.com/opencontainers/image-spec v1.0.1 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/sirupsen/logrus v1.5.0 // indirect
	gitlab.com/magsh-2019/2/gud/gud v0.0.0
	golang.org/x/crypto v0.0.0-20190308221718-c2843e01d9a2
	golang.org/x/time v0.0.0-20200416051211-89c76fbcd5d1 // indirect
	google.golang.org/grpc v1.28.1 // indirect
	gotest.tools v2.2.0+incompatible // indirect
)

replace gitlab.com/magsh-2019/2/gud/gud => ../gud
