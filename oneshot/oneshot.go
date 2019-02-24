package oneshot

import (
    "os"
    "strings"

    "github.com/docker/docker/client"
)

type OneShot struct {
    Config *OneShotConfig
    Client *client.Client
}

type OneShotConfig struct {
    ComposeFiles []string
    CrontabFile string
    SelectorLabel string
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

    value, found = os.LookupEnv("ONESHOT_SELECTOR_LABEL");
    if (!found) {
        value = "wd.cron.job"
    }
    config.SelectorLabel = value;

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

