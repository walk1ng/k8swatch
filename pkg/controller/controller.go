package controller

import (
	"fmt"

	"github.com/walk1ng/k8swatch/pkg/config"
	"github.com/walk1ng/k8swatch/pkg/handlers"
)

// Start starts the controller
func Start(c *config.Config, eventHandler handlers.Handler) {
	fmt.Printf("config:\n%+v\nhandler:\n%+v\n", c, eventHandler)
}
