package stdfd_test

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"testing"

	"github.com/daaku/go.tool"
)

// Make a temporary file.
func tempfile(t *testing.T, suffix string) *os.File {
	file, err := ioutil.TempFile("", "stdfd"+suffix)
	if err != nil {
		t.Fatal(err)
	}
	return file
}

// Make a temporary file, remove it, and return it's path with the hopes that
// no one else create a file with that name.
func tempfilename(t *testing.T, suffix string) string {
	file := tempfile(t, suffix)

	err := file.Close()
	if err != nil {
		t.Fatal(err)
	}

	err = os.Remove(file.Name())
	if err != nil {
		t.Fatal(err)
	}

	return file.Name()
}

func checkcontents(t *testing.T, path string, content string) {
	d, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(d) != content {
		t.Fatalf(`expected "%s" to contain "%s" but got "%s"`, path, content, string(d))
	}
}

// Things to ensure we only build the binary once.
var (
	binOnce sync.Once
	binPath string
)

// Build the test command and return the binary's path.
func bin(t *testing.T) string {
	binOnce.Do(func() {
		const testcmd = "github.com/daaku/go.stdfd/stdfdtest"
		binPath = tempfilename(t, filepath.Base(testcmd)+"-bin-")
		options := tool.Options{
			ImportPaths: []string{testcmd},
			Output:      binPath,
		}
		_, err := options.Command("build")
		if err != nil {
			t.Fatal(err)
		}
	})
	return binPath
}

type Case struct {
	outpath string
	errpath string
}

func run(t *testing.T, c Case) {
	const outText = "out"
	const errText = "err"

	givenStdout := tempfile(t, "-given-stdout")
	givenStderr := tempfile(t, "-given-stderr")
	defer givenStdout.Close()
	defer givenStderr.Close()
	defer os.Remove(givenStdout.Name())
	defer os.Remove(givenStderr.Name())

	if c.outpath != "" {
		defer os.Remove(c.outpath)
	}
	if c.errpath != "" {
		defer os.Remove(c.errpath)
	}

	cmd := exec.Command(
		bin(t),
		"-stdout", c.outpath,
		"-stderr", c.errpath,
		"-out", outText,
		"-err", errText,
	)
	cmd.Stdout = givenStdout
	cmd.Stderr = givenStderr
	if err := cmd.Run(); err != nil {
		t.Fatal(err)
	}

	if c.outpath != "" && c.outpath == c.errpath {
		checkcontents(t, c.outpath, outText+errText)
		return
	}

	if c.outpath == "" {
		checkcontents(t, givenStdout.Name(), outText)
	} else {
		checkcontents(t, c.outpath, outText)
	}

	if c.errpath == "" {
		checkcontents(t, givenStderr.Name(), errText)
	} else {
		checkcontents(t, c.errpath, errText)
	}
}

func TestOutOnly(t *testing.T) {
	t.Parallel()
	run(t, Case{outpath: tempfilename(t, "-arg-stdout")})
}

func TestErrOnly(t *testing.T) {
	t.Parallel()
	run(t, Case{errpath: tempfilename(t, "-arg-stderr")})
}

func TestBothSame(t *testing.T) {
	t.Parallel()
	p := tempfilename(t, "-arg-stderr")
	run(t, Case{outpath: p, errpath: p})
}

func TestBothDifferent(t *testing.T) {
	t.Parallel()
	run(t, Case{
		outpath: tempfilename(t, "-arg-stdout"),
		errpath: tempfilename(t, "-arg-stderr"),
	})
}
