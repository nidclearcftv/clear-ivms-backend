// Package s3 is an Amazon S3 (or S3-compatible) object storage adapter,
// implementing port.ObjectStorage via presigned URLs — the client
// uploads/downloads directly to/from S3, never through this backend.
package s3

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"

	"github.com/nidclearcftv/clear-ivms-backend/core/model"
	"github.com/nidclearcftv/clear-ivms-backend/core/port"
	"github.com/nidclearcftv/clear-ivms-backend/utils/validate"
)

// defaultPresignExpiry is how long a PutURL/GetURL stays valid when
// Options.PresignExpiry isn't set.
const defaultPresignExpiry = 15 * time.Minute

type Options struct {
	Bucket string `validate:"required"`
	Region string `validate:"required"`

	// Endpoint overrides the default AWS S3 endpoint — set this to point
	// at an S3-compatible provider (e.g. MinIO, Cloudflare R2) instead of
	// real AWS S3. Leave empty for real AWS S3.
	Endpoint string
	// UsePathStyle addresses objects as "<endpoint>/<bucket>/<key>"
	// instead of AWS's default "<bucket>.<endpoint>/<key>" — required by
	// most S3-compatible providers when Endpoint is set; leave false for
	// real AWS S3.
	UsePathStyle bool

	// AccessKeyID/SecretAccessKey/SessionToken supply static credentials.
	// Leave all empty to use the AWS SDK's default credential chain
	// instead (environment variables, shared config/credentials files,
	// an EC2/ECS/EKS instance role, ...) — the usual choice in a real AWS
	// deployment.
	AccessKeyID     string
	SecretAccessKey string
	SessionToken    string

	// PresignExpiry bounds how long a PutURL/GetURL stays valid. Defaults
	// to 15 minutes.
	PresignExpiry time.Duration `validate:"omitempty,gt=0"`
}

// Storage implements port.ObjectStorage against an S3 bucket. PutURL/
// GetURL return real S3 presigned URLs; Put/Get go through the S3 API
// directly instead — see port.ObjectStorage's doc comment for who
// actually still calls those.
type Storage struct {
	client        *s3.Client
	presignClient *s3.PresignClient
	bucket        string
}

func NewStorage(ctx context.Context, opts Options) (*Storage, error) {
	if err := validate.Struct(opts); err != nil {
		return nil, err
	}

	expiry := opts.PresignExpiry
	if expiry == 0 {
		expiry = defaultPresignExpiry
	}

	loadOptFns := []func(*awsconfig.LoadOptions) error{
		awsconfig.WithRegion(opts.Region),
	}
	if opts.AccessKeyID != "" || opts.SecretAccessKey != "" {
		loadOptFns = append(loadOptFns, awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(opts.AccessKeyID, opts.SecretAccessKey, opts.SessionToken),
		))
	}

	cfg, err := awsconfig.LoadDefaultConfig(ctx, loadOptFns...)
	if err != nil {
		return nil, fmt.Errorf("storage/s3: failed to load AWS config: %w", err)
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		if opts.Endpoint != "" {
			o.BaseEndpoint = aws.String(opts.Endpoint)
		}
		o.UsePathStyle = opts.UsePathStyle
	})

	return &Storage{
		client:        client,
		presignClient: s3.NewPresignClient(client, s3.WithPresignExpires(expiry)),
		bucket:        opts.Bucket,
	}, nil
}

func (s *Storage) Put(ctx context.Context, key string, r io.Reader, contentType string) error {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        r,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return fmt.Errorf("storage/s3: failed to put object: %w", err)
	}
	return nil
}

// PutURL returns a presigned URL the client can PUT key's bytes to
// directly. contentType is baked into the presigned request's signature
// — the client's actual PUT must send the exact same Content-Type header
// or S3 rejects it as a signature mismatch, which is as close to
// server-side content-type enforcement as a presigned upload gets (see
// model.EquipmentModelAllowedPictureContentTypes for what checks this
// value before it's ever handed to PutURL).
func (s *Storage) PutURL(ctx context.Context, key string, contentType string) (string, error) {
	req, err := s.presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return "", fmt.Errorf("storage/s3: failed to presign put object: %w", err)
	}
	return req.URL, nil
}

func (s *Storage) Get(ctx context.Context, key string) (io.ReadCloser, string, error) {
	out, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		var notFound *types.NoSuchKey
		if errors.As(err, &notFound) {
			return nil, "", model.NewError(model.ErrCodeObjectNotFound, err)
		}
		return nil, "", fmt.Errorf("storage/s3: failed to get object: %w", err)
	}

	contentType := "application/octet-stream"
	if out.ContentType != nil {
		contentType = *out.ContentType
	}
	return out.Body, contentType, nil
}

func (s *Storage) GetURL(ctx context.Context, key string) (string, error) {
	req, err := s.presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return "", fmt.Errorf("storage/s3: failed to presign get object: %w", err)
	}
	return req.URL, nil
}

// Delete removes the object at key — S3's DeleteObject already treats a
// nonexistent key as success, so this needs no extra handling to satisfy
// port.ObjectStorage's "deleting a key that doesn't exist is not an
// error" contract.
func (s *Storage) Delete(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("storage/s3: failed to delete object: %w", err)
	}
	return nil
}

var _ port.ObjectStorage = (*Storage)(nil)
