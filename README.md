# RunCfg Go Client

### Usage in projects

First download dependency using go get
```shell
$ go get -u github.com/runcfg/runcfg-go
```

Then import into your project

```go
import "github.com/runcfg/runcfg-go"
```

### Using your first config

1. Create an account at https://runcfg.com
2. Download your `.runcfg` file from your project page at https://runcfg.com by clicking (get .runcfg file)

![runcfg.PNG](https://raw.githubusercontent.com/runcfg/runcfg-net/main/runcfg.png)

3. Place your `.runcfg` file at the root of your project
4. Create an instance of the client in your code as follows:
   
```go
// my-config is your project name which is prepended to your .runcfg file e.g. `my-config.runcfg`

client, err := runcfg.Create("my-config") 
if err != nil {
    t.Error(err)
}
```

5. create your config type
```go
type ExampleConfig struct {
	Version string `json:"version"`
	Target  string `json:"target"`
	Enabled string `json:"enabled"`
}
```

6. load your remote config into your config type
```go
var config ExampleConfig // create instance of your config type

err = client.LoadConfigAsType(&config)
if err != nil {
	log.Fatalf("LoadConfigAsType failure:\n%s", err)
}

// optional: watch for config changes and invoke callback
client.Watch(watchInterval, func() {
	fmt.Println("config updated")
})	
```

You can now access your configuration from the 
config type which you passed into the `LoadConfigAsType` function for example:

