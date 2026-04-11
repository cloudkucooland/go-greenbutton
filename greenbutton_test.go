package greenbutton

import (
	// "context"
	// "errors"
	"fmt"
	"os"
	// "strings"
	"testing"
	"time"
)

type testfiles struct {
	path       string
	shouldFail bool
}

var tf = []testfiles{
	{"tests/good.xml", false}, {"tests/big.xml", false}, {"tests/daily.xml", false},
}

func TestStaticParse(t *testing.T) {
	// func Parse(r io.Reader, callback func(EnergyReading)) error
	for _, tt := range tf {
		t.Run(tt.path, func(t *testing.T) {
			reader, err := os.Open(tt.path)
			err = Parse(reader, func(enre EnergyReading) {
				prefix := " [IMPORT]"
				if enre.IsExport {
					prefix = " [EXPORT]"
				}
				// Format the time to something human-friendly (RFC822 is short and clean)
				fmt.Printf("%s %s: %.2f Wh\n", prefix, enre.Start.Format(time.RFC822), enre.Value)
			})

			if !tt.shouldFail && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.shouldFail && err == nil {
				t.Fatalf("should have failed")
			}
		})
	}
}
