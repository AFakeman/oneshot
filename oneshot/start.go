package oneshot

import (
    "context"
    "fmt"
    "io"
    "os/exec"
    "regexp"
    "strings"
    "time"

    "github.com/pkg/errors"
    "github.com/docker/cli/cli/compose/types"
    dockertypes "github.com/docker/docker/api/types"
    "github.com/docker/docker/api/types/filters"
    "github.com/go-yaml/yaml"
)

func (oneshot *OneShot) StartOneShotStack(cmd string) (error) {
    env := map[string]string{
        "COMMAND": cmd,
    }
    config, err := loadComposefile(oneshot.Config.ComposeFiles, env)
    if err != nil {
        return err
    }

    err = oneshot.modifyStackConfig(config, cmd)
    if err != nil {
        return errors.Wrap(err, "Couldn't modify config")
    }

    new_config, err := yaml.Marshal(config)
    if err != nil {
        return errors.Wrap(err, "Couldn't dump modified config")
    }

    stack := oneshot.stackNameForCommand(cmd, time.Now())

    duplicate_running, err := oneshot.duplicateAlreadyRunning(cmd)
    if err != nil {
        return err
    }
    if duplicate_running {
        return errors.New("Command already running")
    }

    command := exec.Command("docker", "stack", "deploy", "-c", "-", stack)

    stdin, err := command.StdinPipe()
    if err != nil {
        return errors.Wrap(err, "Couldn't get subprocess stdin")
    }

    go func() {
        defer stdin.Close()
        io.WriteString(stdin, string(new_config))
    }()

    out, err := command.CombinedOutput()
    if err != nil {
        return errors.Errorf("docker stack deploy raised %s. Output:\n%s", err, out)
    }

    return nil
}

var notAllowedCharactersRegex, _ = regexp.Compile("[^a-zA-Z0-9\\.-_]")

func (oneshot *OneShot) stackNameForCommand(cmd string, t time.Time) (string) {
    cmd_san := strings.Replace(cmd, " ", "_", -1)
    cmd_san = strings.Replace(cmd_san, "/", ".", -1)
    cmd_san = notAllowedCharactersRegex.ReplaceAllString(cmd_san, "")
    time_part := t.Format("2006-02-01T15-04-05")
    name := fmt.Sprintf("cron_%s_%s", cmd_san, time_part)
    return name
}

func (oneshot *OneShot) modifyStackConfig(config *types.Config, cmd string) (error) {
    if len(config.Services) != 1 {
        return errors.New("There should only be one service in a stack")
    }

    if config.Services[0].Deploy.RestartPolicy == nil {
        config.Services[0].Deploy.RestartPolicy = &types.RestartPolicy{}
    }

    if config.Services[0].Deploy.RestartPolicy.Condition != "none" {
        return errors.New("RestartPolicy should be 'none'")
    }

    service := &config.Services[0]

    service.Command = []string{"sh", "-c", cmd}
    if service.Deploy.Labels == nil {
        service.Deploy.Labels = types.Labels{}
    }
    service.Deploy.Labels[oneshot.Config.SelectorLabel] = cmd
    return nil
}

func (oneshot *OneShot) duplicateAlreadyRunning(cmd string) (bool, error) {
    label := fmt.Sprintf("%s=%s", oneshot.Config.SelectorLabel, cmd)
    filter := filters.NewArgs(
        filters.KeyValuePair{"label", label})
    services, err := oneshot.Client.ServiceList(context.Background(),
        dockertypes.ServiceListOptions{Filters: filter})
    if err != nil {
        return false, errors.Wrap(err, "Could not list services")
    }
    for _, service := range services {
        finished, err := oneshot.serviceIsNotFinished(context.Background(), service)
        if err != nil {
            return false, err
        }
        if !finished {
            return true, nil
        }
    }
    return false, nil
}
