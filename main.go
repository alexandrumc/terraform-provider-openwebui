package main

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"

	"github.com/nickcecere/terraform-provider-openwebui/internal/provider"
	"flag"
)

func main() {
	var debug bool
	flag.BoolVar(&debug, "debug", false, "set to true to run in debug mode")
	flag.Parse()
	providerserver.Serve(context.Background(), provider.New, providerserver.ServeOpts{
		Address: "registry.terraform.io/nickcecere/openwebui",
		Debug: debug,
	})
}
