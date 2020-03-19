package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
)

type imageBuildData struct {
	Stream string              `json:"stream"`
	Aux    struct{ ID string } `json:"aux"`
}

func execJob(tar io.Reader) (int, []byte, error) {
	cli, err := client.NewClientWithOpts()
	if err != nil {
		return -1, nil, err
	}

	ctx := context.Background()
	var stdout bytes.Buffer

	id, err := buildImage(cli, ctx, tar, &stdout)
	if err != nil {
		return -1, stdout.Bytes(), err
	}

	status, err := runContainer(cli, ctx, id, &stdout)
	return status, stdout.Bytes(), err
}

func runContainer(cli *client.Client, ctx context.Context, image string, stdout io.Writer) (status int, err error) {
	cont, err := cli.ContainerCreate(ctx, &container.Config{
		Image: image,
	}, nil, nil, "")
	if err != nil {
		return -1, err
	}
	for _, warning := range cont.Warnings {
		log.Println("docker warning: ", warning)
	}

	id := cont.ID
	defer func() {
		closeErr := cli.ContainerRemove(ctx, id, types.ContainerRemoveOptions{Force: err != nil})
		if err == nil {
			err = closeErr
		}
	}()

	err = cli.ContainerStart(ctx, id, types.ContainerStartOptions{})
	if err != nil {
		return -1, err
	}

	logs, err := cli.ContainerLogs(ctx, id, types.ContainerLogsOptions{
		Follow:     true,
		ShowStdout: true,
		ShowStderr: true,
	})
	if err != nil {
		return -1, err
	}
	defer logs.Close()

	_, err = stdcopy.StdCopy(stdout, stdout, logs)
	if err != nil {
		return -1, err
	}

	stat, err := cli.ContainerInspect(ctx, id)
	if err != nil {
		return -1, err
	}

	return stat.State.ExitCode, nil
}

func buildImage(cli *client.Client, ctx context.Context, tar io.Reader, out io.StringWriter) (string, error) {
	res, err := cli.ImageBuild(ctx, tar, types.ImageBuildOptions{
		ForceRemove: true,
		// NoCache?
	})
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	dec := json.NewDecoder(res.Body)
	var id string
	for {
		var data imageBuildData
		err = dec.Decode(&data)
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}

		if data.Stream != "" {
			_, err = out.WriteString(data.Stream)
			if err != nil {
				return "", err
			}
		} else if data.Aux.ID != "" {
			id = data.Aux.ID
		}
	}
	return id, nil
}
