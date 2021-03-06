package oneshot

import (
    "context"
    "log"
    "os/exec"

    "github.com/docker/docker/api/types"
    "github.com/docker/docker/api/types/filters"
    "github.com/docker/docker/api/types/swarm"
)

func (oneshot *OneShot) CleanUpSwarmJobs(ctx context.Context) error {
    filter := oneshot.filterForCleanup()
    services, err := oneshot.Client.ServiceList(ctx,
        types.ServiceListOptions{Filters: filter})
    if err != nil {
        panic(err)
    }
    for _, service := range services {
        shouldDelete, err := oneshot.serviceIsNotFinished(ctx, service)
        if err != nil {
            panic(err)
        }
        if shouldDelete {
            err = oneshot.deleteStoppedStack(ctx,
                service.Spec.Labels["com.docker.stack.namespace"])
            if err != nil {
                return err
            }
        }
    }
    return nil
}

func (oneshot *OneShot) filterForCleanup() filters.Args {
    return filters.NewArgs(
        filters.KeyValuePair{"label", oneshot.Config.SelectorLabel})
}

func (oneshot *OneShot) serviceIsNotFinished(ctx context.Context,
                                        service swarm.Service) (bool, error) {
    filter := filters.NewArgs(filters.KeyValuePair{"service", service.ID})
    tasks, err := oneshot.Client.TaskList(
        ctx, types.TaskListOptions{Filters: filter})
    if err != nil {
        return false, err
    }
    foundRunning := false
    for _, task := range tasks {
        switch task.Status.State {
        case "failed":
            log.Printf("Task %s failed\n", task.ID)
        case "complete":
        default:
            foundRunning = true
        }
    }
    return !foundRunning, nil
}

func (*OneShot) deleteStoppedStack(ctx context.Context, stack string) (error) {
    log.Printf("Cleaning up stack %s\n", stack)
    cmd := exec.Command("docker", "stack", "rm", stack)
    err := cmd.Run()
    if err != nil {
        return err
    }
    return nil
}
