package appkey

type Purpose string

const (
	PurposeTOTPSecret         Purpose = "totp-secret"
	PurposeRecoveryCode       Purpose = "recovery-code"
	PurposeRecoveryCodeDigest Purpose = "recovery-code-digest"
	PurposeSignedURL          Purpose = "signed-url"
	//nolint:gosec // A derivation label, not a credential.
	PurposeCredentialFingerprint Purpose = "credential-fingerprint"
)

func Purposes() []Purpose {
	return []Purpose{
		PurposeTOTPSecret,
		PurposeRecoveryCode,
		PurposeRecoveryCodeDigest,
		PurposeSignedURL,
		PurposeCredentialFingerprint,
	}
}
