// Copyright (c) 2015 Pagoda Box Inc
//
// This Source Code Form is subject to the terms of the Mozilla Public License, v.
// 2.0. If a copy of the MPL was not distributed with this file, You can obtain one
// at http://mozilla.org/MPL/2.0/.
//

//
package file

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"github.com/nanobox-io/nanobox/config"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// Tar takes a source and variable writers and walks 'source' writing each file
// found to the tar writer; the purpose for accepting multiple writers is to allow
// for multiple outputs (for example a file, or md5 hash)
func Tar(src string, writers ...io.Writer) error {

	mw := io.MultiWriter(writers...)

	gzw := gzip.NewWriter(mw)
	defer gzw.Close()

	tw := tar.NewWriter(gzw)
	defer tw.Close()

	// walk path
	return filepath.Walk(src, func(file string, fi os.FileInfo, err error) error {

		if err != nil {
			return err
		}

		// only tar files (not dirs)
		if fi.Mode().IsRegular() {

			header := &tar.Header{
				Name: strings.TrimPrefix(strings.Replace(file, src, "", -1), string(filepath.Separator)),
				Mode: int64(fi.Mode()),
				Size: fi.Size(),
			}

			// write the header to the tarball archive
			if err := tw.WriteHeader(header); err != nil {
				return err
			}

			// open the file for taring...
			f, err := os.Open(file)
			defer f.Close()
			if err != nil {
				return err
			}

			// copy from file data into tar writer
			if _, err := io.Copy(tw, f); err != nil {
				return err
			}
		}

		return nil
	})
}

// Untar takes a destination path and a reader; a tar reader loops over the tarfile
// creating the file structure at 'dst' along the way, and writing any files
func Untar(dst string, r io.Reader) error {

	gzr, err := gzip.NewReader(r)
	defer gzr.Close()
	if err != nil {
		return err
	}

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()

		switch {

		// if no more files are found return
		case err == io.EOF:
			return nil

		// return any other error
		case err != nil:
			return err

		// if the header is nil, just skip it
		case header == nil:
			continue
		}

		dir := filepath.Dir(header.Name)
		base := filepath.Base(header.Name)
		path := filepath.Join(dst, dir)

		// check the file type
		switch header.Typeflag {

		// if its a dir and it doesn't exist create it
		case tar.TypeDir:
			if _, err := os.Stat(path); err != nil {
				if err := os.MkdirAll(path, 0755); err != nil {
					return err
				}
			}

		// if it's a file create it
		case tar.TypeReg:
			f, err := os.OpenFile(filepath.Join(path, base), os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			defer f.Close()

			// copy over contents
			if _, err := io.Copy(f, tr); err != nil {
				return err
			}
		}
	}
}

// Download downloads a file
func Download(path string, w io.Writer) error {
	res, err := http.Get(path)
	defer res.Body.Close()
	if err != nil {
		return err
	}

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		config.Fatal("[util/file/file] ioutil.ReadAll() failed - ", err.Error())
	}

	w.Write(b)

	return nil
}

// Progress downloads a file with a fancy progress bar
func Progress(path string, w io.Writer) error {

	//
	download, err := http.Get(path)
	defer download.Body.Close()
	if err != nil {
		return err
	}

	var percent float64
	var down int

	// format the response content length to be more 'friendly'
	total := float64(download.ContentLength) / math.Pow(1024, 2)

	// create a 'buffer' to read into
	p := make([]byte, 2048)

	//
	for {

		// read the response body (streaming)
		n, err := download.Body.Read(p)

		// write to our buffer
		w.Write(p[:n])

		// update the total bytes read
		down += n

		// update the percent downloaded
		percent = (float64(down) / float64(download.ContentLength)) * 100

		// show download progress: down/totalMB [*** progress *** %]
		fmt.Printf("\r   %.2f/%.2fMB [%-41s %.2f%%]", float64(down)/math.Pow(1024, 2), total, strings.Repeat("*", int(percent/2.5)), percent)

		// detect EOF and break the 'stream'
		if err != nil {
			if err == io.EOF {
				fmt.Println("")
				break
			} else {
				return err
			}
		}
	}

	return nil
}
