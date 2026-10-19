package bootstrap

import "github.com/DannyFestor/go-template-html/internal/adapter/smtp"

type Mail struct {
	Host        string          `env:"MAIL_HOST"         envDefault:"localhost"`
	Port        int             `env:"MAIL_PORT"         envDefault:"1025"`
	Username    string          `env:"MAIL_USERNAME"`
	Password    Secret          `env:"MAIL_PASSWORD"`
	Encryption  smtp.Encryption `env:"MAIL_ENCRYPTION"   envDefault:"none"`
	FromAddress string          `env:"MAIL_FROM_ADDRESS" envDefault:"hello@example.com"`
}

func smtpConfig(app App, mail Mail) smtp.Config {
	return smtp.Config{
		Host:        mail.Host,
		Port:        mail.Port,
		Username:    mail.Username,
		Password:    mail.Password.Reveal(),
		Encryption:  mail.Encryption,
		FromName:    app.Name,
		FromAddress: mail.FromAddress,
	}
}
