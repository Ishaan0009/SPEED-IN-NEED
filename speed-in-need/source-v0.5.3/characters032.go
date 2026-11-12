package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
)

type characterVertex032 struct {
	A, NA, B, NB [3]float32
	UV           [2]float32
	BoneA, BoneB uint32
	Weight       float32
}
type characterLOD032 struct {
	Vertices []characterVertex032
	Indices  []uint32
}
type characterAsset032 struct {
	Name, Albedo string
	LODs         []characterLOD032
}

// This pure decoder validates offline imports and packaged data before upload.
func decodeCharacter032(raw []byte) (*characterAsset032, error) {
	if len(raw) < 8 || string(raw[:4]) != "CHR2" {
		return nil, fmt.Errorf("invalid character header")
	}
	n := int(binary.LittleEndian.Uint32(raw[4:8]))
	if n < 2 || n > len(raw)-8 {
		return nil, fmt.Errorf("invalid character metadata")
	}
	var header struct {
		Name, Albedo string
		LODs         []struct{ Vertices, Indices int }
	}
	if err := json.Unmarshal(raw[8:8+n], &header); err != nil {
		return nil, err
	}
	if len(header.LODs) != 3 || header.Name == "" || header.Albedo == "" {
		return nil, fmt.Errorf("incomplete character")
	}
	a := &characterAsset032{Name: header.Name, Albedo: header.Albedo}
	reader := bytes.NewReader(raw[8+n:])
	for _, desc := range header.LODs {
		if desc.Vertices < 3 || desc.Vertices > 40000 || desc.Indices < 3 || desc.Indices > 100000 || desc.Indices%3 != 0 {
			return nil, fmt.Errorf("character exceeds mesh budget")
		}
		lod := characterLOD032{Vertices: make([]characterVertex032, desc.Vertices), Indices: make([]uint32, desc.Indices)}
		if err := binary.Read(reader, binary.LittleEndian, lod.Vertices); err != nil {
			return nil, err
		}
		if err := binary.Read(reader, binary.LittleEndian, lod.Indices); err != nil {
			return nil, err
		}
		for _, v := range lod.Vertices {
			if v.BoneA > 14 || v.BoneB > 14 || !(v.Weight >= 0 && v.Weight <= 1) {
				return nil, fmt.Errorf("invalid character skin weight")
			}
			for _, q := range [][3]float32{v.A, v.NA, v.B, v.NB} {
				for _, x := range q {
					if math.IsNaN(float64(x)) || math.IsInf(float64(x), 0) || math.Abs(float64(x)) > 50 {
						return nil, fmt.Errorf("invalid character coordinate")
					}
				}
			}
			for _, x := range v.UV {
				if math.IsNaN(float64(x)) || math.IsInf(float64(x), 0) {
					return nil, fmt.Errorf("invalid character UV")
				}
			}
		}
		for _, i := range lod.Indices {
			if int(i) >= len(lod.Vertices) {
				return nil, fmt.Errorf("invalid character triangle")
			}
		}
		a.LODs = append(a.LODs, lod)
	}
	if reader.Len() != 0 {
		return nil, fmt.Errorf("trailing character bytes")
	}
	return a, nil
}
