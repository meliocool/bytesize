package config

type Cfg struct {
	DatabaseURL    string
	DataDir        string
	ChunkSize      int64
	Workers        int64
	MaxUploadBytes int64
}
