# Updated TIFF handling library

Please note that this is a forked version of a circa 2014 tiff library from github.com/chai2010/tiff

----

TIFF for Go
===========

[![Build Status](https://travis-ci.org/chai2010/tiff.svg)](https://travis-ci.org/chai2010/tiff)
[![GoDoc](https://godoc.org/github.com/chai2010/tiff?status.svg)](https://godoc.org/github.com/chai2010/tiff)

**Features:**

1. Support BigTiff
2. Support decode multiple image
3. Support decode subifd image
4. Support RGB format
5. Support Float DataType
6. More ...

Install
=======

1. `go get github.com/dhushon/tiff`
2. `cd examples\hello`
3. `go run hello.go`

Example
=======

```Go
package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"

	tiff "github.com/dhushon/tiff"
)

var files = []string{
	"./testdata/BigTIFFSamples/BigTIFFSubIFD8.tif",
	"./testdata/multipage/multipage-gopher.tif",
}

func main() {
	for _, name := range files {
		// Load file data
		f, err := os.Open(name)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()

		// Decode tiff
		m, errors, err := tiff.DecodeAll(f)
		if err != nil {
			log.Println(err)
		}

		// Encode tiff
		for i := 0; i < len(m); i++ {
			for j := 0; j < len(m[i]); j++ {
				newname := fmt.Sprintf("%s-%02d-%02d.tiff", filepath.Base(name), i, j)
				if errors[i][j] != nil {
					log.Printf("%s: %v\n", newname, err)
					continue
				}

				var buf bytes.Buffer
				if err = tiff.Encode(&buf, m[i][j], nil); err != nil {
					log.Fatal(err)
				}
				if err = os.WriteFile(newname, buf.Bytes(), 0666); err != nil {
					log.Fatal(err)
				}
				fmt.Printf("Save %s ok\n", newname)
			}
		}
	}
}
```

BUGS
====

Report bugs to <chaishushan@gmail.com>.
or to <hushon@gmail.com>

Thanks!
