package bootstrap

type Config struct {
	App             App
	HTTP            HTTP
	DB              DB
	Valkey          Valkey
	Session         Session
	PasswordHashing PasswordHashing
	SignedURL       SignedURL
	Queue           Queue
	Mail            Mail
	Log             Log
}
