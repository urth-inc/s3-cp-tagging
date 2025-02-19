package uploader

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/urth-inc/s3-cp-tagging/internal/config"
)

// mockS3Client implements s3ClientAPI for testing
type mockS3Client struct {
	putObjectCalls int
	shouldError    bool
}

func (m *mockS3Client) PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	m.putObjectCalls++
	if m.shouldError {
		return nil, errors.New("mock error")
	}
	return &s3.PutObjectOutput{}, nil
}

func TestUploader_UploadDirectory(t *testing.T) {
	// Create test directory and files
	tmpDir, err := os.MkdirTemp("", "uploader_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test files
	testFiles := []string{"file1.txt", "file2.txt", "subdir/file3.txt"}
	for _, f := range testFiles {
		path := filepath.Join(tmpDir, f)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte("test content"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	tests := []struct {
		name        string
		cfg         *config.Config
		shouldError bool
		wantCalls   int
	}{
		{
			name: "successful upload",
			cfg: &config.Config{
				Bucket:     "test-bucket",
				SourcePath: tmpDir,
				Tagging:    "TagSet=[{Key=version,Value=v1.0.0}]",
			},
			shouldError: false,
			wantCalls:   3, // Expected to upload 3 files
		},
		{
			name: "upload with error",
			cfg: &config.Config{
				Bucket:     "test-bucket",
				SourcePath: tmpDir,
				Tagging:    "TagSet=[{Key=version,Value=v1.0.0}]",
			},
			shouldError: true,
			wantCalls:   1, // Should stop after first error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockS3Client{shouldError: tt.shouldError}
			u := &Uploader{
				client: mockClient,
				cfg:    tt.cfg,
			}

			err := u.UploadDirectory()
			if (err != nil) != tt.shouldError {
				t.Errorf("UploadDirectory() error = %v, wantErr %v", err, tt.shouldError)
			}

			if mockClient.putObjectCalls != tt.wantCalls {
				t.Errorf("UploadDirectory() calls = %v, want %v", mockClient.putObjectCalls, tt.wantCalls)
			}
		})
	}
}

func TestFormatTags(t *testing.T) {
	tests := []struct {
		name string
		tags []types.Tag
		want string
	}{
		{
			name: "single tag",
			tags: []types.Tag{{
				Key:   aws.String("version"),
				Value: aws.String("v1.0.0"),
			}},
			want: "version=v1.0.0",
		},
		{
			name: "multiple tags",
			tags: []types.Tag{
				{Key: aws.String("version"), Value: aws.String("v1.0.0")},
				{Key: aws.String("commit"), Value: aws.String("abc123")},
			},
			want: "version=v1.0.0&commit=abc123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatTags(tt.tags); got != tt.want {
				t.Errorf("formatTags() = %v, want %v", got, tt.want)
			}
		})
	}
}

// mockConfigLoader implements configLoader for testing
type mockConfigLoader struct {
	cfg aws.Config
	err error
}

func (m *mockConfigLoader) LoadDefaultConfig(ctx context.Context, optFns ...func(*awsconfig.LoadOptions) error) (aws.Config, error) {
	return m.cfg, m.err
}

func TestNew(t *testing.T) {
	tests := []struct {
		name      string
		cfg       *config.Config
		mockError error
		wantErr   bool
	}{
		{
			name: "valid config without profile",
			cfg: &config.Config{
				Bucket:     "test-bucket",
				SourcePath: "test-path",
				Tagging:    "TagSet=[{Key=version,Value=v1.0.0}]",
			},
			wantErr: false,
		},
		{
			name: "valid config with profile",
			cfg: &config.Config{
				Bucket:     "test-bucket",
				SourcePath: "test-path",
				Profile:    "test-profile",
				Tagging:    "TagSet=[{Key=version,Value=v1.0.0}]",
			},
			wantErr: false,
		},
		{
			name: "aws config error",
			cfg: &config.Config{
				Bucket:     "test-bucket",
				SourcePath: "test-path",
				Tagging:    "TagSet=[{Key=version,Value=v1.0.0}]",
			},
			mockError: errors.New("mock aws config error"),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLoader := &mockConfigLoader{
				cfg: aws.Config{},
				err: tt.mockError,
			}

			got, err := NewWithLoader(tt.cfg, mockLoader)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got == nil {
				t.Error("New() returned nil but wanted valid uploader")
			}
		})
	}
}

func TestUploader_uploadFile(t *testing.T) {
	// Add test cases for file upload
	tests := []struct {
		name        string
		path        string
		info        os.FileInfo
		shouldError bool
		wantErr     bool
	}{
		{
			name:        "directory should be skipped",
			path:        "testdir",
			info:        &mockFileInfo{isDir: true},
			shouldError: false,
			wantErr:     false,
		},
		{
			name:        "map file should be skipped",
			path:        "test.map",
			info:        &mockFileInfo{isDir: false},
			shouldError: false,
			wantErr:     false,
		},
		{
			name:        "invalid relative path",
			path:        string([]byte{0}), // invalid path
			info:        &mockFileInfo{isDir: false},
			shouldError: false,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockS3Client{shouldError: tt.shouldError}
			u := &Uploader{
				client: mockClient,
				cfg: &config.Config{
					Bucket:     "test-bucket",
					SourcePath: "test-source",
					Tagging:    "TagSet=[{Key=version,Value=v1.0.0}]",
				},
			}

			err := u.uploadFile(tt.path, tt.info, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("uploadFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// mockFileInfo implements os.FileInfo for testing
type mockFileInfo struct {
	isDir bool
}

func (m *mockFileInfo) Name() string       { return "mock" }
func (m *mockFileInfo) Size() int64       { return 0 }
func (m *mockFileInfo) Mode() os.FileMode { return 0 }
func (m *mockFileInfo) ModTime() time.Time { return time.Time{} }
func (m *mockFileInfo) IsDir() bool       { return m.isDir }
func (m *mockFileInfo) Sys() interface{}  { return nil } 