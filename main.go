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
)

var logger = log.New(os.Stderr, "", 0)

func exit(msg interface{}) {
	logger.Println(msg)
	os.Exit(1)
}

func printWarning(msg interface{}) {
	logger.Printf("warning: %v\n", msg)
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
		exit(fmt.Sprintf("can't create temporary zip file (reason: %s)", strconv.Quote(err.Error())))
	}
	defer func() {
		os.Remove(tmpFile.Name())
	}()

	s, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		exit(err)
	}
	if _, err := tmpFile.Write(s); err != nil {
		exit(fmt.Sprintf("can't write to temporary zip file (reason: %s)", strconv.Quote(err.Error())))
	}

	r, err := zip.OpenReader(tmpFile.Name())
	if err != nil {
		exit(fmt.Sprintf("can't read from temporary zip file (reason: %s)", strconv.Quote(err.Error())))
	}

	for _, f := range r.File {
		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(path.Join(".", f.Name), f.Mode()); err != nil {
				printWarning(fmt.Sprintf("can't make directory %s (reason: %s)", strconv.Quote(f.Name), strconv.Quote(err.Error())))
			} else {
				fmt.Println(f.Name)
			}
			continue
		}

		r, err := f.Open()
		if err != nil {
			printWarning(fmt.Sprintf("can't open file %s in zip file (reason: %s)", strconv.Quote(f.Name), strconv.Quote(err.Error())))
			continue
		}

		if err := os.MkdirAll(path.Join(".", path.Dir(f.Name)), os.ModeDir); err != nil {
			printWarning(fmt.Sprintf("can't make directory %s (reason: %s)", strconv.Quote(path.Dir(f.Name)), strconv.Quote(err.Error())))
		}
		w, err := os.OpenFile(path.Join(".", f.Name), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			printWarning(fmt.Sprintf("can't open file %s (reason: %s)", strconv.Quote(f.Name), strconv.Quote(err.Error())))
			continue
		}
		if _, err := io.Copy(w, r); err != nil {
			printWarning(fmt.Sprintf("can't copy file contents: %s (reason: %s)", strconv.Quote(f.Name), strconv.Quote(err.Error())))
			continue
		}
		fmt.Println(f.Name)
	}
}
