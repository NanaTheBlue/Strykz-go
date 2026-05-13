package gameserverconfig

import "os"

type Config struct {
	AMI           string
	SubnetID      string
	SecurityGroup string
	InstanceType  string
}

func Load() Config {
	return Config{
		AMI:           getEnv("AMI_ID", "ami-0a2dfeedd475ba8ed"),
		SubnetID:      getEnv("SUBNET_ID", ""),
		SecurityGroup: getEnv("SECURITY_GROUP_ID", ""),
		InstanceType:  getEnv("INSTANCE_TYPE", "t3.micro"),
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
