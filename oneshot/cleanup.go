package oneshot

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/docker/docker/api/types"
    "github.com/docker/docker/api/types/filters"
    "github.com/docker/docker/api/types/swarm"
)

func (oneshot *OneShot) CleanUpSwarmJobs(ctx context.Context,
                                         filter filters.Args) error {
    services, err := oneshot.Client.ServiceList(ctx,
        types.ServiceListOptions{Filters: filter})
	if err != nil {
		panic(err)
	}
    for _, service := range services {
        shouldDelete, err := oneshot.ShouldDeleteService(ctx, service)
        if err != nil {
            panic(err)
        }
        if shouldDelete {
            oneshot.DeleteStoppedStack(ctx,
                service.Spec.Labels["com.docker.stack.namespace"])
        }
    }
    return nil
}

func (oneshot *OneShot) ShouldDeleteService(ctx context.Context,
                                        service swarm.Service) (bool, error) {
    filter := filters.NewArgs(filters.KeyValuePair{"service", service.ID})
    tasks, err := oneshot.Client.TaskList(
        ctx, types.TaskListOptions{Filters: filter})
	if err != nil {
		panic(err)
	}
    foundRunning := false
    for _, task := range tasks {
        switch task.Status.State {
        case "failed":
            fmt.Printf("%s failed\n", task.ID)
        case "finished":
            fmt.Printf("%s finished\n", task.ID)
        case "running":
            fmt.Printf("%s running\n", task.ID)
            foundRunning = true
        default:
            fmt.Printf("Unexpected state for %s: %s\n", task.ID,
                task.Status.State)
        }
    }
    return !foundRunning, nil
}

func (*OneShot) DeleteStoppedStack(ctx context.Context, stack string) {
    fmt.Printf("Got asked to delete stack %s\n", stack)
    cmd := exec.Command("docker", "stack", "rm", stack)
    err := cmd.Run()
    if err != nil {
        panic(err)
    }
}
