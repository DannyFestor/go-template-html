package bootstrap

type PasswordHashing struct {
	MemoryKiB   uint32 `env:"PASSWORD_HASH_MEMORY_KIB"  envDefault:"19456"`
	Iterations  uint32 `env:"PASSWORD_HASH_ITERATIONS"  envDefault:"2"`
	Parallelism uint8  `env:"PASSWORD_HASH_PARALLELISM" envDefault:"1"`
}
