//go:build unit

package gameserverconfig

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoad_WithEnvironmentVariables(t *testing.T) {
	// Set environment variables
	os.Setenv("AMI_ID", "ami-customvalue")
	os.Setenv("SUBNET_ID", "subnet-custom123")
	os.Setenv("SECURITY_GROUP_ID", "sg-custom456")
	os.Setenv("INSTANCE_TYPE", "t3.large")

	cfg := Load()

	assert.Equal(t, "ami-customvalue", cfg.AMI)
	assert.Equal(t, "subnet-custom123", cfg.SubnetID)
	assert.Equal(t, "sg-custom456", cfg.SecurityGroup)
	assert.Equal(t, "t3.large", cfg.InstanceType)

	// Clean up
	os.Unsetenv("AMI_ID")
	os.Unsetenv("SUBNET_ID")
	os.Unsetenv("SECURITY_GROUP_ID")
	os.Unsetenv("INSTANCE_TYPE")
}

func TestLoad_WithDefaults(t *testing.T) {
	// Make sure environment variables are not set
	os.Unsetenv("AMI_ID")
	os.Unsetenv("SUBNET_ID")
	os.Unsetenv("SECURITY_GROUP_ID")
	os.Unsetenv("INSTANCE_TYPE")

	cfg := Load()

	assert.Equal(t, "ami-0a2dfeedd475ba8ed", cfg.AMI)
	assert.Equal(t, "", cfg.SubnetID)
	assert.Equal(t, "", cfg.SecurityGroup)
	assert.Equal(t, "t3.micro", cfg.InstanceType)
}

func TestLoad_PartialEnvironmentVariables(t *testing.T) {
	// Set only some environment variables
	os.Setenv("AMI_ID", "ami-partial123")
	os.Unsetenv("SUBNET_ID")
	os.Setenv("SECURITY_GROUP_ID", "sg-partial789")
	os.Unsetenv("INSTANCE_TYPE")

	cfg := Load()

	assert.Equal(t, "ami-partial123", cfg.AMI)
	assert.Equal(t, "", cfg.SubnetID)
	assert.Equal(t, "sg-partial789", cfg.SecurityGroup)
	assert.Equal(t, "t3.micro", cfg.InstanceType)

	// Clean up
	os.Unsetenv("AMI_ID")
	os.Unsetenv("SECURITY_GROUP_ID")
}

func TestLoad_EmptyStringFallsBackToDefault(t *testing.T) {
	// Set empty strings to test fallback behavior
	os.Setenv("AMI_ID", "")
	os.Setenv("INSTANCE_TYPE", "")

	cfg := Load()

	assert.Equal(t, "ami-0a2dfeedd475ba8ed", cfg.AMI)
	assert.Equal(t, "t3.micro", cfg.InstanceType)

	// Clean up
	os.Unsetenv("AMI_ID")
	os.Unsetenv("INSTANCE_TYPE")
}

func TestGetEnv_WithValue(t *testing.T) {
	os.Setenv("TEST_KEY", "test_value")

	value := getEnv("TEST_KEY", "fallback")

	assert.Equal(t, "test_value", value)

	os.Unsetenv("TEST_KEY")
}

func TestGetEnv_WithoutValueUsesFallback(t *testing.T) {
	os.Unsetenv("TEST_KEY")

	value := getEnv("TEST_KEY", "fallback")

	assert.Equal(t, "fallback", value)
}

func TestConfigStructure(t *testing.T) {
	cfg := Config{
		AMI:           "ami-123456",
		SubnetID:      "subnet-789",
		SecurityGroup: "sg-abc",
		InstanceType:  "t3.small",
	}

	assert.NotNil(t, cfg)
	assert.Equal(t, "ami-123456", cfg.AMI)
	assert.Equal(t, "subnet-789", cfg.SubnetID)
	assert.Equal(t, "sg-abc", cfg.SecurityGroup)
	assert.Equal(t, "t3.small", cfg.InstanceType)
}
