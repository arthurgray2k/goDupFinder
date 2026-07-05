package hasher

import (
	"bytes"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"hash"
	"io"
	"os"

	"golang.org/x/crypto/blake2b"
)

// getHash returns a new hash.Hash based on the specified algorithm.
func getHash(algo string) (hash.Hash, error) {
	switch algo {
	case "sha256":
		return sha256.New(), nil
	case "sha1":
		return sha1.New(), nil
	case "md5":
		return md5.New(), nil
	case "blake2":
		h, err := blake2b.New256(nil)
		if err != nil {
			return nil, err
		}
		return h, nil
	default:
		return sha256.New(), nil
	}
}

// FullHash computes the full hash of the file using the specified algorithm.
func FullHash(path string, algo string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h, err := getHash(algo)
	if err != nil {
		return "", err
	}

	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

// PartialHash computes the hash of the first 'size' bytes of the file.
func PartialHash(path string, algo string, size int64) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h, err := getHash(algo)
	if err != nil {
		return "", err
	}

	if _, err := io.CopyN(h, f, size); err != nil {
		if err == io.EOF {
			// If file is smaller than 'size', that's fine, we hashed everything we could read
		} else {
			return "", err
		}
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

// CompareFiles compares two files byte-for-byte to confirm they are exactly identical.
func CompareFiles(path1, path2 string) (bool, error) {
	f1, err := os.Open(path1)
	if err != nil {
		return false, err
	}
	defer f1.Close()

	f2, err := os.Open(path2)
	if err != nil {
		return false, err
	}
	defer f2.Close()

	const chunkSize = 32 * 1024
	buf1 := make([]byte, chunkSize)
	buf2 := make([]byte, chunkSize)

	for {
		n1, err1 := io.ReadFull(f1, buf1)
		n2, err2 := io.ReadFull(f2, buf2)

		if err1 == io.ErrUnexpectedEOF || err1 == io.EOF {
			err1 = nil
		}
		if err2 == io.ErrUnexpectedEOF || err2 == io.EOF {
			err2 = nil
		}

		if err1 != nil {
			return false, err1
		}
		if err2 != nil {
			return false, err2
		}

		if n1 != n2 {
			return false, nil // different size? Should not happen if pre-checked, but just in case
		}
		if n1 == 0 {
			break // EOF reached
		}

		if !bytes.Equal(buf1[:n1], buf2[:n2]) {
			return false, nil
		}
	}

	return true, nil
}
