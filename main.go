package main

import (
    "log"
    "os"
    "os/signal"
    "syscall"

    "github.com/afakeman/oneshot/oneshot"
)


func main() {
    oneshot, err := oneshot.NewOneShot();
    if err != nil {
        panic(err)
    }

    sigs := make(chan os.Signal, 1)
    done := make(chan bool, 1)

    signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

    go func() {
        sig := <-sigs
        log.Printf("Received %s, shutting down...\n", sig)
        done <- true
    }()

    oneshot.Start()
    <-done
    oneshot.Stop()
}
