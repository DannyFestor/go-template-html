package smtp

type Config struct{ Host string }

type Client struct{}

type Handler struct{}

type (
	Option     func(*Config)
	Middleware func(next *Handler) *Handler
	Hook       func(*Config, int)
)

func New(cfg Config, opts ...func(*Config)) *Client {
	return &Client{}
}

func (c *Client) With(fns ...func()) {}

func Join(parts ...string) {}
