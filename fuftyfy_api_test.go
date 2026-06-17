package fuftyfy_api

import (
	"testing"
)

func TestParseGCSURL(t *testing.T) {
	tests := []struct {
		url        string
		wantBucket string
		wantObject string
		wantOk     bool
	}{
		{
			url:        "https://storage.googleapis.com/zinovik-gallery/hedgehogs/103a.jpg",
			wantBucket: "zinovik-gallery",
			wantObject: "hedgehogs/103a.jpg",
			wantOk:     true,
		},
		{
			url:        "gs://my-bucket/folder/subfolder/file.txt",
			wantBucket: "my-bucket",
			wantObject: "folder/subfolder/file.txt",
			wantOk:     true,
		},
		{
			url:        "https://example.com/not-gcs/file.jpg",
			wantBucket: "",
			wantObject: "",
			wantOk:     false,
		},
		{
			url:        "",
			wantBucket: "",
			wantObject: "",
			wantOk:     false,
		},
	}

	for _, tt := range tests {
		bucket, object, ok := parseGCSURL(tt.url)
		if ok != tt.wantOk {
			t.Errorf("parseGCSURL(%q) ok = %v; want %v", tt.url, ok, tt.wantOk)
		}
		if bucket != tt.wantBucket {
			t.Errorf("parseGCSURL(%q) bucket = %q; want %q", tt.url, bucket, tt.wantBucket)
		}
		if object != tt.wantObject {
			t.Errorf("parseGCSURL(%q) object = %q; want %q", tt.url, object, tt.wantObject)
		}
	}
}
