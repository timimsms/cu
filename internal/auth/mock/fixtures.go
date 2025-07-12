// Package mock provides test fixtures for authentication testing
package mock

import (
	"errors"
	"fmt"
	"time"

	"github.com/tim/cu/internal/auth"
	cuerrors "github.com/tim/cu/internal/errors"
)

// Common test tokens
const (
	// ValidToken represents a valid API token
	ValidToken = "pk_12345678_ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890" // #nosec G101 - Test fixture
	
	// ExpiredToken represents an expired API token
	ExpiredToken = "pk_87654321_ZYXWVUTSRQPONMLKJIHGFEDCBA0987654321" // #nosec G101 - Test fixture
	
	// InvalidToken represents a malformed token
	InvalidToken = "invalid_token_format"
	
	// LegacyToken represents a legacy format token (plain string)
	LegacyToken = "1234567890abcdef" // #nosec G101 - Test fixture
	
	// RefreshToken represents a token used for refresh scenarios
	RefreshToken = "pk_refresh_NEWTOKEN1234567890ABCDEFGHIJKLMNOP" // #nosec G101 - This is a test fixture, not a real credential
)

// Common test workspaces
const (
	// DefaultWorkspace is the default workspace name
	DefaultWorkspace = "default"
	
	// TestWorkspace is a test workspace name
	TestWorkspace = "test-workspace"
	
	// ProductionWorkspace is a production workspace name
	ProductionWorkspace = "production"
	
	// StagingWorkspace is a staging workspace name
	StagingWorkspace = "staging"
)

// Common test emails
const (
	// TestEmail is a test user email
	TestEmail = "test@example.com"
	
	// AdminEmail is an admin user email
	AdminEmail = "admin@example.com"
)

// TokenFixtures provides pre-configured tokens for testing
var TokenFixtures = struct {
	Valid      *auth.Token
	WithEmail  *auth.Token
	Legacy     *auth.Token
	Production *auth.Token
	Staging    *auth.Token
}{
	Valid: &auth.Token{
		Value:     ValidToken,
		Workspace: DefaultWorkspace,
	},
	WithEmail: &auth.Token{
		Value:     ValidToken,
		Workspace: DefaultWorkspace,
		Email:     TestEmail,
	},
	Legacy: &auth.Token{
		Value:     LegacyToken,
		Workspace: DefaultWorkspace,
	},
	Production: &auth.Token{
		Value:     ValidToken,
		Workspace: ProductionWorkspace,
		Email:     AdminEmail,
	},
	Staging: &auth.Token{
		Value:     ValidToken,
		Workspace: StagingWorkspace,
		Email:     TestEmail,
	},
}

// Scenarios provides pre-configured auth scenarios
type Scenarios struct {
	provider *AuthProvider
}

// NewScenarios creates a new scenarios helper
func NewScenarios(provider *AuthProvider) *Scenarios {
	return &Scenarios{provider: provider}
}

// NotAuthenticated sets up a scenario with no authentication
func (s *Scenarios) NotAuthenticated() *AuthProvider {
	s.provider.Reset()
	return s.provider
}

// Authenticated sets up a scenario with valid authentication
func (s *Scenarios) Authenticated() *AuthProvider {
	s.provider.Reset()
	s.provider.SetToken(DefaultWorkspace, ValidToken, time.Time{})
	return s.provider
}

// AuthenticatedWithEmail sets up authentication with email
func (s *Scenarios) AuthenticatedWithEmail() *AuthProvider {
	s.provider.Reset()
	s.provider.SetTokenWithEmail(DefaultWorkspace, ValidToken, TestEmail)
	return s.provider
}

// MultipleWorkspaces sets up multiple authenticated workspaces
func (s *Scenarios) MultipleWorkspaces() *AuthProvider {
	s.provider.Reset()
	s.provider.SetTokenWithEmail(DefaultWorkspace, ValidToken, TestEmail)
	s.provider.SetTokenWithEmail(ProductionWorkspace, ValidToken, AdminEmail)
	s.provider.SetTokenWithEmail(StagingWorkspace, ValidToken, TestEmail)
	return s.provider
}

// ExpiredToken sets up a scenario with an expired token
func (s *Scenarios) ExpiredToken() *AuthProvider {
	s.provider.Reset()
	s.provider.SetToken(DefaultWorkspace, ExpiredToken, time.Now().Add(-1*time.Hour))
	return s.provider
}

// ExpiredWithRefresh sets up expired token with refresh behavior
func (s *Scenarios) ExpiredWithRefresh() *AuthProvider {
	s.provider.Reset()
	s.provider.SetToken(DefaultWorkspace, ExpiredToken, time.Now().Add(-1*time.Hour))
	
	// Set up refresh behavior
	s.provider.SetRefreshBehavior(func(workspace string) (*auth.Token, error) {
		return &auth.Token{
			Value:     RefreshToken,
			Workspace: workspace,
			Email:     TestEmail,
		}, nil
	})
	
	return s.provider
}

