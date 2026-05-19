package sample_test

import (
	"fmt"

	"github.com/go-srvc/website/internal/docparse/testdata/sample"
)

func ExampleGreet() {
	fmt.Println(sample.Greet("world"))
	// Output: hello, world
}
