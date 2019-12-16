package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/micro/cli"
	"github.com/micro/go-micro/config/cmd"
	gorun "github.com/micro/go-micro/runtime"
	"github.com/micro/go-micro/util/log"
	"github.com/micro/micro/api"
)

func main() {
	app := cli.NewApp()
	app.Commands = append(app.Commands, api.Commands()...)
	app.Commands = append(app.Commands, newRuntime())
	app.Run(os.Args)
}

func newRuntime() cli.Command {

	services := []string{
		"api",
	}

	env := os.Environ()
	muRuntime := cmd.DefaultCmd.Options().Runtime

	command := cli.Command{
		Name: "runtime",
		Action: func(ctx *cli.Context) {
			for _, serv := range services {
				// TODO(refs) there should be a better way to do this rather than this.
				// this is calling itself with the services on "service" as part of the subcommands
				// it therefore depends on the binary having the right set of subcommands
				args := []gorun.CreateOption{
					gorun.WithCommand(os.Args[0], serv),
					gorun.WithEnv(env),
					gorun.WithOutput(os.Stdout),
				}

				muService := &gorun.Service{Name: serv}
				// this uses micro runtime to fork in a new proccess
				if err := (*muRuntime).Create(muService, args...); err != nil {
					log.Errorf("Failed to create runtime enviroment: %v", err)
					os.Exit(1)
				}

				shutdown := make(chan os.Signal, 1)
				signal.Notify(shutdown, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

				log.Info("Starting service runtime")

				// start the runtime
				if err := (*muRuntime).Start(); err != nil {
					log.Fatal(err)
				}

				log.Info("Service runtime started")

				select {
				case <-shutdown:
					log.Info("Shutdown signal received")
					log.Info("Stopping service runtime")
				}

				// stop all the things
				if err := (*muRuntime).Stop(); err != nil {
					log.Fatal(err)
				}

				log.Info("Service runtime shutdown")

				// exit success
				os.Exit(0)
			}
		},
	}
	return command
}