// NetworkError sets up a scenario with network errors
func (s *Scenarios) NetworkError() *AuthProvider {
	s.provider.Reset()
	s.provider.SetGetError(errors.New("network error: connection timeout"))
	return s.provider
}

// KeyringError sets up a scenario with keyring errors
func (s *Scenarios) KeyringError() *AuthProvider {
	s.provider.Reset()
	s.provider.SetSaveError(errors.New("keyring error: access denied"))
	s.provider.SetGetError(errors.New("keyring error: access denied"))
	return s.provider
}

// PartialError sets up specific workspace errors
func (s *Scenarios) PartialError() *AuthProvider {
	s.provider.Reset()
	s.provider.SetToken(DefaultWorkspace, ValidToken, time.Time{})
	s.provider.SetError(ProductionWorkspace, errors.New("production access denied"))
	return s.provider
}

// InvalidToken sets up a scenario with invalid token format
func (s *Scenarios) InvalidToken() *AuthProvider {
	s.provider.Reset()
	s.provider.SetToken(DefaultWorkspace, InvalidToken, time.Time{})
	s.provider.SetError(DefaultWorkspace, cuerrors.ErrInvalidToken)
	return s.provider
}

// LegacyFormat sets up a scenario with legacy token format
func (s *Scenarios) LegacyFormat() *AuthProvider {
	s.provider.Reset()
	s.provider.SetToken(DefaultWorkspace, LegacyToken, time.Time{})
	return s.provider
}

// TestScenario represents a test scenario configuration
type TestScenario struct {
	Name        string
	Description string
	Setup       func(*AuthProvider)
	Validate    func(*AuthProvider) error
}

// CommonScenarios provides a set of common test scenarios
var CommonScenarios = []TestScenario{
	{
		Name:        "not_authenticated",
		Description: "No authentication tokens present",
		Setup: func(p *AuthProvider) {
			p.Reset()
		},
		Validate: func(p *AuthProvider) error {
			if p.IsAuthenticated(DefaultWorkspace) {
				return errors.New("expected not authenticated")
			}
			return nil
		},
	},
	{
		Name:        "valid_authentication",
		Description: "Valid token in default workspace",
		Setup: func(p *AuthProvider) {
			p.Reset()
			p.SetToken(DefaultWorkspace, ValidToken, time.Time{})
		},
		Validate: func(p *AuthProvider) error {
			if !p.IsAuthenticated(DefaultWorkspace) {
				return errors.New("expected authenticated")
			}
			token, err := p.GetToken(DefaultWorkspace)
			if err != nil {
				return err
			}
			if token.Value != ValidToken {
				return errors.New("unexpected token value")
			}
			return nil
		},
	},
	{
		Name:        "expired_token",
		Description: "Token has expired",
		Setup: func(p *AuthProvider) {
			p.Reset()
			p.SetToken(DefaultWorkspace, ExpiredToken, time.Now().Add(-1*time.Hour))
		},
		Validate: func(p *AuthProvider) error {
			_, err := p.GetToken(DefaultWorkspace)
			if err != cuerrors.ErrTokenExpired {
				return errors.New("expected token expired error")
			}
			return nil
		},
	},
	{
		Name:        "multiple_workspaces",
		Description: "Multiple workspaces with different tokens",
		Setup: func(p *AuthProvider) {
			p.Reset()
			p.SetTokenWithEmail(DefaultWorkspace, ValidToken, TestEmail)
			p.SetTokenWithEmail(ProductionWorkspace, ValidToken, AdminEmail)
			p.SetTokenWithEmail(StagingWorkspace, ValidToken, TestEmail)
		},
		Validate: func(p *AuthProvider) error {
			workspaces, err := p.ListWorkspaces()
			if err != nil {
				return err
			}
			if len(workspaces) != 3 {
				return errors.New("expected 3 workspaces")
			}
			
			// Check each workspace
			for _, ws := range []string{DefaultWorkspace, ProductionWorkspace, StagingWorkspace} {
				if !p.IsAuthenticated(ws) {
					return fmt.Errorf("workspace %s not authenticated", ws)
				}
			}
			
			// Check emails
			prodToken, _ := p.GetToken(ProductionWorkspace)
			if prodToken.Email != AdminEmail {
				return errors.New("unexpected email for production workspace")
			}
			
			return nil
		},
	},
}

// ErrorScenarios provides common error scenarios
var ErrorScenarios = struct {
	NetworkTimeout    error
	KeyringAccess     error
	InvalidToken      error
	TokenExpired      error
	NotAuthenticated  error
	PermissionDenied  error
}{
	NetworkTimeout:   errors.New("network error: connection timeout"),
	KeyringAccess:    errors.New("keyring error: access denied"),
	InvalidToken:     cuerrors.ErrInvalidToken,
	TokenExpired:     cuerrors.ErrTokenExpired,
	NotAuthenticated: cuerrors.ErrNotAuthenticated,
	PermissionDenied: errors.New("permission denied: insufficient privileges"),
}