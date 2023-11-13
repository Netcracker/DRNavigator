package main

import (
	"github.com/netcracker/drnavigator/site-manager-cr-controller/cmd"
)

// @title           site-manager
// @version         1.0
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	cmd.Execute()
}
