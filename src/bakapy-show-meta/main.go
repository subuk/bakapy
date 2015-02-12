package main

import (
	"bakapy"
	"fmt"
	"os"
	"path"
	"sort"
)

var USAGE = "Usage: bakapy-show-meta files..."

type ByStartTime []*bakapy.Metadata

func (a ByStartTime) Len() int           { return len(a) }
func (a ByStartTime) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByStartTime) Less(i, j int) bool { return a[i].StartTime.Before(a[j].StartTime) }

func printMetadata(metadata *bakapy.Metadata) {
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
	if len(os.Args[1:]) == 0 {
		fmt.Println(USAGE)
		os.Exit(1)
	}

	var metas []*bakapy.Metadata
	for _, metaPath := range os.Args[1:] {
		metaman := bakapy.NewMetaMan(&bakapy.Config{MetadataDir: path.Dir(metaPath)})
		meta, err := metaman.View(bakapy.TaskId(path.Base(metaPath)))
		if err != nil {
			fmt.Fprintf(os.Stderr, "[warning] %s: %s\n", metaPath, err)
			continue
		}
		metas = append(metas, &meta)
	}
	if len(metas) == 0 {
		fmt.Println("[warning] no valid metadatas found")
		return
	}

	sort.Sort(ByStartTime(metas))

	for _, m := range metas {
		printMetadata(m)
	}

}
