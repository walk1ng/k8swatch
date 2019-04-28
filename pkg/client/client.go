package client

import (
	"github.com/Sirupsen/logrus"
	"github.com/walk1ng/k8swatch/pkg/config"
	"github.com/walk1ng/k8swatch/pkg/controller"
	"github.com/walk1ng/k8swatch/pkg/handlers"
)

// Run runs the event processing with the given handler
func Run(c *config.Config) {
	var eventHandler handlers.Handler
	// currently only support the Default handler

	switch {
	default:
		eventHandler = &handlers.Default{}
	}

	if err := eventHandler.Init(c); err != nil {
		logrus.Fatal(err)
	}

	controller.Start(c, eventHandler)
}
