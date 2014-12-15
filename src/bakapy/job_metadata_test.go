package bakapy

import (
	"testing"
	"time"
)

func TestJobMetadataDuration_Ok(t *testing.T) {
	meta := JobMetadata{
		StartTime: time.Date(2010, 9, 1, 14, 30, 0, 0, time.UTC),
		EndTime:   time.Date(2011, 2, 21, 20, 20, 0, 0, time.UTC),
	}
	d := meta.Duration()
	if d.String() != "4157h50m0s" {
		t.Fatal("duration must be 4157h50m0s, not", d)
	}
}

func TestJobMetadataDuration_StartTimeAfterEndTime(t *testing.T) {
	meta := JobMetadata{
		StartTime: time.Date(2012, 9, 1, 14, 30, 0, 0, time.UTC),
		EndTime:   time.Date(2011, 2, 21, 20, 20, 0, 0, time.UTC),
	}
	d := meta.Duration()
	if d.String() != "0" {
		t.Fatal("duration must be 0, not", d)
	}
}

func TestJobMetadataDuration_NoEndTime(t *testing.T) {
	meta := JobMetadata{
		StartTime: time.Date(2010, 9, 1, 14, 30, 0, 0, time.UTC),
	}
	d := meta.Duration()
	if d.String() != "0" {
		t.Fatal("duration must be 0, not", d)
	}
}

func TestJobMetadataDuration_NoStartTime(t *testing.T) {
	meta := JobMetadata{
		EndTime: time.Date(2010, 9, 1, 14, 30, 0, 0, time.UTC),
	}
	d := meta.Duration()
	if d.String() != "0" {
		t.Fatal("duration must be 0, not", d)
	}
}

func TestJobMetadataDuration_NoStartNoEndTime(t *testing.T) {
	meta := JobMetadata{}
	d := meta.Duration()
	if d.String() != "0" {
		t.Fatal("duration must be 0, not", d)
	}
}

func TestJobMetadataAvgSpeed_Ok(t *testing.T) {
	meta := JobMetadata{
		TotalSize: 102 * 1024 * 1024,
		StartTime: time.Date(2011, 9, 1, 19, 30, 0, 0, time.UTC),
		EndTime:   time.Date(2011, 9, 1, 20, 20, 0, 0, time.UTC),
	}
	s := meta.AvgSpeed()
	if s != 35651 {
		t.Fatal("duration must be 35651, not", s)
	}
}

func TestJobMetadataAvgSpeed_ZeroDuration(t *testing.T) {
	meta := JobMetadata{
		TotalSize: 102 * 1024 * 1024,
	}
	s := meta.AvgSpeed()
	if s != 0 {
		t.Fatal("duration must be 0, not", s)
	}
}

func TestJobMetadataSaveTo_Ok(t *testing.T) {
	meta := JobMetadata{
		TotalSize: 102 * 1024 * 1024,
	}
	err := meta.Save("/dev/null")
	if err != nil {
		t.Fatal("Cannot save metadata:", err)
	}
}
