// Command stdfdtest is a part of the unit tests for stdfd.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/facebookgo/stdfd"
)

func main() {
	out := flag.String("out", "", "out text")
	err := flag.String("err", "", "err text")
	stdout := flag.String("stdout", "", "stdout path")
	stderr := flag.String("stderr", "", "stderr path")

	flag.Parse()
	stdfd.RedirectOutputs(*stdout, *stderr)

	fmt.Fprint(os.Stdout, *out)
	fmt.Fprint(os.Stderr, *err)
}
