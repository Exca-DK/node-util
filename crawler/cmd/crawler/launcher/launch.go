package launcher

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"

	application "github.com/Exca-DK/node-util/crawler/app"

	"github.com/Exca-DK/node-util/crawler/log"
	"github.com/urfave/cli/v2"
)

var (
	app = &cli.App{
		Name:        filepath.Base(os.Args[0]),
		Usage:       "crawl p2p network for specific node types",
		Version:     application.VersionWithMeta(),
		Writer:      os.Stdout,
		HideVersion: false,
	}
)

func init() {
	app.Before = func(ctx *cli.Context) error {
		log.NewLogger()
		var cancel func()
		ctx.Context, cancel = context.WithCancel(ctx.Context)
		ch := make(chan os.Signal, 10)
		signal.Notify(ch, os.Interrupt)
		var canceled bool
		go func() {
			for i := 0; i < 10; i++ {
				<-ch
				if !canceled {
					cancel()
					canceled = true
				}
				fmt.Printf("%v gracefull attempts left before forceed crash\n", 10-i)
			}
			os.Exit(1)
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
		CrawlerCommand,
	}
}

func Launch(args []string) error {
	return app.Run(args)
}
