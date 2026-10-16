package argon2id

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	phcPrefix        = "$argon2id$"
	phcVersionFormat = "v=%d"
	phcParamsFormat  = "m=%d,t=%d,p=%d"
)

const (
	phcVersionField = iota + 2
	phcParamsField
	phcSaltField
	phcKeyField
	phcFieldCount
)

var (
	errNotArgon2id         = errors.New("not an argon2id PHC string")
	errUnsupportedVersion  = errors.New("unsupported argon2 version")
	errUnparseableEncoding = errors.New("unparseable argon2id PHC string")
)

type phcHash struct {
	params Params
	salt   []byte
	key    []byte
}

func (h phcHash) String() string {
	return phcPrefix +
		fmt.Sprintf(phcVersionFormat, argon2.Version) + "$" +
		fmt.Sprintf(phcParamsFormat, h.params.MemoryKiB, h.params.Iterations, h.params.Parallelism) + "$" +
		base64.RawStdEncoding.EncodeToString(h.salt) + "$" +
		base64.RawStdEncoding.EncodeToString(h.key)
}

func parsePHCHash(encoded string) (phcHash, error) {
	fields := strings.Split(encoded, "$")
	if len(fields) != phcFieldCount || !strings.HasPrefix(encoded, phcPrefix) {
		return phcHash{}, errNotArgon2id
	}

	err := checkVersion(fields[phcVersionField])
	if err != nil {
		return phcHash{}, err
	}

	params, err := decodeParams(fields[phcParamsField])
	if err != nil {
		return phcHash{}, err
	}

	return decodeSaltAndKey(params, fields[phcSaltField], fields[phcKeyField])
}

func checkVersion(field string) error {
	var version int

	_, err := fmt.Sscanf(field, phcVersionFormat, &version)
	if err != nil {
		return fmt.Errorf("%w: version %q", errUnparseableEncoding, field)
	}

	if version != argon2.Version {
		return fmt.Errorf("%w: %d", errUnsupportedVersion, version)
	}

	return nil
}

func decodeParams(field string) (Params, error) {
	var params Params

	_, err := fmt.Sscanf(field, phcParamsFormat, &params.MemoryKiB, &params.Iterations, &params.Parallelism)
	if err != nil {
		return Params{}, fmt.Errorf("%w: params %q", errUnparseableEncoding, field)
	}

	err = params.checkPositiveCost()
	if err != nil {
		return Params{}, fmt.Errorf("%w: %w", errUnparseableEncoding, err)
	}

	return params, nil
}

func decodeSaltAndKey(params Params, saltField, keyField string) (phcHash, error) {
	salt, err := base64.RawStdEncoding.DecodeString(saltField)
	if err != nil {
		return phcHash{}, fmt.Errorf("%w: salt: %w", errUnparseableEncoding, err)
	}

	key, err := base64.RawStdEncoding.DecodeString(keyField)
	if err != nil {
		return phcHash{}, fmt.Errorf("%w: key: %w", errUnparseableEncoding, err)
	}

	if len(key) != rfc9106KeyBytes {
		return phcHash{}, fmt.Errorf("%w: a %d-byte key", errUnparseableEncoding, len(key))
	}

	return phcHash{params: params, salt: salt, key: key}, nil
}
