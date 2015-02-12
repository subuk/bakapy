package bakapy

import (
	"testing"
	"time"
)

func TestMetadataDuration_Ok(t *testing.T) {
	meta := Metadata{
		StartTime: time.Date(2010, 9, 1, 14, 30, 0, 0, time.UTC),
		EndTime:   time.Date(2011, 2, 21, 20, 20, 0, 0, time.UTC),
	}
	d := meta.Duration()
	if d.String() != "4157h50m0s" {
		t.Fatal("duration must be 4157h50m0s, not", d)
	}
}

func TestMetadataDuration_StartTimeAfterEndTime(t *testing.T) {
	meta := Metadata{
		StartTime: time.Date(2012, 9, 1, 14, 30, 0, 0, time.UTC),
		EndTime:   time.Date(2011, 2, 21, 20, 20, 0, 0, time.UTC),
	}
	d := meta.Duration()
	if d.String() != "0" {
		t.Fatal("duration must be 0, not", d)
	}
}

func TestMetadataDuration_NoEndTime(t *testing.T) {
	meta := Metadata{
		StartTime: time.Date(2010, 9, 1, 14, 30, 0, 0, time.UTC),
	}
	d := meta.Duration()
	if d.String() != "0" {
		t.Fatal("duration must be 0, not", d)
	}
}

func TestMetadataDuration_NoStartTime(t *testing.T) {
	meta := Metadata{
		EndTime: time.Date(2010, 9, 1, 14, 30, 0, 0, time.UTC),
	}
	d := meta.Duration()
	if d.String() != "0" {
		t.Fatal("duration must be 0, not", d)
	}
}

func TestMetadataDuration_NoStartNoEndTime(t *testing.T) {
	meta := Metadata{}
	d := meta.Duration()
	if d.String() != "0" {
		t.Fatal("duration must be 0, not", d)
	}
}

func TestMetadataAvgSpeed_Ok(t *testing.T) {
	meta := Metadata{
		TotalSize: 102 * 1024 * 1024,
		StartTime: time.Date(2011, 9, 1, 19, 30, 0, 0, time.UTC),
		EndTime:   time.Date(2011, 9, 1, 20, 20, 0, 0, time.UTC),
	}
	s := meta.AvgSpeed()
	if s != 35651 {
		t.Fatal("duration must be 35651, not", s)
	}
}

func TestMetadataAvgSpeed_ZeroDuration(t *testing.T) {
	meta := Metadata{
		TotalSize: 102 * 1024 * 1024,
	}
	s := meta.AvgSpeed()
	if s != 0 {
		t.Fatal("duration must be 0, not", s)
	}
}

// func TestMetadataSaveTo_Ok(t *testing.T) {
// 	meta := Metadata{
// 		TotalSize: 102 * 1024 * 1024,
// 	}
// 	err := meta.Save("/dev/null")
// 	if err != nil {
// 		t.Fatal("Cannot save metadata:", err)
// 	}
// }
