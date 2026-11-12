package main

import (
	"bytes"
	"errors"
	"testing"
)

type checksumFailure026 struct{}

func (checksumFailure026) Read(p []byte) (int, error) {
	copy(p, "abc")
	return 3, errors.New("fixture checksum failure")
}

func TestAssetReader026RetainsEOFAndChecksumValidation(t *testing.T) {
	for _, data := range [][]byte{nil, []byte("abc"), bytes.Repeat([]byte{7}, 1<<20)} {
		got, err := readAssetBytes026(bytes.NewReader(data), int64(len(data)))
		if err != nil || !bytes.Equal(got, data) {
			t.Fatal("asset changed or valid EOF rejected", err)
		}
	}
	for _, size := range []int64{2, 4, -1, 193 << 20} {
		if _, err := readAssetBytes026(bytes.NewReader([]byte("abc")), size); err == nil {
			t.Fatalf("invalid size %d accepted", size)
		}
	}
	if _, err := readAssetBytes026(checksumFailure026{}, 3); err == nil || err.Error() != "fixture checksum failure" {
		t.Fatal("checksum failure was hidden", err)
	}
}
