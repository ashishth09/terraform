package main

import (
	"github.com/ashishth09/terraform/builtin/providers/bluemix"
	"github.com/hashicorp/terraform/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: bluemix.Provider,
	})
}
