package main

import (
	"bakapy"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
)

var USAGE = "Usage: bakapy-show-meta files..."

type ByStartTime []*bakapy.JobMetadata

func (a ByStartTime) Len() int           { return len(a) }
func (a ByStartTime) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByStartTime) Less(i, j int) bool { return a[i].StartTime.Before(a[j].StartTime) }

func readMetadata(metaPath string) (*bakapy.JobMetadata, error) {
	data, err := ioutil.ReadFile(metaPath)
	if err != nil {
		return nil, err
	}
	var metadata bakapy.JobMetadata
	err = json.Unmarshal(data, &metadata)
	if err != nil {
		return nil, err
	}
	return &metadata, nil
}

func printMetadata(metadata *bakapy.JobMetadata) {
	fmt.Println("==>", metadata.TaskId)
	fmt.Println("==> Success:", metadata.Success)
	fmt.Println("==> Name:", metadata.Command)
	fmt.Println("==> PID:", metadata.Pid)
	fmt.Println("==> Start:", metadata.StartTime)
	fmt.Println("==> End:", metadata.EndTime)
	fmt.Println("==> Duration", metadata.Duration())
	fmt.Println("==> Files", metadata.Files)
	fmt.Println("==> Size:", metadata.TotalSize)
	fmt.Printf("==> Output:\n%s\n", string(metadata.Output))
	fmt.Printf("==> Errput:\n%s\n", string(metadata.Errput))
	fmt.Println("==================================")
}

func main() {
	if len(os.Args[1:]) == 0 {
		fmt.Println(USAGE)
		os.Exit(1)
	}
	var metas []*bakapy.JobMetadata
	for _, metaPath := range os.Args[1:] {
		meta, err := readMetadata(metaPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[warning] %s: %s\n", metaPath, err)
			continue
		}
		metas = append(metas, meta)
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
