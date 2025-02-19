package config

import (
	"testing"
)

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr bool
	}{
		{
			name: "valid config",
			cfg: Config{
				SourcePath: "./testdata",
				Bucket:    "test-bucket",
				Tagging:   "TagSet=[{Key=version,Value=v1.0.0}]",
			},
			wantErr: false,
		},
		{
			name: "missing source",
			cfg: Config{
				Bucket:  "test-bucket",
				Tagging: "TagSet=[{Key=version,Value=v1.0.0}]",
			},
			wantErr: true,
		},
		{
			name: "missing bucket",
			cfg: Config{
				SourcePath: "./testdata",
				Tagging:    "TagSet=[{Key=version,Value=v1.0.0}]",
			},
			wantErr: true,
		},
		{
			name: "missing tagging",
			cfg: Config{
				SourcePath: "./testdata",
				Bucket:    "test-bucket",
			},
			wantErr: true,
		},
		{
			name: "normalize s3:// prefix",
			cfg: Config{
				SourcePath: "./testdata",
				Bucket:    "s3://test-bucket",
				Tagging:   "TagSet=[{Key=version,Value=v1.0.0}]",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && tt.cfg.Bucket == "s3://test-bucket" {
				if tt.cfg.Bucket != "test-bucket" {
					t.Errorf("Bucket not normalized, got = %v, want %v", tt.cfg.Bucket, "test-bucket")
				}
			}
		})
	}
}

func TestConfig_ParseTags(t *testing.T) {
	tests := []struct {
		name    string
		tagging string
		want    []Tag
		wantErr bool
	}{
		{
			name:    "single tag",
			tagging: "TagSet=[{Key=version,Value=v1.0.0}]",
			want: []Tag{
				{Key: "version", Value: "v1.0.0"},
			},
			wantErr: false,
		},
		{
			name:    "multiple tags",
			tagging: "TagSet=[{Key=version,Value=v1.0.0},{Key=commit,Value=abc123}]",
			want: []Tag{
				{Key: "version", Value: "v1.0.0"},
				{Key: "commit", Value: "abc123"},
			},
			wantErr: false,
		},
		{
			name:    "invalid format",
			tagging: "invalid",
			wantErr: true,
		},
		{
			name:    "empty tags",
			tagging: "TagSet=[]",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{Tagging: tt.tagging}
			got, err := c.ParseTags()
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseTags() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if len(got) != len(tt.want) {
					t.Errorf("ParseTags() got = %v, want %v", got, tt.want)
					return
				}
				for i := range got {
					if got[i] != tt.want[i] {
						t.Errorf("ParseTags() got[%d] = %v, want %v", i, got[i], tt.want[i])
					}
				}
			}
		})
	}
}

func TestConfig_validateTagging(t *testing.T) {
	tests := []struct {
		name    string
		tagging string
		wantErr bool
	}{
		{
			name:    "invalid tag format",
			tagging: "TagSet=[{InvalidKey=value}]",
			wantErr: true,
		},
		{
			name:    "missing value",
			tagging: "TagSet=[{Key=version}]",
			wantErr: true,
		},
		{
			name:    "empty tag set",
			tagging: "TagSet=[]",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{Tagging: tt.tagging}
			err := c.validateTagging()
			if (err != nil) != tt.wantErr {
				t.Errorf("validateTagging() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
} 