package lan

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"fmt"
	"io"
	"time"
)

const datagram051 = 1200

// Keep UDP below the usual Ethernet/Wi-Fi MTU. A snapshot may span bounded
// application fragments; no IP fragmentation or unbounded allocation is needed.
func frames051(raw []byte, id uint64) [][]byte {
	if len(raw) > maxPacket {
		return nil
	}
	if len(raw) > 900 {
		var b bytes.Buffer
		b.WriteString("SZ51")
		w, _ := zlib.NewWriterLevel(&b, zlib.BestSpeed)
		w.Write(raw)
		w.Close()
		if b.Len() < len(raw) {
			raw = b.Bytes()
		}
	}
	if len(raw) <= datagram051 {
		return [][]byte{raw}
	}
	const payload = datagram051 - 14
	count := (len(raw) + payload - 1) / payload
	out := make([][]byte, 0, count)
	for i := 0; i < count; i++ {
		end := (i + 1) * payload
		if end > len(raw) {
			end = len(raw)
		}
		b := make([]byte, 14+end-i*payload)
		copy(b, "SC51")
		binary.LittleEndian.PutUint64(b[4:12], id)
		b[12], b[13] = byte(i), byte(count)
		copy(b[14:], raw[i*payload:end])
		out = append(out, b)
	}
	return out
}

type assembly051 struct {
	at    time.Time
	parts [][]byte
	size  int
}
type assembler051 map[string]*assembly051

func (a assembler051) push(address string, b []byte, now time.Time) ([]byte, bool) {
	if len(b) > datagram051 {
		return nil, false
	}
	if len(b) < 4 || string(b[:4]) != "SC51" {
		return decode051(b)
	}
	if len(b) < 15 {
		return nil, false
	}
	id := binary.LittleEndian.Uint64(b[4:12])
	part, count := int(b[12]), int(b[13])
	if count < 2 || count > 16 || part >= count {
		return nil, false
	}
	for k, v := range a {
		if now.Sub(v.at) > 400*time.Millisecond {
			delete(a, k)
		}
	}
	key := fmt.Sprintf("%s/%d", address, id)
	p := a[key]
	if p == nil {
		if len(a) >= 32 {
			return nil, false
		}
		p = &assembly051{at: now, parts: make([][]byte, count)}
		a[key] = p
	}
	if len(p.parts) != count {
		delete(a, key)
		return nil, false
	}
	if p.parts[part] == nil {
		p.parts[part] = append([]byte(nil), b[14:]...)
		p.size += len(b) - 14
	}
	if p.size > maxPacket {
		delete(a, key)
		return nil, false
	}
	for _, v := range p.parts {
		if v == nil {
			return nil, false
		}
	}
	raw := make([]byte, 0, p.size)
	for _, v := range p.parts {
		raw = append(raw, v...)
	}
	delete(a, key)
	return decode051(raw)
}
func decode051(b []byte) ([]byte, bool) {
	if len(b) >= 4 && string(b[:4]) == "SZ51" {
		z, err := zlib.NewReader(bytes.NewReader(b[4:]))
		if err != nil {
			return nil, false
		}
		defer z.Close()
		raw, err := io.ReadAll(io.LimitReader(z, maxPacket+1))
		return raw, err == nil && len(raw) <= maxPacket
	}
	return b, len(b) <= maxPacket
}

func ping051(old int, sent int64, now time.Time) int {
	ms := now.UnixMilli() - sent
	if sent <= 0 || ms < 0 || ms > 5000 {
		return old
	}
	if ms < 1 {
		ms = 1
	}
	if old <= 0 {
		return int(ms)
	}
	return (old*3 + int(ms)) / 4
}
