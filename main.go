package main

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/watchly/ngbuild/core"
	"github.com/watchly/ngbuild/integrations/github"
	"github.com/watchly/ngbuild/integrations/slack"
	"github.com/watchly/ngbuild/integrations/web"
)

func main() {
	fmt.Println(",.-~*´¨¯¨`*·~-.¸-(_NGBuild_)-,.-~*´¨¯¨`*·~-.¸")
	fmt.Println("   Building your dreams, one step at a time\n")

	httpDone := core.StartHTTPServer()

	integrations := []core.Integration{
		web.NewWeb(),
		github.New(),
		slack.NewSlack(),
	}
	core.SetIntegrations(integrations)

	fmt.Println("Available Integrations:")
	for _, integration := range core.GetIntegrations() {
		fmt.Printf("    %s\n", integration.Identifier())
	}

	apps := core.GetApps()
	if len(apps) < 1 {
		fmt.Println(`You have no configured apps, or we can't find your apps directory
To create an app, create an apps/ directory in your ngbuild directory and create subdirectories per app`)
		os.Exit(1)
	}

	fmt.Println("Apps:")
	for _, app := range apps {
		fmt.Printf("    %s\n", app.Name())
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Kill, os.Interrupt)

	select {
	case <-signals:
	case <-httpDone:
	}

	fmt.Println("Thank you for choosing ngbuild, goodbye.")
	// cleanup
	for _, app := range apps {
		app.Shutdown()
	}
	for _, integration := range core.GetIntegrations() {
		integration.Shutdown()
	}
}
