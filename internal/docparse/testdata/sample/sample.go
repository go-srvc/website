// Package sample is a small fixture used by docparse tests.
package sample

import "fmt"

// Pi is a constant.
const Pi = 3.14

// Author is a variable.
var Author = "go-srvc"

// Greet returns a greeting addressed to name.
func Greet(name string) string {
	return fmt.Sprintf("hello, %s", name)
}

// Counter counts things.
type Counter struct {
	n int
}

// Inc increments the counter.
func (c *Counter) Inc() { c.n++ }

// Value returns the current count.
func (c *Counter) Value() int { return c.n }
