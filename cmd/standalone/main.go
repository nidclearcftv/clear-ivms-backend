package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nidclearcftv/clear-ivms-backend/adapter/cache/memory"
	"github.com/nidclearcftv/clear-ivms-backend/adapter/db/postgres"
	httpapi "github.com/nidclearcftv/clear-ivms-backend/adapter/http"
	"github.com/nidclearcftv/clear-ivms-backend/adapter/recaptcha"
	"github.com/nidclearcftv/clear-ivms-backend/core/service"
	"github.com/nidclearcftv/clear-ivms-backend/utils/env"
	"github.com/nidclearcftv/clear-ivms-backend/utils/logger"
)

type Env struct {
	DatabaseURL            string `env:"DATABASE_URL,required=true"`
	DatabaseSchemaPath     string `env:"DATABASE_SCHEMA_PATH,default=adapter/db/postgres/sql/schema.sql"`
	DatabaseMigrationsPath string `env:"DATABASE_MIGRATIONS_PATH,default=adapter/db/postgres/sql/migrations"`

	HTTPAddr                 string   `env:"HTTP_ADDR,default=:8080"`
	HTTPAllowedOrigins       []string `env:"HTTP_ALLOWED_ORIGINS,separator=,"`
	HTTPAllowInsecureCookies bool     `env:"HTTP_ALLOW_INSECURE_COOKIES,default=false"`

	// SeedOrganizationName/SeedAdmin* bootstrap a default organization and
	// admin account on startup (see service.SeedService) — only run when
	// both SeedAdminEmail and SeedAdminPassword are set, so existing
	// deployments that don't want a seeded account are unaffected.
	SeedOrganizationName string `env:"SEED_ORGANIZATION_NAME,default="`
	SeedAdminName        string `env:"SEED_ADMIN_NAME,default="`
	SeedAdminEmail       string `env:"SEED_ADMIN_EMAIL,default="`
	SeedAdminPassword    string `env:"SEED_ADMIN_PASSWORD,default="`

	// RecaptchaEnabled gates /api/v1/login with reCAPTCHA verification —
	// see httpapi.RecaptchaOptions. Defaults to false so an existing
	// deployment upgrading to this version isn't suddenly locked out of
	// login by an unconfigured feature; set it to true only once both
	// secret keys below are also set.
	RecaptchaEnabled bool `env:"RECAPTCHA_ENABLED,default=false"`
	// RecaptchaV3SecretKey/RecaptchaV2SecretKey are the secret keys for
	// the invisible (v3) and checkbox (v2) reCAPTCHA site keys
	// respectively — two different reCAPTCHA products, each with its own
	// site/secret key pair (the site keys themselves are frontend-only
	// config, not read here).
	RecaptchaV3SecretKey string `env:"RECAPTCHA_V3_SECRET_KEY,default="`
	RecaptchaV2SecretKey string `env:"RECAPTCHA_V2_SECRET_KEY,default="`
	// RecaptchaScoreThreshold is the minimum v3 score (0-1) accepted
	// without a step-up v2 challenge.
	RecaptchaScoreThreshold float64 `env:"RECAPTCHA_SCORE_THRESHOLD,default=0.5"`
}

