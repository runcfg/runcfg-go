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
2. Download your `.runcfg` file from your project page at https://runcfg.com
3. Place your `.runcfg` file at the root of your project
4. Create an instance of the client in your code as follows:
   
```go
client, err := runcfg.Create()
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
var config ExampleConfig // create instance of config type

err = client.LoadConfigAsType("1.0.0", &config)
if err != nil {
    t.Error(err)
}
```

You can now access your configuration from the 
config type which you passed into the `LoadConfigAsType` function for example:

```go
fmt.Println("My Config Version", config.Version)
```