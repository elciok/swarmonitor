package notifier

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/mail"
	"net/smtp"
)

type SMTPConfig struct {
	Address    string
	Port       string
	User       string
	Password   string
	Domain     string
	AuthMethod string
	From       string
	To         string
}

func SendNotification(config *SMTPConfig, subject string, body string) error {
	var err error
	from := mail.Address{
		Name:    "",
		Address: config.From}
	to := mail.Address{
		Name:    "",
		Address: config.To}

	// Setup headers
	headers := make(map[string]string)
	headers["From"] = from.String()
	headers["To"] = to.String()
	headers["Subject"] = subject

	// Setup message
	message := ""
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + body

	// Connect to the SMTP Server
	servername := fmt.Sprintf("%s:%s", config.Address, config.Port)
	host, _, _ := net.SplitHostPort(servername)

	auth := smtp.PlainAuth("", config.User, config.Password, host)

	// TLS config
	tlsconfig := &tls.Config{
		InsecureSkipVerify: false,
		ServerName:         host,
	}

	var client *smtp.Client
	if config.AuthMethod == "plain" {
		client, err = smtp.Dial(servername)
		if err != nil {
			return err
		}

		err = client.StartTLS(tlsconfig)
		if err != nil {
			return err
		}
	} else {
		// Here is the key, you need to call tls.Dial instead of smtp.Dial
		// for smtp servers running on 465 that require an ssl connection
		// from the very beginning (no starttls)
		conn, err := tls.Dial("tcp", servername, tlsconfig)
		if err != nil {
			return err
		}

		client, err = smtp.NewClient(conn, host)
		if err != nil {
			return err
		}
	}

	//domain
	client.Hello(config.Domain)

	// Auth
	if err := client.Auth(auth); err != nil {
		return err
	}

	// To && From
	if err := client.Mail(from.Address); err != nil {
		return err
	}

	if err := client.Rcpt(to.Address); err != nil {
		return err
	}

	// Data
	w, err := client.Data()
	if err != nil {
		return err
	}

	_, err = w.Write([]byte(message))
	if err != nil {
		return err
	}

	err = w.Close()
	if err != nil {
		return err
	}

	client.Quit()
	return nil
}
