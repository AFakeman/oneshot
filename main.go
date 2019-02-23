package main

import (
	"context"

    "github.com/docker/docker/api/types/filters"

    "github.com/afakeman/oneshot/oneshot"
)


func main() {
    oneshot, err := oneshot.NewOneShot();
	if err != nil {
		panic(err)
	}

    filter := filters.NewArgs(filters.KeyValuePair{"label", "wd.cron.job"})

	err = oneshot.CleanUpSwarmJobs(context.Background(), filter)
	if err != nil {
		panic(err)
	}
}