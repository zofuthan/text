// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package display

import (
	"fmt"
	"testing"

	"github.com/zofuthan/text/internal/testtext"
)

func TestLinking(t *testing.T) {
	base := getSize(t, `display.Tags(language.English).Name(language.English)`)
	compact := getSize(t, `display.English.Languages().Name(language.English)`)

	if d := base - compact; d < 1.5*1024*1024 {
		t.Errorf("size(base)-size(compact) was %d; want > 1.5MB", base, compact)
	}
}

func getSize(t *testing.T, main string) int {
	size, err := testtext.CodeSize(fmt.Sprintf(body, main))
	if err != nil {
		t.Fatal(err)
	}
	return size
}

const body = `package main
import (
	"github.com/zofuthan/text/display"
	"github.com/zofuthan/text/language"
)
func main() {
	%s
}
`
