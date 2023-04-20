package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"

	application "github.com/Exca-DK/node-util/crawler/app"

	v4 "github.com/Exca-DK/node-util/crawler/cmd/disc/v4"
	"github.com/urfave/cli/v2"
)

var (
	app = &cli.App{
		Name:        filepath.Base(os.Args[0]),
		Usage:       "crypto discovery tool",
		Version:     application.VersionWithMeta(),
		Writer:      os.Stdout,
		HideVersion: false,
	}
)

func init() {
	app.Before = func(ctx *cli.Context) error {
		var cancel func()
		ctx.Context, cancel = context.WithCancel(ctx.Context)
		go func() {
			ch := make(chan os.Signal)
			signal.Notify(ch, os.Interrupt)
			<-ch
			cancel()
		}()
		return nil
	}
	app.After = func(ctx *cli.Context) error {
		fmt.Sprintln("After")
		return nil
	}
	app.CommandNotFound = func(ctx *cli.Context, cmd string) {
		fmt.Fprintf(os.Stderr, "No such command: %s\n", cmd)
		os.Exit(1)
	}
	// Add subcommands.
	app.Commands = []*cli.Command{
		v4.Discv4Command,
	}
}

func main() {
	exit(app.Run(os.Args))
}

func exit(err interface{}) {
	if err == nil {
		os.Exit(0)
	}
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
