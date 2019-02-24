package oneshot

import (
    "io"
    "os/exec"

    "github.com/pkg/errors"
    "github.com/docker/cli/cli/compose/types"
    "github.com/go-yaml/yaml"
)

func (oneshot *OneShot) StartOneShotStack(cmd string, stack string) (error) {
    config, err := loadComposefile(oneshot.Config.ComposeFiles)
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
    service.Deploy.Labels[oneshot.Config.SelectorLabel] = "true"
    return nil
}
