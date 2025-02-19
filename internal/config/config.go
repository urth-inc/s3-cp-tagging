package config

import (
	"fmt"
	"strings"
)

type Config struct {
	Bucket     string
	SourcePath string
	Profile    string
	Tagging    string
}

type Tag struct {
	Key   string `json:"Key"`
	Value string `json:"Value"`
}

type TagSet struct {
	TagSet []Tag `json:"TagSet"`
}

// Validate checks if all required fields are set and normalizes the bucket name
func (c *Config) Validate() error {
	var missing []string
	
	if c.SourcePath == "" {
		missing = append(missing, "SOURCE")
	}
	if c.Bucket == "" {
		missing = append(missing, "BUCKET")
	}
	if c.Tagging == "" {
		missing = append(missing, "--tagging")
	}
	
	if len(missing) > 0 {
		return fmt.Errorf("missing required arguments/options: %s", strings.Join(missing, ", "))
	}

	// Normalize bucket name format
	c.Bucket = strings.TrimPrefix(c.Bucket, "s3://")
	c.Bucket = strings.TrimPrefix(c.Bucket, "s3:/")

	// Validate tag format
	if err := c.validateTagging(); err != nil {
		return fmt.Errorf("invalid tagging format: %v", err)
	}
	
	return nil
}

func (c *Config) validateTagging() error {
	// Remove "TagSet=" prefix
	tagStr := strings.TrimPrefix(c.Tagging, "TagSet=")
	
	// Split string into key-value pairs
	tagStr = strings.Trim(tagStr, "[]")
	if tagStr == "" {
		return fmt.Errorf("no tags provided")
	}

	// Parse each tag
	tagPairs := strings.Split(tagStr, "},{")
	for i, pair := range tagPairs {
		pair = strings.Trim(pair, "{}")
		if !strings.Contains(pair, "Key=") || !strings.Contains(pair, "Value=") {
			return fmt.Errorf("invalid tag format in pair %d: %s", i+1, pair)
		}
	}
	
	return nil
}

func (c *Config) ParseTags() ([]Tag, error) {
	tagStr := strings.TrimPrefix(c.Tagging, "TagSet=")
	tagStr = strings.Trim(tagStr, "[]")
	
	var tags []Tag
	tagPairs := strings.Split(tagStr, "},{")
	
	for _, pair := range tagPairs {
		pair = strings.Trim(pair, "{}")
		parts := strings.Split(pair, ",")
		
		var tag Tag
		for _, part := range parts {
			kv := strings.Split(part, "=")
			if len(kv) != 2 {
				continue
			}
			switch strings.TrimSpace(kv[0]) {
			case "Key":
				tag.Key = strings.TrimSpace(kv[1])
			case "Value":
				tag.Value = strings.TrimSpace(kv[1])
			}
		}
		
		if tag.Key != "" && tag.Value != "" {
			tags = append(tags, tag)
		}
	}
	
	if len(tags) == 0 {
		return nil, fmt.Errorf("no valid tags found")
	}
	
	return tags, nil
} 