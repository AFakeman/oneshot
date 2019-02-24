package oneshot

import (
    "context"
    "log"
    "os"
    "strings"

    "github.com/docker/docker/client"
    "gopkg.in/robfig/cron.v2"

    "github.com/afakeman/oneshot/oneshot/cronparse"
)

type OneShot struct {
    Config *OneShotConfig
    Client *client.Client
    Cron *cron.Cron
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

    oneshot := OneShot{Config: config, Client: cli, Cron: cron.New()}
    commands, err := cronparse.ReadCrontab(oneshot.Config.CrontabFile)

    oneshot.Cron.AddFunc("* * * * *", func() {
        log.Printf("Starting cleanup...\n")
        err := oneshot.CleanUpSwarmJobs(context.Background())
        if err != nil {
            log.Printf("Could not execute cleanup: %s\n", err)
        } else {
            log.Printf("Cleanup complete")
        }
    })
    for _, command := range(commands) {
        oneshot.Cron.AddFunc(command.Timespec, func() {
            err := oneshot.StartOneShotStack(command.Command)
            if err != nil {
                log.Printf("Could not deploy a stack: %s\n", err)
            } else {
                log.Printf("Started a job for command '%s'", command)
            }
        })
    }

    return &oneshot, nil
}

func (oneshot *OneShot) Start() {
    oneshot.Cron.Start()
}

func (oneshot *OneShot) Stop() {
    oneshot.Cron.Stop()
}
