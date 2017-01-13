package main

import (
	"flag"
	"fmt"
	"time"
        "github.com/vharitonsky/iniflags"
)

type sl []string

type Mongo struct {
	Addresses []string
	User      string
	Pass      string
}

type Config struct {
	Interval time.Duration
	Mongo    Mongo
}

func (s *sl) String() string {
	return fmt.Sprintf("%s", *s)
}

func (s *sl) Set(value string) error {
	*s = append(*s, value)
	return nil
}

var mongo_addresses sl

func LoadConfig() Config {
	var (
		mongo_user     = flag.String("mongo_user", "", "MongoDB User")
		mongo_pass     = flag.String("mongo_pass", "", "MongoDB Password")
		interval       = flag.Duration("interval", 5*time.Second, "Polling interval")
	)

	flag.Var(&mongo_addresses, "mongo_address", "List of mongo addresses in host:port format")
        iniflags.Parse()
	if len(mongo_addresses) == 0 {
		mongo_addresses = append(mongo_addresses, "localhost:27017")
	}
	cfg := Config{
		Interval: *interval,
		Mongo: Mongo{
			Addresses: mongo_addresses,
			User:      *mongo_user,
			Pass:      *mongo_pass,
		},
	}

	return cfg
}
