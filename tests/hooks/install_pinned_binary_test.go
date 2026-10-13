package hooks_test

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
	"time"
)

const (
	installScript  = "install-pinned-binary.sh"
	unreachableURL = "http://127.0.0.1:9/tool"
)

type pinnedAsset struct {
	path         string
	url          string
	checksumFile string
	destination  string
}

// newPinnedAsset serves a binary asset from a file:// URL, so an install needs no network.
func newPinnedAsset(t *testing.T) pinnedAsset {
	t.Helper()

	dir := t.TempDir()
	content := []byte("#!/bin/sh\necho " + t.Name() + "\n")
	writeFile(t, dir, "tool", string(content))

	sum := sha256.Sum256(content)
	writeFile(t, dir, "tool.sha256", hex.EncodeToString(sum[:])+"  tool\n")

	return pinnedAsset{
		path:         filepath.Join(dir, "tool"),
		url:          "file://" + filepath.Join(dir, "tool"),
		checksumFile: filepath.Join(dir, "tool.sha256"),
		destination:  filepath.Join(dir, "bin", "tool"),
	}
}

func (a pinnedAsset) install(t *testing.T, url string) scriptRun {
	t.Helper()

	return runScript(t, filepath.Dir(a.path), "", installScript, a.destination, url, a.checksumFile)
}

func TestInstallPinnedBinaryInstallsTheVerifiedAsset(t *testing.T) {
	t.Parallel()

	asset := newPinnedAsset(t)

	requirePassed(t, asset.install(t, asset.url))

	info, err := os.Stat(asset.destination)
	if err != nil || info.Mode().Perm()&0o100 == 0 {
		t.Fatalf("stat %s = %v, %v, want an executable", asset.destination, info, err)
	}
}

func TestInstallPinnedBinaryRefusesAnAssetWithAnotherChecksum(t *testing.T) {
	t.Parallel()

	asset := newPinnedAsset(t)
	writeFile(t, filepath.Dir(asset.path), "tool", "tampered\n")

	requireRefused(t, asset.install(t, asset.url), "checksum mismatch")
}

func TestInstallPinnedBinarySkipsTheDownloadWhenTheSamePinIsInstalled(t *testing.T) {
	t.Parallel()

	asset := newPinnedAsset(t)
	requirePassed(t, asset.install(t, asset.url))
	removeAsset(t, asset)
	makeOlderThanTheChecksumFile(t, asset)

	requirePassed(t, asset.install(t, asset.url))
	requireNewerThanTheChecksumFile(t, asset)
}

func TestInstallPinnedBinaryDownloadsAgainWhenThePinChanged(t *testing.T) {
	t.Parallel()

	asset := newPinnedAsset(t)
	requirePassed(t, asset.install(t, asset.url))

	requireRefused(t, asset.install(t, unreachableURL), "curl")
}

func removeAsset(t *testing.T, asset pinnedAsset) {
	t.Helper()

	err := os.Remove(asset.path)
	if err != nil {
		t.Fatalf("remove %s: %v", asset.path, err)
	}
}

// A restored CI cache holds binaries older than the freshly checked-out pins, which Make reads as stale.
func makeOlderThanTheChecksumFile(t *testing.T, asset pinnedAsset) {
	t.Helper()

	past := time.Now().Add(-time.Hour)

	err := os.Chtimes(asset.destination, past, past)
	if err != nil {
		t.Fatalf("age %s: %v", asset.destination, err)
	}
}

func requireNewerThanTheChecksumFile(t *testing.T, asset pinnedAsset) {
	t.Helper()

	binary, binaryErr := os.Stat(asset.destination)
	checksums, checksumsErr := os.Stat(asset.checksumFile)

	if binaryErr != nil || checksumsErr != nil || binary.ModTime().Before(checksums.ModTime()) {
		t.Errorf("binary is older than %s, so make would download it again", asset.checksumFile)
	}
}
