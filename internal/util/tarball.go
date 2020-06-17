package util

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
)

// From https://www.arthurkoziel.com/writing-tar-gz-files-in-go/
func TarFiles(filename string, files []string, preserveDir bool) error {
	// Create output file
	out, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer out.Close()

	// Create the archive and write the output to the "out" Writer
	err = createArchive(files, out, preserveDir)
	if err != nil {
		return err
	}

	return nil
}

func createArchive(files []string, buf io.Writer, preserveDir bool) error {
	// Create new Writers for gzip and tar
	// These writers are chained. Writing to the tar writer will
	// write to the gzip writer which in turn will write to
	// the "buf" writer
	gw := gzip.NewWriter(buf)
	defer gw.Close()
	tw := tar.NewWriter(gw)
	defer tw.Close()

	// Iterate over files and add them to the tar archive
	for _, file := range files {
		err := addToArchive(tw, file, preserveDir)
		if err != nil {
			return err
		}
	}

	return nil
}

func addToArchive(tw *tar.Writer, filename string, preserveDir bool) error {
	// Open the file which will be written into the archive
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Get FileInfo about our file providing file size, mode, etc.
	info, err := file.Stat()
	if err != nil {
		return err
	}

	// Create a tar Header from the FileInfo data
	header, err := tar.FileInfoHeader(info, info.Name())
	if err != nil {
		return err
	}

	// Use full path as name (FileInfoHeader only takes the basename)
	// If we don't do this the directory structure would
	// not be preserved
	// https://golang.org/src/archive/tar/common.go?#L626
	if preserveDir {
		header.Name = filename
	}

	// Write file header to the tar archive
	err = tw.WriteHeader(header)
	if err != nil {
		return err
	}

	// Copy file content to tar archive
	_, err = io.Copy(tw, file)
	if err != nil {
		return err
	}

	return nil
}
