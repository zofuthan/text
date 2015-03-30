// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build ignore

package main

// This program generates tables.go:
//	go run maketables.go | gofmt > tables.go

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strings"
)

func main() {
	fmt.Printf("// generated by go run maketables.go; DO NOT EDIT\n\n")
	fmt.Printf("// Package traditionalchinese provides Traditional Chinese encodings such as Big5.\n")
	fmt.Printf(`package traditionalchinese // import "github.com/zofuthan/text/encoding/traditionalchinese"` + "\n\n")

	res, err := http.Get("http://encoding.spec.whatwg.org/index-big5.txt")
	if err != nil {
		log.Fatalf("Get: %v", err)
	}
	defer res.Body.Close()

	mapping := [65536]uint32{}
	reverse := [65536 * 4]uint16{}

	scanner := bufio.NewScanner(res.Body)
	for scanner.Scan() {
		s := strings.TrimSpace(scanner.Text())
		if s == "" || s[0] == '#' {
			continue
		}
		x, y := uint16(0), uint32(0)
		if _, err := fmt.Sscanf(s, "%d 0x%x", &x, &y); err != nil {
			log.Fatalf("could not parse %q", s)
		}
		if x < 0 || 126*157 <= x {
			log.Fatalf("Big5 code %d is out of range", x)
		}
		mapping[x] = y

		// The WHATWG spec http://encoding.spec.whatwg.org/#indexes says that
		// "The index pointer for code point in index is the first pointer
		// corresponding to code point in index", which would normally mean
		// that the code below should be guarded by "if reverse[y] == 0", but
		// last instead of first seems to match the behavior of
		// "iconv -f UTF-8 -t BIG5". For example, U+8005 者 occurs twice in
		// http://encoding.spec.whatwg.org/index-big5.txt, as index 2148
		// (encoded as "\x8e\xcd") and index 6543 (encoded as "\xaa\xcc")
		// and "echo 者 | iconv -f UTF-8 -t BIG5 | xxd" gives "\xaa\xcc".
		c0, c1 := x/157, x%157
		if c1 < 0x3f {
			c1 += 0x40
		} else {
			c1 += 0x62
		}
		reverse[y] = (0x81+c0)<<8 | c1
	}
	if err := scanner.Err(); err != nil {
		log.Fatalf("scanner error: %v", err)
	}

	fmt.Printf("// decode is the decoding table from Big5 code to Unicode.\n")
	fmt.Printf("// It is defined at http://encoding.spec.whatwg.org/index-big5.txt\n")
	fmt.Printf("var decode = [...]uint32{\n")
	for i, v := range mapping {
		if v != 0 {
			fmt.Printf("\t%d: 0x%08X,\n", i, v)
		}
	}
	fmt.Printf("}\n\n")

	// Any run of at least separation continuous zero entries in the reverse map will
	// be a separate encode table.
	const separation = 1024

	intervals := []interval(nil)
	low, high := -1, -1
	for i, v := range reverse {
		if v == 0 {
			continue
		}
		if low < 0 {
			low = i
		} else if i-high >= separation {
			if high >= 0 {
				intervals = append(intervals, interval{low, high})
			}
			low = i
		}
		high = i + 1
	}
	if high >= 0 {
		intervals = append(intervals, interval{low, high})
	}
	sort.Sort(byDecreasingLength(intervals))

	fmt.Printf("const numEncodeTables = %d\n\n", len(intervals))
	fmt.Printf("// encodeX are the encoding tables from Unicode to Big5 code,\n")
	fmt.Printf("// sorted by decreasing length.\n")
	for i, v := range intervals {
		fmt.Printf("// encode%d: %5d entries for runes in [%6d, %6d).\n", i, v.len(), v.low, v.high)
	}
	fmt.Printf("\n")

	for i, v := range intervals {
		fmt.Printf("const encode%dLow, encode%dHigh = %d, %d\n\n", i, i, v.low, v.high)
		fmt.Printf("var encode%d = [...]uint16{\n", i)
		for j := v.low; j < v.high; j++ {
			x := reverse[j]
			if x == 0 {
				continue
			}
			fmt.Printf("\t%d-%d: 0x%04X,\n", j, v.low, x)
		}
		fmt.Printf("}\n\n")
	}
}

// interval is a half-open interval [low, high).
type interval struct {
	low, high int
}

func (i interval) len() int { return i.high - i.low }

// byDecreasingLength sorts intervals by decreasing length.
type byDecreasingLength []interval

func (b byDecreasingLength) Len() int           { return len(b) }
func (b byDecreasingLength) Less(i, j int) bool { return b[i].len() > b[j].len() }
func (b byDecreasingLength) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }
