package main

import (
	"bakapy"
	"flag"
	"fmt"
	"os"
	"reflect"
)

var USAGE = "Usage: bakapy-show-meta files..."
var CONFIG_PATH = flag.String("config", "scheduler.conf", "Bakapy configuration file")
var ONLY_KEY = flag.String("key", "", "Show only specified key from metadata")

func printMetadata(metadata bakapy.Metadata) {
	fmt.Printf("==> [%s]%s\n", metadata.JobName, metadata.TaskId)
	fmt.Println("==> Success:", metadata.Success)
	fmt.Println("==> Command:", metadata.Command)
	fmt.Println("==> AvgSpeed:", metadata.AvgSpeed())
	fmt.Println("==> PID:", metadata.Pid)
	fmt.Println("==> Start:", metadata.StartTime)
	fmt.Println("==> End:", metadata.EndTime)
	fmt.Println("==> Duration:", metadata.Duration())
	fmt.Println("==> Files:", metadata.Files)
	fmt.Println("==> Size:", metadata.TotalSize)
	fmt.Println("==> Expire:", metadata.ExpireTime)
	fmt.Printf("==> Output:\n%s\n", string(metadata.Output))
	fmt.Printf("==> Errput:\n%s\n", string(metadata.Errput))
	fmt.Println("==================================")
}

func main() {
	flag.Parse()
	if len(flag.Args()) == 0 {
		fmt.Println(USAGE)
		os.Exit(1)
	}
	if err := bakapy.SetupLogging("warning"); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	config, err := bakapy.ParseConfig(*CONFIG_PATH)
	if err != nil {
		fmt.Println("Cannot read config", err)
		os.Exit(1)
	}

	metaman := bakapy.NewMetaManClient(config.MetadataAddr, config.Secret)
	taskId := bakapy.TaskId(flag.Arg(0))

	meta, err := metaman.View(taskId)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot load metadata: %s\n", err)
		os.Exit(1)
	}
	if *ONLY_KEY != "" {
		field := reflect.ValueOf(meta).FieldByName(*ONLY_KEY)
		if !field.IsValid() {
			fmt.Fprintf(os.Stderr, "Key %s not found\n", *ONLY_KEY)
			os.Exit(1)
		}
		fmt.Println(field.Interface())
	} else {
		printMetadata(meta)
	}

}
