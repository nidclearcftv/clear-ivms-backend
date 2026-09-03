package service

import (
	"context"
	"fmt"

	"golang.org/x/crypto/bcrypt"

	"github.com/nidclearcftv/clear-ivms-backend/core/model"
	"github.com/nidclearcftv/clear-ivms-backend/core/port"
	"github.com/nidclearcftv/clear-ivms-backend/utils/validate"
)

type SeedOptions struct {
	Organizations port.OrganizationService `validate:"required"`
	Accounts      port.AccountService      `validate:"required"`

	// OrganizationName is the default organization created (if none with
	// this name exists yet) on Seed.
	OrganizationName string `validate:"required"`

	// AdminName/AdminEmail/AdminPassword seed the default admin account,
	// created (if AdminEmail isn't already registered) and attached to the
	// default organization. AdminPassword has no fallback default — a
	// hardcoded default admin password would be a real vulnerability the
	// moment this runs against a real database, so the deployer must set
	// one explicitly.
	AdminName     string `validate:"required"`
	AdminEmail    string `validate:"required,email"`
	AdminPassword string `validate:"required"`
}

// SeedService bootstraps a fresh database with a default organization and
// admin account. It's meant to run once at startup; every step is
// idempotent (safe to run again against an already-seeded database), since
// there's no other signal available for "has this already run".
type SeedService struct {
	organizations port.OrganizationService
	accounts      port.AccountService
	opts          SeedOptions
}

func NewSeedService(opts SeedOptions) (*SeedService, error) {
	if err := validate.Struct(opts); err != nil {
		return nil, err
	}

	return &SeedService{organizations: opts.Organizations, accounts: opts.Accounts, opts: opts}, nil
}

// Seed ensures the default organization and admin account exist, creating
// whichever are missing, and makes sure the admin belongs to the
// organization.
func (s *SeedService) Seed(ctx context.Context) error {
	org, err := s.findOrCreateOrganization(ctx)
	if err != nil {
		return fmt.Errorf("service: failed to seed default organization: %w", err)
	}

	admin, err := s.findOrCreateAdmin(ctx)
	if err != nil {
		return fmt.Errorf("service: failed to seed admin account: %w", err)
	}

	// AddOrganization is idempotent, so this is safe whether admin was
	// just created or already existed.
	if err := s.accounts.AddOrganization(ctx, admin.ID, org.ID); err != nil {
		return fmt.Errorf("service: failed to attach admin account to default organization: %w", err)
	}

	return nil
}

// findOrCreateOrganization matches by name: organizations.name has no
// unique constraint, so this is a client-side check rather than relying on
// a conflict error the way findOrCreateAdmin does with email.
func (s *SeedService) findOrCreateOrganization(ctx context.Context) (model.Organization, error) {
	organizations, err := s.organizations.List(ctx, model.OrganizationFilters{})
	if err != nil {
		return model.Organization{}, err
	}

	for _, o := range organizations.Items {
		if o.Name == s.opts.OrganizationName {
			return o, nil
		}
	}

	return s.organizations.Create(ctx, model.Organization{Name: s.opts.OrganizationName})
}

// findOrCreateAdmin matches by email, same as findOrCreateOrganization —
// kept consistent rather than relying on AccountService.Create's
// ErrCodeAccountEmailAlreadyExists, since that path can't hand back the
// existing account's ID that Seed needs for AddOrganization.
func (s *SeedService) findOrCreateAdmin(ctx context.Context) (model.Account, error) {
	accounts, err := s.accounts.List(ctx, model.AccountFilters{})
	if err != nil {
		return model.Account{}, err
	}

	for _, a := range accounts.Items {
		if a.Email == s.opts.AdminEmail {
			return a, nil
		}
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(s.opts.AdminPassword), bcrypt.DefaultCost)
	if err != nil {
		return model.Account{}, fmt.Errorf("failed to hash admin password: %w", err)
	}

	return s.accounts.Create(ctx, model.Account{
		Name:  s.opts.AdminName,
		Email: s.opts.AdminEmail,
		Type:  model.AccountTypeAdmin,
	}, string(passwordHash))
}
