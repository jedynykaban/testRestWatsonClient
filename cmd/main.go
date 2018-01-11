package main

import (
	"fmt"
	"io"
	"strings"

	"github.com/jedynykaban/testRestWatsonClient/services/categories/watson"
	"github.com/jedynykaban/testRestWatsonClient/utils"

	log "github.com/Sirupsen/logrus"
	//"github.com/hashicorp/go-multierror"
)

var config Config

func init() {
	config = getConfig()
	setupLogging(config.Service.LogOutput, config.Service.LogLevel, config.Service.LogFormat)
	config.Log()
}

func setupLogging(output io.Writer, level log.Level, format string) {
	log.SetOutput(output)
	log.SetLevel(level)
	if strings.EqualFold(format, "json") {
		log.SetFormatter(&log.JSONFormatter{})
	}
}

func main() {
	log.Info("application started")

	fmt.Println("Hello world")

	httpOptions := utils.DefaultHttpProxyOptions()
	httpOptions.MaxRetries = config.Watson.RetrieveRetries
	watsonClient := createWatsonClient(config.Watson, httpOptions)
	testWatsonClient(watsonClient)

	log.Info("application completed")
}

// createWatsonClient creates and initializes IBM Watson client pkg
func createWatsonClient(cfg WatsonConfig, httpOptions utils.HttpProxyOptions) watson.Client {
	watsonHTTPProxy := utils.NewHTTPProxy(cfg.RetrieveTimeout)
	watsonClient := watson.NewClient(cfg.NaturalLangAPIEndpoint, watsonHTTPProxy, httpOptions)

	return watsonClient
}

func testWatsonClient(client watson.Client) {
	log.Infoln("starging Watson tests")

	testText := "I%20still%20have%20a%20dream%2C%20a%20dream%20deeply%20rooted%20in%20the%20American%20dream%20–%20one%20day%20this%20nation%20will%20rise%20up%20and%20live%20up%20to%20its%20creed%2C%20We%20hold%20these%20truths%20to%20be%20self%20evident%3A%20that%20all%20men%20are%20created%20equal"
	//testURL := "https://newatlas.com/stem-cell-muscle-growth/52894/"

	categories, err := client.GetFeatures("2017-02-27", map[string]string{"text": testText}, []string{"categories"})
	if err != nil {
		log.Errorf("Categories error: %v", err)
		return
	}

	idx := 0
	for _, cat := range categories {
		idx++
		log.Infoln("Category %d: %v", idx, cat)
	}

	log.Infoln("completed Watson tests")
}
