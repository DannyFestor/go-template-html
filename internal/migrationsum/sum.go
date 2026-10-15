package migrationsum

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io/fs"
	"strings"
)

const (
	FileName      = "migrations.sum"
	migrationGlob = "*.sql"
	hashPrefix    = "h1:"
)

type fileHash struct {
	name string
	hash [sha256.Size]byte
}

func (f fileHash) String() string {
	return f.name + " " + encodeHash(f.hash)
}

type Sum struct {
	total [sha256.Size]byte
	files []fileHash
}

func Compute(fsys fs.FS) (Sum, error) {
	names, err := fs.Glob(fsys, migrationGlob)
	if err != nil {
		return Sum{}, fmt.Errorf("list the migrations: %w", err)
	}

	files := make([]fileHash, 0, len(names))

	for _, name := range names {
		content, err := fs.ReadFile(fsys, name)
		if err != nil {
			return Sum{}, fmt.Errorf("read migration %s: %w", name, err)
		}

		files = append(files, fileHash{name: name, hash: sha256.Sum256(content)})
	}

	return Sum{total: totalOf(files), files: files}, nil
}

func totalOf(files []fileHash) [sha256.Size]byte {
	total := sha256.New()
	for _, file := range files {
		total.Write([]byte(file.name))
		total.Write(file.hash[:])
	}

	return [sha256.Size]byte(total.Sum(nil))
}

func (s Sum) Total() [sha256.Size]byte {
	return s.total
}

func (s Sum) String() string {
	var lines strings.Builder

	lines.WriteString(encodeHash(s.total) + "\n")

	for _, file := range s.files {
		lines.WriteString(file.String() + "\n")
	}

	return lines.String()
}

func encodeHash(hash [sha256.Size]byte) string {
	return hashPrefix + base64.StdEncoding.EncodeToString(hash[:])
}
