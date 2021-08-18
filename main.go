package main

import (
	"archive/zip"
	"crypto/rand"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strconv"
	"flag"
)

var (
	logger = log.New(os.Stderr, "", 0)
	outDir = "."
)

func init() {
	flag.Usage = func() {
		prog := path.Base(os.Args[0])
		out := func (format string, a ...interface{}) {
			fmt.Fprintf(flag.CommandLine.Output(), format+"\n", a...)
		}
		out("Usage of %s", prog)
		out("\t%s [outDir]", prog)
		out("Args:")
		out("\toutDir\t output root directory. if this argument is set, all files and directories will be extracted to the specified directory. (default: \".\"")
		out("Example:")
		out("\t$ echo foo > bar.txt")
		out("\t$ zip baz.zip bar.txt")
		out("\t$ cat baz.zip | %s extracted", prog)
		out("\t$ ls extracted")
		out("\tbar.txt")
		out("\t$ cat extracted/bar.txt")
		out("\tfoo")
	}
	flag.Parse()
	args := flag.Args()
	if len(args) >= 1 {
		outDir = args[0]
		if err := os.MkdirAll(outDir, 0755); err != nil {
			Fatal("can't create output directory", err)
		}
	}
}

func TempFile(size uint) (*os.File, error) {
	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		return nil, err
	}
	tmpFile, err := ioutil.TempFile("", fmt.Sprintf("%x-*", buf))
	if err != nil {
		return nil, err
	}
	return tmpFile, nil
}

func main() {
	tmpFile, err := TempFile(32)
	if err != nil {
		Fatal("can't create temporary zip file", err)
	}
	defer func() {
		os.Remove(tmpFile.Name())
	}()

	s, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		Fatal("can't read zip file", err)
	}
	if _, err := tmpFile.Write(s); err != nil {
		Fatal("can't write to temporary zip file", err)
	}

	r, err := zip.OpenReader(tmpFile.Name())
	if err != nil {
		Fatal("can't read from temporary zip file", err)
	}

	for _, f := range r.File {
		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(path.Join(outDir, f.Name), f.Mode()); err != nil {
				Warn("can't make directory %s", err, strconv.Quote(f.Name))
			} else {
				fmt.Println(f.Name)
			}
			continue
		}

		r, err := f.Open()
		if err != nil {
			Warn("can't open file %s in zip file", err, strconv.Quote(f.Name))
			continue
		}

		if err := os.MkdirAll(path.Join(outDir, path.Dir(f.Name)), 0755); err != nil {
			Warn("can't make directory %s", err, strconv.Quote(path.Dir(f.Name)))
		}
		w, err := os.OpenFile(path.Join(outDir, f.Name), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			Warn("can't open file %s", err, strconv.Quote(f.Name))
			continue
		}
		if _, err := io.Copy(w, r); err != nil {
			Warn("can't copy file contents: %s", err, strconv.Quote(f.Name))
			continue
		}
		fmt.Println(f.Name)
	}
}
