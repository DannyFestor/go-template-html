package postgres

import (
	"context"
	"fmt"
	"uuid"

	"github.com/DannyFestor/go-template-html/internal/adapter/postgres/internal/sqlcgen"
	"github.com/DannyFestor/go-template-html/internal/domain"
)

type digester interface {
	Digest(message []byte) []byte
}

type RecoveryCodeStore struct {
	pool      *Pool
	encrypter encrypter
	digester  digester
}

func NewRecoveryCodeStore(pool *Pool, encrypter encrypter, digester digester) *RecoveryCodeStore {
	return &RecoveryCodeStore{pool: pool, encrypter: encrypter, digester: digester}
}

func (s *RecoveryCodeStore) Add(ctx context.Context, userID domain.UserID, codes []domain.RecoveryCode) error {
	id, err := parseUserID(userID)
	if err != nil {
		return err
	}

	err = s.pool.queries(ctx).AddRecoveryCodes(ctx, s.addParams(id, codes))
	if err != nil {
		return fmt.Errorf("add the recovery codes of user %s: %w", userID, err)
	}

	return nil
}

func (s *RecoveryCodeStore) List(ctx context.Context, userID domain.UserID) ([]domain.RecoveryCode, error) {
	id, err := parseUserID(userID)
	if err != nil {
		return nil, err
	}

	encrypted, err := s.pool.queries(ctx).ListRecoveryCodes(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("list the recovery codes of user %s: %w", userID, err)
	}

	codes, err := s.decryptAll(encrypted)
	if err != nil {
		return nil, fmt.Errorf("read the recovery codes of user %s: %w", userID, err)
	}

	return codes, nil
}

func (s *RecoveryCodeStore) Consume(ctx context.Context, userID domain.UserID, code domain.RecoveryCode) (bool, error) {
	id, err := parseUserID(userID)
	if err != nil {
		return false, err
	}

	deleted, err := s.pool.queries(ctx).ConsumeRecoveryCode(ctx, sqlcgen.ConsumeRecoveryCodeParams{
		UserID: id,
		Digest: s.digester.Digest([]byte(code)),
	})
	if err != nil {
		return false, fmt.Errorf("consume a recovery code of user %s: %w", userID, err)
	}

	return deleted == 1, nil
}

func (s *RecoveryCodeStore) DeleteAll(ctx context.Context, userID domain.UserID) error {
	id, err := parseUserID(userID)
	if err != nil {
		return err
	}

	err = s.pool.queries(ctx).DeleteRecoveryCodes(ctx, id)
	if err != nil {
		return fmt.Errorf("delete the recovery codes of user %s: %w", userID, err)
	}

	return nil
}

func (s *RecoveryCodeStore) addParams(userID uuid.UUID, codes []domain.RecoveryCode) sqlcgen.AddRecoveryCodesParams {
	params := sqlcgen.AddRecoveryCodesParams{
		UserID:         userID,
		Digests:        make([][]byte, 0, len(codes)),
		EncryptedCodes: make([][]byte, 0, len(codes)),
	}

	for _, code := range codes {
		params.Digests = append(params.Digests, s.digester.Digest([]byte(code)))
		params.EncryptedCodes = append(params.EncryptedCodes, s.encrypter.Encrypt([]byte(code)))
	}

	return params
}

func (s *RecoveryCodeStore) decryptAll(encrypted [][]byte) ([]domain.RecoveryCode, error) {
	codes := make([]domain.RecoveryCode, 0, len(encrypted))

	for _, ciphertext := range encrypted {
		code, err := s.encrypter.Decrypt(ciphertext)
		if err != nil {
			return nil, fmt.Errorf("decrypt a recovery code: %w", err)
		}

		codes = append(codes, domain.RecoveryCode(code))
	}

	return codes, nil
}
