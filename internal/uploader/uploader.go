package uploader

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/urth-inc/s3-cp-tagging/internal/config"
)

type s3ClientAPI interface {
	PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
}

// configLoader defines the interface for AWS configuration loading
type configLoader interface {
	LoadDefaultConfig(ctx context.Context, optFns ...func(*awsconfig.LoadOptions) error) (aws.Config, error)
}

type Uploader struct {
	client s3ClientAPI
	cfg    *config.Config
}

// NewWithLoader creates an Uploader with a specified config loader (for testing)
func NewWithLoader(cfg *config.Config, loader configLoader) (*Uploader, error) {
	var opts []func(*awsconfig.LoadOptions) error
	if cfg.Profile != "" {
		opts = append(opts, awsconfig.WithSharedConfigProfile(cfg.Profile))
	}

	awsCfg, err := loader.LoadDefaultConfig(context.TODO(), opts...)
	if err != nil {
		return nil, fmt.Errorf("unable to load AWS configuration: %v", err)
	}

	return &Uploader{
		client: s3.NewFromConfig(awsCfg),
		cfg:    cfg,
	}, nil
}

// New creates an Uploader with the default config loader
func New(cfg *config.Config) (*Uploader, error) {
	return NewWithLoader(cfg, &defaultConfigLoader{})
}

// defaultConfigLoader implements the default AWS configuration loading
type defaultConfigLoader struct{}

func (d *defaultConfigLoader) LoadDefaultConfig(ctx context.Context, optFns ...func(*awsconfig.LoadOptions) error) (aws.Config, error) {
	return awsconfig.LoadDefaultConfig(ctx, optFns...)
}

func (u *Uploader) UploadDirectory() error {
	return filepath.Walk(u.cfg.SourcePath, u.uploadFile)
}

func (u *Uploader) uploadFile(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}
	if info.IsDir() {
		return nil
	}
	if filepath.Ext(path) == ".map" {
		return nil
	}

	relPath, err := filepath.Rel(u.cfg.SourcePath, path)
	if err != nil {
		return fmt.Errorf("failed to get relative path: %v", err)
	}

	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	tags, err := u.cfg.ParseTags()
	if err != nil {
		return fmt.Errorf("failed to parse tags: %v", err)
	}

	tagging := types.Tagging{
		TagSet: make([]types.Tag, len(tags)),
	}
	for i, tag := range tags {
		tagging.TagSet[i] = types.Tag{
			Key:   aws.String(tag.Key),
			Value: aws.String(tag.Value),
		}
	}

	_, err = u.client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:  aws.String(u.cfg.Bucket),
		Key:     aws.String(relPath),
		Body:    file,
		Tagging: aws.String(formatTags(tagging.TagSet)),
	})
	if err != nil {
		return fmt.Errorf("failed to upload %s: %v", relPath, err)
	}

	fmt.Printf("upload: %s to s3://%s/%s\n", path, u.cfg.Bucket, relPath)
	return nil
}

func formatTags(tags []types.Tag) string {
	var tagPairs []string
	for _, tag := range tags {
		tagPairs = append(tagPairs, fmt.Sprintf("%s=%s", *tag.Key, *tag.Value))
	}
	return strings.Join(tagPairs, "&")
} 