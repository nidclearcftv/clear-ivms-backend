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
	"github.com/nidclearcftv/clear-ivms-backend/adapter/storage/local"
	"github.com/nidclearcftv/clear-ivms-backend/adapter/storage/s3"
	"github.com/nidclearcftv/clear-ivms-backend/core/port"
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
	// HTTPMaxRequestBodyBytes bounds every request body, including
	// picture uploads (see registerEquipmentModelRoutes' PUT
	// /:id/picture) — raised well above httpapi's own 1 MiB default so a
	// normal photo isn't rejected outright.
	HTTPMaxRequestBodyBytes int64 `env:"HTTP_MAX_REQUEST_BODY_BYTES,default=8388608"`
	// HTTPPublicBaseURL is this backend's own externally-reachable API
	// base URL — e.g. the real deployed address, not just "localhost" in
	// anything but local dev. Used by adapter/storage/local to build URLs
	// under registerObjectRoutes that a browser can actually reach; see
	// local.Options.PublicBaseURL.
	HTTPPublicBaseURL string `env:"HTTP_PUBLIC_BASE_URL,default=http://localhost:8080/api/v1"`

	// ObjectStorageDriver selects the port.ObjectStorage implementation:
	// "local" (adapter/storage/local) or "s3" (adapter/storage/s3).
	// Unrecognized values fail startup rather than silently falling back
	// to something unintended.
	ObjectStorageDriver string `env:"OBJECT_STORAGE_DRIVER,default=local"`
	// ObjectStorageLocalPath is the directory adapter/storage/local
	// stores objects under, only used when ObjectStorageDriver is
	// "local".
	ObjectStorageLocalPath string `env:"OBJECT_STORAGE_LOCAL_PATH,default=data/objects"`

	// ObjectStorageS3* configure adapter/storage/s3, only used when
	// ObjectStorageDriver is "s3" — see s3.Options for what each maps to.
	// AccessKeyID/SecretAccessKey/SessionToken are optional: leave all
	// empty to use the AWS SDK's default credential chain instead.
	ObjectStorageS3Bucket          string        `env:"OBJECT_STORAGE_S3_BUCKET,default="`
	ObjectStorageS3Region          string        `env:"OBJECT_STORAGE_S3_REGION,default="`
	ObjectStorageS3Endpoint        string        `env:"OBJECT_STORAGE_S3_ENDPOINT,default="`
	ObjectStorageS3UsePathStyle    bool          `env:"OBJECT_STORAGE_S3_USE_PATH_STYLE,default=false"`
	ObjectStorageS3AccessKeyID     string        `env:"OBJECT_STORAGE_S3_ACCESS_KEY_ID,default="`
	ObjectStorageS3SecretAccessKey string        `env:"OBJECT_STORAGE_S3_SECRET_ACCESS_KEY,default="`
	ObjectStorageS3SessionToken    string        `env:"OBJECT_STORAGE_S3_SESSION_TOKEN,default="`
	ObjectStorageS3PresignExpiry   time.Duration `env:"OBJECT_STORAGE_S3_PRESIGN_EXPIRY,default=15m"`

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
	vehicleEquipmentRepository := postgres.NewVehicleEquipmentRepository(db)

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

	// registerObjectRoutesEnabled tracks whether the configured driver
	// needs registerObjectRoutes (see server.go's Options.ObjectStorage)
	// — true for "local", whose PutURL/GetURL point back at those routes;
	// a driver with real external presigned URLs (e.g. a future S3 one)
	// would leave this false, needing no such thing.
	var objectStorage port.ObjectStorage
	registerObjectRoutesEnabled := false
	switch envOptions.ObjectStorageDriver {
	case "local":
		objectStorage, err = local.NewStorage(local.Options{
			BasePath:      envOptions.ObjectStorageLocalPath,
			PublicBaseURL: envOptions.HTTPPublicBaseURL,
		})
		if err != nil {
			log.Fatalw("failed to create local object storage", "error", err)
		}
		registerObjectRoutesEnabled = true
	case "s3":
		objectStorage, err = s3.NewStorage(ctx, s3.Options{
			Bucket:          envOptions.ObjectStorageS3Bucket,
			Region:          envOptions.ObjectStorageS3Region,
			Endpoint:        envOptions.ObjectStorageS3Endpoint,
			UsePathStyle:    envOptions.ObjectStorageS3UsePathStyle,
			AccessKeyID:     envOptions.ObjectStorageS3AccessKeyID,
			SecretAccessKey: envOptions.ObjectStorageS3SecretAccessKey,
			SessionToken:    envOptions.ObjectStorageS3SessionToken,
			PresignExpiry:   envOptions.ObjectStorageS3PresignExpiry,
		})
		if err != nil {
			log.Fatalw("failed to create s3 object storage", "error", err)
		}
		// registerObjectRoutesEnabled stays false: S3's PutURL/GetURL
		// point straight at AWS, so this backend needs no self-serving
		// passthrough for it.
	default:
		log.Fatalw("unsupported object storage driver", "driver", envOptions.ObjectStorageDriver)
	}

	equipmentModelService, err := service.NewEquipmentModelService(service.EquipmentModelServiceOptions{
		Repository: equipmentModelRepository,
		Storage:    objectStorage,
	})
	if err != nil {
		log.Fatalw("failed to create equipment model service", "error", err)
	}

	vehicleEquipmentService, err := service.NewVehicleEquipmentService(service.VehicleEquipmentServiceOptions{
		Repository:      vehicleEquipmentRepository,
		Vehicles:        vehicleRepository,
		EquipmentModels: equipmentModelRepository,
	})
	if err != nil {
		log.Fatalw("failed to create vehicle equipment service", "error", err)
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

	var httpObjectStorage port.ObjectStorage
	if registerObjectRoutesEnabled {
		httpObjectStorage = objectStorage
	}

	httpServer, err := httpapi.NewServer(httpapi.Options{
		Logger:                  log,
		Addr:                    envOptions.HTTPAddr,
		AllowedOrigins:          envOptions.HTTPAllowedOrigins,
		AllowInsecureCookies:    envOptions.HTTPAllowInsecureCookies,
		MaxRequestBodyBytes:     envOptions.HTTPMaxRequestBodyBytes,
		Recaptcha:               recaptchaOptions,
		VehicleService:          vehicleService,
		AccountService:          accountService,
		OrganizationService:     organizationService,
		BrandingService:         brandingService,
		GroupService:            groupService,
		EquipmentModelService:   equipmentModelService,
		VehicleEquipmentService: vehicleEquipmentService,
		ObjectStorage:           httpObjectStorage,
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
