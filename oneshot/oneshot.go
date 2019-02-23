package oneshot

import (
	"context"
	"fmt"
    "os"
	"os/exec"
    "strings"
    //"github.com/robfig/cron"

	"github.com/docker/docker/api/types"
    "github.com/docker/docker/api/types/filters"
    "github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/client"
)

type OneShot struct {
    Config *OneShotConfig
    Client *client.Client
}

type OneShotConfig struct {
    ComposeFiles []string
    CrontabFile string
}

type OneShotError struct {
    s string;
}

func (err *OneShotError) Error() string {
    return err.s;
}

func NewOneShotConfig() (*OneShotConfig, error) {
    config := OneShotConfig{};
    value, found := os.LookupEnv("ONESHOT_COMPOSE_FILE");
    if (!found) {
        return nil, &OneShotError{
            "No environment variable ONESHOT_COMPOSE_FILE found",
        }
    }
    config.ComposeFiles = strings.Split(value, ":")
    value, found = os.LookupEnv("ONESHOT_CRONTAB_FILE");
    if (!found) {
        value = "/etc/oneshot/crontab"
    }
    config.CrontabFile = value;
    return &config, nil;
}

func NewOneShot() (*OneShot, error) {
	cli, err := client.NewClientWithOpts(
        client.WithVersion("1.39"), client.FromEnv)
	config, err := NewOneShotConfig();
	if err != nil {
		return nil, err
	}
    oneshot := OneShot{Config: config, Client: cli}
    return &oneshot, nil
}

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