type App struct {
	HTTP *httpapi.Server
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log, err := logger.New(logger.Options{})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer log.Sync()

	var envOptions Env
	err = env.LoadEnv(&envOptions)
	if err != nil {
		log.Fatalw("failed to load environment configuration", "error", err)
	}

	db, err := postgres.NewDB(ctx, postgres.Options{
		Logger:           log,
		ConnectionString: envOptions.DatabaseURL,
		SchemaPath:       envOptions.DatabaseSchemaPath,
		MigrationsPath:   envOptions.DatabaseMigrationsPath,
	})
	if err != nil {
		log.Fatalw("failed to create database connection", "error", err)
	}
	defer db.Close()

	if err := db.Initialize(ctx); err != nil {
		log.Fatalw("failed to initialize database", "error", err)
	}

	vehicleRepository := postgres.NewVehicleRepository(db)
	accountRepository := postgres.NewAccountRepository(db)
	organizationRepository := postgres.NewOrganizationRepository(db)
	brandingRepository := postgres.NewBrandingRepository(db)
	groupRepository := postgres.NewGroupRepository(db)
	equipmentModelRepository := postgres.NewEquipmentModelRepository(db)

	vehicleService, err := service.NewVehicleService(service.VehicleServiceOptions{
		Repository: vehicleRepository,
	})
	if err != nil {
		log.Fatalw("failed to create vehicle service", "error", err)
	}

	sharedCache, err := memory.NewCache(memory.Options{DefaultExpiration: 5 * time.Minute})
	if err != nil {
		log.Fatalw("failed to create cache", "error", err)
	}

	accountService, err := service.NewAccountService(service.AccountServiceOptions{
		Repository: accountRepository,
		Cache:      sharedCache,
	})
	if err != nil {
		log.Fatalw("failed to create account service", "error", err)
	}

	organizationService, err := service.NewOrganizationService(service.OrganizationServiceOptions{
		Repository: organizationRepository,
		Accounts:   accountRepository,
	})
	if err != nil {
		log.Fatalw("failed to create organization service", "error", err)
	}

	brandingService, err := service.NewBrandingService(service.BrandingServiceOptions{
		Repository: brandingRepository,
		Cache:      sharedCache,
	})
	if err != nil {
		log.Fatalw("failed to create branding service", "error", err)
	}

	groupService, err := service.NewGroupService(service.GroupServiceOptions{
		Repository: groupRepository,
		Accounts:   accountRepository,
		Vehicles:   vehicleRepository,
	})
	if err != nil {
		log.Fatalw("failed to create group service", "error", err)
	}

	equipmentModelService, err := service.NewEquipmentModelService(service.EquipmentModelServiceOptions{
		Repository: equipmentModelRepository,
	})
	if err != nil {
		log.Fatalw("failed to create equipment model service", "error", err)
	}

	if envOptions.SeedAdminEmail != "" && envOptions.SeedAdminPassword != "" {
		seedService, err := service.NewSeedService(service.SeedOptions{
			Organizations:    organizationService,
			Accounts:         accountService,
			OrganizationName: envOptions.SeedOrganizationName,
			AdminName:        envOptions.SeedAdminName,
			AdminEmail:       envOptions.SeedAdminEmail,
			AdminPassword:    envOptions.SeedAdminPassword,
		})
		if err != nil {
			log.Fatalw("failed to create seed service", "error", err)
		}

		if err := seedService.Seed(ctx); err != nil {
			log.Fatalw("failed to seed database", "error", err)
		}
		log.Infow("seeded default organization and admin account", "email", envOptions.SeedAdminEmail)
	}

	recaptchaOptions := httpapi.RecaptchaOptions{
		Enabled:        envOptions.RecaptchaEnabled,
		V3SecretKey:    envOptions.RecaptchaV3SecretKey,
		V2SecretKey:    envOptions.RecaptchaV2SecretKey,
		ScoreThreshold: envOptions.RecaptchaScoreThreshold,
	}
	if envOptions.RecaptchaEnabled {
		recaptchaOptions.Verifier = recaptcha.NewClient()
	}

	httpServer, err := httpapi.NewServer(httpapi.Options{
		Logger:                log,
		Addr:                  envOptions.HTTPAddr,
		AllowedOrigins:        envOptions.HTTPAllowedOrigins,
		AllowInsecureCookies:  envOptions.HTTPAllowInsecureCookies,
		Recaptcha:             recaptchaOptions,
		VehicleService:        vehicleService,
		AccountService:        accountService,
		OrganizationService:   organizationService,
		BrandingService:       brandingService,
		GroupService:          groupService,
		EquipmentModelService: equipmentModelService,
	})
	if err != nil {
		log.Fatalw("failed to create http server", "error", err)
	}
	defer httpServer.Close()

	if err := httpServer.Start(); err != nil {
		log.Fatalw("failed to start http server", "error", err)
	}

	log.Info("clear-ivms-backend started")

	<-ctx.Done()

	log.Info("shutting down")
}
