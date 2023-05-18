package config

import (
	"os"
)

func Config() {
	os.Setenv("DB_USERNAME", "postgres")
	os.Setenv("DB_PASSWORD", "1234")
	os.Setenv("DB_HOST", "localhost:5432")
	os.Setenv("DB_NAME", "postgres")

	os.Setenv("DB_POOL_MAXCONN", "5")
	os.Setenv("DB_POOL_MAXCONN_LIFETIME", "300")

	os.Setenv("NATS_HOSTS", "nats://localhost:4223")
	os.Setenv("NATS_CLUSTER_ID", "test-cluster")
	os.Setenv("NATS_CLIENT_ID", "evgeniy")
	os.Setenv("NATS_SUBJECT", "testing")
	os.Setenv("NATS_DURABLE_NAME", "message")
	os.Setenv("NATS_ACK_WAIT_SECONDS", "30")

	os.Setenv("CACHE_SIZE", "10")
	os.Setenv("APP_KEY", "WB-1")
}
