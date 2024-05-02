package task

import (
	"context"
	"io"
	"log"
	"os"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
)

type State string

const (
	Pending   State = "Pending"
	Running   State = "Running"
	Completed State = "Completed"
	Failed    State = "Failed"
)

type Task struct {
	Spec   TaskSpec
	Status TaskStatus
}

type TaskSpec struct {
	Image string
	Cmd   []string
}

type TaskStatus struct {
	ContainerId string
	State       State
}

func (t *Task) Start() {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Printf("Error creating docker client: %v", err)
		t.Status.State = Failed
		return
	}
	ctx := context.Background()
	reader, err := cli.ImagePull(ctx, t.Spec.Image, image.PullOptions{})
	if err != nil {
		log.Printf("Error pulling image: %v", err)
		t.Status.State = Failed
		return
	}
	defer reader.Close()
	io.Copy(os.Stdout, reader)

	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: t.Spec.Image,
		Cmd:   t.Spec.Cmd,
	}, nil, nil, nil, "")
	if err != nil {
		log.Printf("Error creating container: %v", err)
		t.Status.State = Failed
		return
	}

	if err := cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		log.Printf("Error starting container: %v", err)
		t.Status.State = Failed
		return
	}

	t.Status.State = Running
	t.Status.ContainerId = resp.ID
}

func (t *Task) Stop() {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())

	if err != nil {
		log.Printf("Error creating docker client: %v", err)
		return
	}

	ctx := context.Background()

	if err := cli.ContainerStop(ctx, t.Status.ContainerId, container.StopOptions{}); err != nil {
		log.Printf("Error stopping container: %v", err)
		return
	}

	t.Status.State = Completed
}
