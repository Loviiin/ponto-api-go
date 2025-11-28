package email

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"html/template"
	"log"
	"net/smtp"
	"strconv"
	"time"
)

type EmailService struct {
	host        string
	port        int
	username    string
	password    string
	fromName    string
	fromEmail   string
	frontendURL string
}

func NewEmailService(host, portStr, username, password, fromName, fromEmail, frontendURL string) *EmailService {
	// Se qualquer configuração crítica estiver vazia, retorna nil
	if host == "" || portStr == "" || username == "" || password == "" {
		log.Printf("⚠️  SMTP não configurado: defina SMTP_HOST, SMTP_PORT, SMTP_USER e SMTP_PASSWORD no .env")
		return nil
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		log.Printf("⚠️  SMTP_PORT inválido: %v", err)
		return nil
	}

	log.Printf("✅ Serviço de email configurado: %s@%s:%d", username, host, port)

	return &EmailService{
		host:        host,
		port:        port,
		username:    username,
		password:    password,
		fromName:    fromName,
		fromEmail:   fromEmail,
		frontendURL: frontendURL,
	}
}

func (s *EmailService) SendEmail(to []string, subject string, templateName string, data interface{}) error {
	// Verifica se o serviço está configurado
	if s == nil {
		return fmt.Errorf("serviço de email não configurado - verifique as variáveis SMTP no .env")
	}

	// Parse template
	tmpl, err := template.New(templateName).Parse(GetTemplate(templateName))
	if err != nil {
		return fmt.Errorf("failed to parse template: %v", err)
	}

	var body bytes.Buffer
	if err := tmpl.Execute(&body, data); err != nil {
		return fmt.Errorf("failed to execute template: %v", err)
	}

	// Setup headers
	headers := make(map[string]string)
	headers["From"] = fmt.Sprintf("%s <%s>", s.fromName, s.fromEmail)
	headers["To"] = to[0] // Simplification for single recipient
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=\"UTF-8\""

	message := ""
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + body.String()

	// Authentication
	auth := smtp.PlainAuth("", s.username, s.password, s.host)

	// TLS Config
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         s.host,
	}

	// Connect to SMTP Server
	addr := fmt.Sprintf("%s:%d", s.host, s.port)
	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		// Fallback to non-TLS if direct TLS fails (e.g. port 587 often uses STARTTLS)
		return s.sendWithStartTLS(addr, auth, s.fromEmail, to, []byte(message))
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, s.host)
	if err != nil {
		return err
	}
	defer client.Quit()

	if err = client.Auth(auth); err != nil {
		return err
	}

	if err = client.Mail(s.fromEmail); err != nil {
		return err
	}

	for _, recipient := range to {
		if err = client.Rcpt(recipient); err != nil {
			return err
		}
	}

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

	return nil
}

func (s *EmailService) sendWithStartTLS(addr string, auth smtp.Auth, from string, to []string, msg []byte) error {
	return smtp.SendMail(addr, auth, from, to, msg)
}

// Retry logic wrapper
func (s *EmailService) SendEmailWithRetry(to []string, subject string, templateName string, data interface{}) error {
	var err error
	for i := 0; i < 3; i++ {
		err = s.SendEmail(to, subject, templateName, data)
		if err == nil {
			return nil
		}
		log.Printf("Failed to send email (attempt %d/3): %v", i+1, err)
		time.Sleep(time.Second * time.Duration(i+1))
	}
	return fmt.Errorf("failed to send email after 3 attempts: %v", err)
}

// BuildPasswordResetLink constrói o link completo de reset de senha
func (s *EmailService) BuildPasswordResetLink(token string) string {
	if s == nil || s.frontendURL == "" {
		return fmt.Sprintf("/reset-password?token=%s", token)
	}
	return fmt.Sprintf("%s/reset-password?token=%s", s.frontendURL, token)
}
