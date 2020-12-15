package notifier

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/mail"
	"net/smtp"
	"strings"

	"github.com/elciok/swarmonitor/config"
	"github.com/elciok/swarmonitor/status"
)

func SendNotification(config *config.Config, status *status.Status) error {
	var err error
	from := mail.Address{
		Name:    "",
		Address: config.SMTP.From}
	to := mail.Address{
		Name:    "",
		Address: config.SMTP.To}
	subj := fmt.Sprintf("swarmonitor - %s is %s", status.Target, statusString(status))
	body := bodyString(status)

	// Setup headers
	headers := make(map[string]string)
	headers["From"] = from.String()
	headers["To"] = to.String()
	headers["Subject"] = subj

	// Setup message
	message := ""
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + body

	// Connect to the SMTP Server
	servername := fmt.Sprintf("%s:%s", config.SMTP.Address, config.SMTP.Port)
	host, _, _ := net.SplitHostPort(servername)

	auth := smtp.PlainAuth("", config.SMTP.User, config.SMTP.Password, host)

	// TLS config
	tlsconfig := &tls.Config{
		InsecureSkipVerify: false,
		ServerName:         host,
	}

	var client *smtp.Client
	if config.SMTP.AuthMethod == "plain" {
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
	client.Hello(config.SMTP.Domain)

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

func bodyString(status *status.Status) string {
	var builder strings.Builder
	fmt.Fprintf(&builder, "Status: %s\r\n\r\n", statusString(status))
	fmt.Fprint(&builder, "Labels:\r\n")
	for labelKey, labelValue := range status.Labels {
		fmt.Fprintf(&builder, "\t- %s = %s\r\n", labelKey, labelValue)
	}
	return builder.String()
}

func statusString(status *status.Status) string {
	if status.Ok() {
		return "OK"
	} else {
		return "DOWN"
	}
}
