package main

import (
	"github.com/hashicorp/terraform/builtin/providers/bluemix"
	"github.com/hashicorp/terraform/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: bluemix.Provider,
	})
}
