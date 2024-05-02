package task

import (
	"context"
	"testing"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

func TestStartStopTask(t *testing.T) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		t.Fatal(err)
	}
	defer cli.Close()

	task := Task{Spec: TaskSpec{Image: "alpine:latest", Cmd: []string{"sleep", "5"}}, Status: TaskStatus{State: Pending}}
	task.Start()

	assertTaskState(t, Running, task.Status.State)
	assertContainerState(t, cli, task.Status.ContainerId, "running")

	task.Stop()
	time.Sleep(6 * time.Second)
	assertTaskState(t, Completed, task.Status.State)
	assertContainerState(t, cli, task.Status.ContainerId, "exited")
}

func assertContainerState(t *testing.T, cli *client.Client, containerId string, expected string) {
	t.Helper()
	containers, err := cli.ContainerList(context.Background(), container.ListOptions{All: true})
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range containers {
		if c.ID == containerId {
			if c.State != expected {
				t.Errorf("Expected container with id %s to have status '%s', got '%s'", containerId, expected, c.State)
			}
			return
		}
	}
	t.Errorf("Expected container with id %s to exist", containerId)
}

func assertTaskState(t *testing.T, expected State, got State) {
	t.Helper()
	if expected != got {
		t.Errorf("Expected task state to be %v, got %v", expected, got)
	}
}
