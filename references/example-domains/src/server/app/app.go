package app

import (
	vineapp "go.yorun.ai/vine/app"
	"go.yorun.ai/vine/app/standalone"

	"example.com/shop/src/server/catalog"
	"example.com/shop/src/server/orders"
)

// New runs two independent App assemblies with one shared local infrastructure.
func New(seedPath string, adminListen string) vineapp.App {
	return standalone.NewBundledWithOption(standalone.Option{
		HubSeedDataFile: seedPath,
		HubAdminListen:  adminListen,
	}, standalone.New[*catalog.App](), standalone.New[*orders.App]())
}
