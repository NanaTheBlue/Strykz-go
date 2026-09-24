//go:build unit

package gameserverconfig

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoad_Defaults(t *testing.T) {
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

func TestLoad_CustomValues(t *testing.T) {
	os.Setenv("AMI_ID", "ami-custom")
	os.Setenv("SUBNET_ID", "subnet-custom")
	os.Setenv("SECURITY_GROUP_ID", "sg-custom")
	os.Setenv("INSTANCE_TYPE", "m5.large")
	defer func() {
		os.Unsetenv("AMI_ID")
		os.Unsetenv("SUBNET_ID")
		os.Unsetenv("SECURITY_GROUP_ID")
		os.Unsetenv("INSTANCE_TYPE")
	}()

	cfg := Load()

	assert.Equal(t, "ami-custom", cfg.AMI)
	assert.Equal(t, "subnet-custom", cfg.SubnetID)
	assert.Equal(t, "sg-custom", cfg.SecurityGroup)
	assert.Equal(t, "m5.large", cfg.InstanceType)
}

func TestGetEnv_Fallback(t *testing.T) {
	os.Unsetenv("NONEXISTENT_KEY")
	val := getEnv("NONEXISTENT_KEY", "default")
	assert.Equal(t, "default", val)
}

func TestGetEnv_EnvSet(t *testing.T) {
	os.Setenv("EXISTING_KEY", "custom-value")
	defer os.Unsetenv("EXISTING_KEY")

	val := getEnv("EXISTING_KEY", "default")
	assert.Equal(t, "custom-value", val)
}

func TestGetEnv_EmptyEnvUsesFallback(t *testing.T) {
	os.Setenv("EMPTY_KEY", "")
	defer os.Unsetenv("EMPTY_KEY")

	val := getEnv("EMPTY_KEY", "fallback")
	assert.Equal(t, "fallback", val)
}
