package examples

// import (
// 	"fmt"
// 	"log"
// 	"time"

// 	runcfg "github.com/runcfg/runcfg-go/pkg"
// )

// type ExampleConfig struct {
// 	Version string `json:"version"`
// 	Enabled string `json:"enabled"`
// }

// func main() {
// 	var config ExampleConfig
// 	var watchInterval time.Duration = 15

// 	client, err := runcfg.Create("hyper-testing")
// 	if err != nil {
// 		log.Fatalf("%s", err)
// 	}

// 	err = client.LoadConfigAsType("1.0.0", &config)
// 	if err != nil {
// 		log.Fatalf("LoadConfigAsType failure:\n%s", err)
// 	}

// 	client.Watch(watchInterval, func() {
// 		fmt.Println("config updated")
// 	})
// }
