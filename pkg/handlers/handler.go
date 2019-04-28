package handlers

import (
	"github.com/walk1ng/k8swatch/pkg/config"
)

// Handler interface can be implemented by any handler
// The handler methods are used to process events
type Handler interface {
	Init(c *config.Config) error
	ObjCreated(obj interface{})
	ObjUpdated(obj interface{})
	ObjDeleted(old, new interface{})
}

// Default handler implement
// Print event with json format
type Default struct {
}

// Init initializes handler configuration
// do nothing for Default handler
func (d *Default) Init(c *config.Config) error {
	return nil
}

// ObjCreated handle object created event
func (d *Default) ObjCreated(obj interface{}) {

}

// ObjUpdated handle object updated event
func (d *Default) ObjUpdated(obj interface{}) {

}

// ObjDeleted handle object deleted event
func (d *Default) ObjDeleted(old, new interface{}) {

}
