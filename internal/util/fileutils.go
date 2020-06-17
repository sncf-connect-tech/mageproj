package util

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"os/exec"
)

// FileDetails provides check sums and size for a given file
type FileDetails struct {
	Checksum ChecksumDetails
	Size     int64
}

// ChecksumDetails provides check sums
type ChecksumDetails struct {
	Md5    string
	Sha1   string
	Sha256 string
}

// GetFileDetails returns file details for a given file
func GetFileDetails(filePath string) (*FileDetails, error) {
	var err error

	details := new(FileDetails)
	details.Checksum, err = GetChecksumDetails(filePath)

	file, err := os.Open(filePath)
	defer file.Close()
	if err != nil {
		return nil, err
	}
	fileInfo, err := file.Stat()
	if err != nil {
		return nil, err
	}
	details.Size = fileInfo.Size()
	return details, nil
}

// GetChecksumDetails returns file checksums for a given file
func GetChecksumDetails(filePath string) (ChecksumDetails, error) {
	file, err := os.Open(filePath)
	defer file.Close()
	if err != nil {
		return ChecksumDetails{}, err
	}

	sha256, err := Sha256(filePath)
	if err != nil {
		return ChecksumDetails{}, err
	}
	sha1, err := Sha1(filePath)
	if err != nil {
		return ChecksumDetails{}, err
	}
	md5, err := Md5(filePath)
	if err != nil {
		return ChecksumDetails{}, err
	}

	return ChecksumDetails{Md5: md5, Sha1: sha1, Sha256: sha256}, nil
}

// Sha256 returns sha256sum for the given file
func Sha256(file string) (string, error) {
	hasher := sha256.New()
	err := sumFile(hasher, file)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

// Sha1 returns sha1sum for the given file
func Sha1(file string) (string, error) {
	hasher := sha1.New()
	err := sumFile(hasher, file)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

// Md5 returns md5sum for the given file
func Md5(file string) (string, error) {
	hasher := md5.New()
	err := sumFile(hasher, file)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func sumFile(hasher io.Writer, file string) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := io.Copy(hasher, f); err != nil {
		return err
	}
	return nil
}

// ExecOutput executes a command and returns its standard output
func ExecOutput(cmd string, args ...string) (string, error) {
	out, err := exec.Command(cmd, args...).Output()
	if err != nil {
		return "", err
	}
	if len(out) > 0 {
		return string(out), nil
	}
	return "", nil
}
