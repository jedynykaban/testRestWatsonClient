package main

import (
	//"context"
	"fmt"
	"sync"
	//"time"
	//"encoding/json"
	"io"
	"strings"

	"cloud.google.com/go/datastore"
	log "github.com/Sirupsen/logrus"

	"github.com/hashicorp/go-multierror"
	//"github.com/jinzhu/now"
)

type x struct {
	ID    string
	Name  string
	Value int
}

var config Config

func init() {
	config = getConfig()
	setupLogging(config.Service.LogOutput, config.Service.LogLevel, config.Service.LogFormat)
}

func setupLogging(output io.Writer, level log.Level, format string) {
	log.SetOutput(output)
	log.SetLevel(level)
	if strings.EqualFold(format, "json") {
		log.SetFormatter(&log.JSONFormatter{})
	}
}

func mainC() {
	log.Info("application started")

	err := updateConncurencyTest()
	if err != nil {
		log.Error(err)
	}

	log.Info("application completed")
}

func mainX() {
	log.Info("application started")

	set1 := []string{"Ala", "ma", "kota"}
	set1Flatten := strings.Join(set1, ",")
	log.Info("Set 1: ", set1Flatten)
	log.Info("Set 1 len: ", len(strings.Split(set1Flatten, ",")))

	set2 := []string{}
	set2Flatten := strings.Join(set2, ",")
	log.Info("Set 2: ", set2Flatten)
	log.Info("Set 2 len: ", len(strings.Split(set2Flatten, ",")))
	log.Info("Set 2 value: ", strings.Split(set2Flatten, ",")[0])

	log.Info("application completed")
}

func main() {
	log.Info("application started")

	//ctx := context.Background()
	//client, _ := datastore.NewClient(ctx, "mosaiqio-dev")

	// for idx := range pubs {
	// 	log.Infof("Publisher: %v (key: %v, decoded: %v)\n", pubs[idx], keys[idx], keys[idx].Encode())
	// }

	// Dagens PS Rss Source: Key(RssSource, 5630999308271616)
	// Dagens PS production Publication Id: Key(Publication, 5705562083819520): EhYKC1B1YmxpY2F0aW9uEICAgIrbpZEK

	eKey := "EhAKBU1pdGVtEICAgO6yqsEK"
	dKey := datastore.Key{
		Kind: "Publication",
		ID:   5705562083819520,
	}

	if len(eKey) > 0 {
		key, _ := datastore.DecodeKey(eKey)
		log.Infof("Encoded key: %v;  Decoded key: %v)\n", eKey, key)
	} else {
		log.Infof("Decoded key: %v;  Encoded key: %v)\n", dKey, dKey.Encode())
	}

	log.Info("application completed")
}

func updateConncurencyTest() error {
	// fist need to find all playlists that particular mitem belongs to (in this partcular case, regardless of serviced one)
	numbers := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}

	// for each run update
	var wg sync.WaitGroup
	errorChan := make(chan error, len(numbers))
	for _, n := range numbers {
		wg.Add(1)
		go func(number int, wg *sync.WaitGroup, errorChan chan error) {
			defer wg.Done()
			errorChan <- checkIfEven(number)
			return
		}(n, &wg, errorChan)
	}
	wg.Wait()
	close(errorChan)

	var ret error
	for err := range errorChan {
		if err != nil {
			ret = multierror.Append(ret, err)
		}
	}
	return ret
}

func checkIfEven(number int) error {
	if number%2 != 0 {
		return fmt.Errorf("number: %v is not even number", number)
	}
	return nil
}
