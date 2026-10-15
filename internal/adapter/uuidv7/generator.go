package uuidv7

import (
	"encoding/binary"
	"time"
	"uuid"
)

const (
	timestampBytes = 6
	versionIndex   = 6
	version7Bits   = 0x70
	versionMask    = 0x0f
)

type Generator struct {
	clock clock
}

func NewGenerator(clock clock) *Generator {
	return &Generator{clock: clock}
}

func (g *Generator) NewID() string {
	// The standard library's NewV7 reads the system clock, so the ID starts as a version 4 UUID, which already holds
	// the random bits and the variant, and only its timestamp and version are rewritten.
	id := uuid.NewV4()
	writeTimestamp(&id, g.clock.Now())
	markVersion7(&id)

	return id.String()
}

func writeTimestamp(id *uuid.UUID, now time.Time) {
	var milliseconds [8]byte

	binary.BigEndian.PutUint64(milliseconds[:], uint64(now.UnixMilli()))
	copy(id[:timestampBytes], milliseconds[len(milliseconds)-timestampBytes:])
}

func markVersion7(id *uuid.UUID) {
	id[versionIndex] = id[versionIndex]&versionMask | version7Bits
}
