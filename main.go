package main

import (
	"fmt"
        "os"
        "os/signal"
        "syscall"
	"time"
        "strings"
	"encoding/json"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"github.com/bitly/go-simplejson"
)

func serverStatus(mongo_config Mongo) *simplejson.Json {
	info := mgo.DialInfo{
		Addrs:   mongo_config.Addresses,
		Direct:  false,
		Timeout: time.Second * 30,
	}

	session, err := mgo.DialWithInfo(&info)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	if len(mongo_config.User) > 0 {
		cred := mgo.Credential{Username: mongo_config.User, Password: mongo_config.Pass}
		err = session.Login(&cred)
		if err != nil {
			panic(err)
		}
	}

	// Optional. Switch the session to a monotonic behavior.
	session.SetMode(mgo.Monotonic, true)

	var result = &bson.M{}
	if err := session.Run("serverStatus", result); err != nil {
		panic(err)
	}

        resultJs, err := json.Marshal(result)
        if err != nil {
		panic(err)
        }

        js, err := simplejson.NewJson(resultJs)
        if err != nil {
                panic(err)
        }

	return js
}

func printSpec(j *simplejson.Json, o []string) string {
        result := ""
	for _, v := range o {
		result = result + fmt.Sprintf(" %s %d", strings.Replace(v, " ", "_", -1), j.Get(v).MustInt64())
	}
	return result
}

func printAll(j *simplejson.Json) {
	fmt.Println("-----------------")

	assert := j.Get("asserts")
	fmt.Println(fmt.Sprintf("Asserts:%s @ %s", printSpec(assert, []string{"regular", "warning", "msg", "user", "rollovers"}), j.Get("localTime").MustString()))

	conn := j.Get("connections")
	fmt.Println(fmt.Sprintf("Connections:%s @ %s", printSpec(conn, []string{"current", "available", "totalCreated"}), j.Get("localTime").MustString()))

	mem := j.Get("mem")
	extra := j.Get("extra_info")
	fmt.Println(fmt.Sprintf("Memory:%s %s @ %s", printSpec(mem, []string{"resident", "virtual"}), printSpec(extra, []string{"page_faults", "heap_usage_bytes"}), j.Get("localTime").MustString()))

	globallock := j.Get("globalLock")
	fmt.Println(fmt.Sprintf("GlobalLock:%s @ %s", printSpec(globallock, []string{"totalTime"}), j.Get("localTime").MustString()))
	currentQueue := globallock.Get("currentQueue")
	fmt.Println(fmt.Sprintf("GlobalLocks.currentQueue:%s @ %s", printSpec(currentQueue, []string{"readers", "writers", "total"}), j.Get("localTime").MustString()))
	activeClients := globallock.Get("activeClients")
	fmt.Println(fmt.Sprintf("GlobalLocks.activeClients:%s @ %s", printSpec(activeClients, []string{"readers", "writers", "total"}), j.Get("localTime").MustString()))

	op := j.Get("opcounters")
	fmt.Println(fmt.Sprintf("Op:%s @ %s", printSpec(op, []string{"insert", "query", "update", "delete", "getmore", "command"}), j.Get("localTime").MustString()))

        // wiredTiger

	wiredTiger := j.Get("wiredTiger")

	concurrent := wiredTiger.Get("concurrentTransactions")

	read := concurrent.Get("read")
	fmt.Println(fmt.Sprintf("WiredTigger.concurrentTransactions.read:%s @ %s", printSpec(read, []string{"out", "available", "totalTickets"}), j.Get("localTime").MustString()))
	write := concurrent.Get("write")
	fmt.Println(fmt.Sprintf("WiredTigger.concurrentTransactions.write:%s @ %s", printSpec(write, []string{"out", "available", "totalTickets"}), j.Get("localTime").MustString()))

	transaction := wiredTiger.Get("transaction")
	fmt.Println(fmt.Sprintf("WiredTigger.transaction:%s @ %s", printSpec(transaction, []string{"transactions committed", "transactions rolled back"}), j.Get("localTime").MustString()))

	log := wiredTiger.Get("log")
	fmt.Println(fmt.Sprintf("WiredTigger.log:%s @ %s", printSpec(log, []string{"log bytes written", "log write operations", "total log buffer size"}), j.Get("localTime").MustString()))

	cache := wiredTiger.Get("cache")
	fmt.Println(fmt.Sprintf("WiredTigger.cache:%s @ %s", printSpec(cache, []string{"bytes currently in the cache", "bytes read into cache", "bytes written from cache", "tracked dirty bytes in the cache", "modified pages evicted", "unmodified pages evicted"}), j.Get("localTime").MustString()))
}

func main() {
	config := LoadConfig()

        ticker := time.NewTicker(config.Interval)
        quit := make(chan struct{})
        go func() {
                for {
                        select {
                        case <-ticker.C:
				status := serverStatus(config.Mongo)
				printAll(status)
                        }
                }
        }()
        ch := make(chan os.Signal)
        signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
        sig := <-ch
        fmt.Println("Received " + sig.String())
        close(quit)
}
