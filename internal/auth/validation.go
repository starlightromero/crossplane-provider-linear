package auth

import "fmt"

// ValidCredentialSources lists all supported credential source values.
var ValidCredentialSources = []CredentialSource{
	SourceSecret,
	SourceOAuth2ClientCredentials,
	SourceOAuth2,
}

// ValidateCredentialSource validates that the given source is a supported credential source.
// Returns an error if the source is unsupported.
func ValidateCredentialSource(source CredentialSource) error {
	switch source {
	case SourceSecret, SourceOAuth2ClientCredentials, SourceOAuth2:
		return nil
	default:
		return fmt.Errorf("unsupported credential source %q: must be one of Secret, OAuth2ClientCredentials, OAuth2", source)
	}
}

// ValidateCredentials validates that the Credentials struct specifies exactly one
// valid authentication method and has the required fields for that method.
func ValidateCredentials(creds *Credentials) error {
	if creds == nil {
		return fmt.Errorf("credentials must not be nil")
	}

	if err := ValidateCredentialSource(creds.Source); err != nil {
		return err
	}

	// All supported sources require a secret reference
	if creds.SecretRef.Name == "" {
		return fmt.Errorf("credentials source %q requires secretRef.name", creds.Source)
	}
	if creds.SecretRef.Namespace == "" {
		return fmt.Errorf("credentials source %q requires secretRef.namespace", creds.Source)
	}

	return nil
}

// NewTokenProvider creates the appropriate TokenProvider based on the credential source.
// It validates the credentials and returns an error if the configuration is invalid.
func NewTokenProvider(creds *Credentials, reader SecretReader, writer SecretWriter, exchanger HTTPTokenExchanger) (TokenProvider, error) {
	if err := ValidateCredentials(creds); err != nil {
		return nil, err
	}

	switch creds.Source {
	case SourceSecret:
		return NewSecretTokenProvider(reader, creds.SecretRef), nil
	case SourceOAuth2ClientCredentials:
		return NewClientCredentialsTokenProvider(reader, exchanger, creds.SecretRef, creds.Scope), nil
	case SourceOAuth2:
		return NewAuthorizationCodeTokenProvider(reader, writer, exchanger, creds.SecretRef), nil
	default:
		return nil, fmt.Errorf("unsupported credential source %q", creds.Source)
	}
}
