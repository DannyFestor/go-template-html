package auth

// Broken: Run takes no context.
type (
	Enable    struct{}
	EnableIn  struct{}
	EnableOut struct{}
)

func NewEnable() *Enable { return &Enable{} }

func (*Enable) Run(in EnableIn) (EnableOut, error) { return EnableOut{}, nil }
