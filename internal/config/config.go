package config

import (
	"fmt"
	"os"
	"runtime"
	"time"
)

type Config struct {
	Env                  string
	DiskCacheDir         string
	DiskCacheMaxBytes    int64
	RamCacheMaxBytes     int64
	RamHotFileMaxBytes   int64
	BlobRoot             string
	ResultsURL           string
	WorkRoot             string
	ImagesDir            string
	OverlayFSDir         string
	StorageDir           string
	LibcontainerDir      string
	RootfsDir            string
	ServerPort           int
	MaxParallelSandboxes int
	MetricsAddr          string
	QueuePollInterval    time.Duration
	Database             DatabaseConfig
	RabbitMQ             RabbitMQConfig
	ProblemCacheDir      string
}

type DatabaseConfig struct {
	DSN string
}

type RabbitMQConfig struct {
	URL      string
	Queue    string
}

func Load() Config {
	return Config{
		Env:                  get("ENV", "dev"),
		BlobRoot:             get("BLOB_ROOT", "/var/castletown/blobs"),
		ResultsURL:           get("RESULTS_URL", "http://backend/results"),
		WorkRoot:             get("WORK_ROOT", "/tmp/castletown/work"),
		ImagesDir:            get("IMAGES_DIR", "/tmp/castletown/images"),
		OverlayFSDir:         get("OVERLAYFS_DIR", "/tmp/castletown/overlayfs"),
		StorageDir:           get("STORAGE_DIR", "/tmp/castletown/storage"),
		LibcontainerDir:      get("LIBCONTAINER_DIR", "/tmp/castletown/libcontainer"),
		RootfsDir:            get("ROOTFS_DIR", "/tmp/castletown/rootfs"),
		ServerPort:           getInt("SERVER_PORT", 8000),
		MaxParallelSandboxes: getInt("MAX_PARALLEL_SANDBOXES", runtime.NumCPU()),
		MetricsAddr:          get("METRICS_ADDR", ":9090"),
		QueuePollInterval:    time.Duration(getInt64("QUEUE_POLL_MS", 200)) * time.Millisecond,
		RabbitMQ: RabbitMQConfig{
			URL:   get("RABBITMQ_URL", "amqp://castletown:castletown@localhost:5672/"),
			Queue: get("RABBITMQ_QUEUE", "submissions"),
		},
		ProblemCacheDir: get("PROBLEM_CACHE_DIR", "/var/castletown/problems"),
		Database: DatabaseConfig{
			DSN: get("DATABASE_DSN", "user=castletown password=castletown dbname=castletown host=localhost port=5432 sslmode=disable"),
		},
	}
}

func get(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func getInt(k string, def int) int {
	return int(getInt64(k, int64(def)))
}

func getInt64(k string, def int64) int64 {
	if v := os.Getenv(k); v != "" {
		if x, err := parseInt64(v); err == nil {
			return x
		}
	}
	return def
}

func parseInt64(s string) (int64, error) { var x int64; _, err := fmt.Sscan(s, &x); return x, err }
