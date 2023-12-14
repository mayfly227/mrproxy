package common

import (
	"github.com/oschwald/geoip2-golang"
	"log"
	"sync"
)

var (
	mmdb     *geoip2.Reader
	onceMMDB sync.Once
)

func LoadFromBytes(buffer []byte) {
	onceMMDB.Do(func() {
		var err error
		mmdb, err = geoip2.FromBytes(buffer)
		if err != nil {
			log.Fatalf("Can't load mmdb: %s\n", err.Error())
		}
	})
}

func Verify() bool {
	instance, err := geoip2.Open("Country.mmdb")
	if err == nil {
		err := instance.Close()
		if err != nil {
			return false
		}
	}
	return err == nil
}

func Instance() *geoip2.Reader {
	onceMMDB.Do(func() {
		var err error
		mmdb, err = geoip2.Open("./Country.mmdb")
		if err != nil {
			log.Fatalf("Can't load mmdb: %s\n", err.Error())
		}
	})
	return mmdb
}
