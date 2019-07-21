package main

import (
	"github.com/antonrutkevich/simple-memcache/config"
	"github.com/antonrutkevich/simple-memcache/pkg/domain"
	"github.com/antonrutkevich/simple-memcache/pkg/infrastructure/log"
	"github.com/antonrutkevich/simple-memcache/pkg/infrastructure/transport/telnet"
	"github.com/antonrutkevich/simple-memcache/pkg/memcache"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	conf := config.MustLoad()
	logger := log.CreateLogger(conf.Log)

	logger.Infof("Starting memcache %s:%s from %s on port %s\n", version, commit, date, conf.Server.Port)



	memcache.NewMemCache(
		conf.Server,
		logger,
		telnet.NewEncoder(),
		telnet.NewDecoder(),
		domain.NewEngine(),
	)
}