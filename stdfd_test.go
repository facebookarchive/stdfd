package stdfd_test

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/facebookgo/stdfd"
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

const customMain = "CUSTOM_MAIN"

var customMainEnv = []string{customMain + "=1"}

func TestMain(m *testing.M) {
	if os.Getenv(customMain) == "" {
		os.Exit(m.Run())
	}

	out := flag.String("out", "", "out text")
	err := flag.String("err", "", "err text")
	stdout := flag.String("stdout", "", "stdout path")
	stderr := flag.String("stderr", "", "stderr path")
	flag.Parse()
	stdfd.RedirectOutputs(*stdout, *stderr)
	fmt.Fprint(os.Stdout, *out)
	fmt.Fprint(os.Stderr, *err)
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
		os.Args[0],
		"-stdout", c.outpath,
		"-stderr", c.errpath,
		"-out", outText,
		"-err", errText,
	)
	cmd.Env = customMainEnv
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

func TestCreateDirectory(t *testing.T) {
	t.Parallel()
	run(t, Case{
		outpath: filepath.Join(tempfilename(t, "-arg-stdout"), "out"),
		errpath: filepath.Join(tempfilename(t, "-arg-stderr"), "err"),
	})
}
