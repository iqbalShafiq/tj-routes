package utils

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"strings"

	"tj-routes/internal/config"

	"go.uber.org/zap"
)

// EmailService handles sending emails
type EmailService struct {
	cfg    *config.EmailConfig
	logger *zap.Logger
}

// NewEmailService creates a new email service
func NewEmailService(cfg *config.EmailConfig, logger *zap.Logger) *EmailService {
	return &EmailService{
		cfg:    cfg,
		logger: logger,
	}
}

// IsConfigured returns true if SMTP is properly configured
func (s *EmailService) IsConfigured() bool {
	return s.cfg != nil &&
		s.cfg.SMTPHost != "" &&
		s.cfg.SMTPPort > 0 &&
		s.cfg.FromAddress != ""
}

// SendPasswordResetEmail sends a password reset email to the user
func (s *EmailService) SendPasswordResetEmail(to, username, token string) error {
	if !s.IsConfigured() {
		s.logger.Warn("SMTP not configured, skipping password reset email",
			zap.String("to", to),
			zap.String("hint", "Set SMTP_HOST, SMTP_PORT, and SMTP_FROM_ADDRESS in .env"))
		return nil
	}

	resetLink := fmt.Sprintf("%s/reset-password?token=%s", s.cfg.FrontendURL, token)

	subject := "Password Reset Request - TransJakarta Routes"
	body := s.buildPasswordResetEmailHTML(username, resetLink)

	return s.sendEmail(to, subject, body, true)
}

// buildPasswordResetEmailHTML builds the HTML email body for password reset
func (s *EmailService) buildPasswordResetEmailHTML(username, resetLink string) string {
	supportEmail := s.cfg.SupportEmail
	if supportEmail == "" {
		supportEmail = "support@transjakarta-routes.com"
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Password Reset</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #dc3545; color: white; padding: 20px; text-align: center; border-radius: 5px 5px 0 0; }
        .content { background-color: #f9f9f9; padding: 30px; border: 1px solid #ddd; border-top: none; }
        .button { display: inline-block; background-color: #dc3545; color: white; padding: 12px 30px; text-decoration: none; border-radius: 5px; margin: 20px 0; }
        .button:hover { background-color: #c82333; }
        .footer { text-align: center; padding: 20px; color: #666; font-size: 12px; }
        .warning { background-color: #fff3cd; border: 1px solid #ffc107; padding: 15px; border-radius: 5px; margin: 15px 0; }
    </style>
</head>
<body>
    <div class="header">
        <h1>Password Reset Request</h1>
    </div>
    <div class="content">
        <p>Hello %s,</p>
        <p>We received a request to reset your password for your TransJakarta Routes account.</p>
        <p>Click the button below to reset your password:</p>
        <p style="text-align: center;">
            <a href="%s" class="button">Reset Password</a>
        </p>
        <p>Or copy and paste this link into your browser:</p>
        <p style="word-break: break-all; background: #e9ecef; padding: 10px; border-radius: 3px;">%s</p>
        <div class="warning">
            <strong>Important:</strong> This link will expire in 1 hour. If you didn't request a password reset, please ignore this email or contact support if you have concerns.
        </div>
    </div>
    <div class="footer">
        <p>This email was sent by TransJakarta Routes</p>
        <p>If you have any questions, contact us at %s</p>
    </div>
</body>
</html>`, username, resetLink, resetLink, supportEmail)
}

// sendEmail sends an email using SMTP
func (s *EmailService) sendEmail(to, subject, body string, isHTML bool) error {
	from := s.cfg.FromAddress
	addr := fmt.Sprintf("%s:%d", s.cfg.SMTPHost, s.cfg.SMTPPort)

	// Build email headers
	contentType := "text/plain"
	if isHTML {
		contentType = "text/html"
	}

	headers := make(map[string]string)
	headers["From"] = from
	headers["To"] = to
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = fmt.Sprintf("%s; charset=\"UTF-8\"", contentType)

	var message strings.Builder
	for k, v := range headers {
		message.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	message.WriteString("\r\n")
	message.WriteString(body)

	// Set up authentication if credentials provided
	var auth smtp.Auth
	if s.cfg.SMTPUsername != "" && s.cfg.SMTPPassword != "" {
		auth = smtp.PlainAuth("", s.cfg.SMTPUsername, s.cfg.SMTPPassword, s.cfg.SMTPHost)
	}

	// Send email with or without TLS
	if s.cfg.SMTPUseTLS {
		return s.sendEmailTLS(addr, from, to, []byte(message.String()), auth)
	}

	// Plain SMTP (not recommended for production)
	return smtp.SendMail(addr, auth, from, []string{to}, []byte(message.String()))
}

// sendEmailTLS sends email using TLS connection
func (s *EmailService) sendEmailTLS(addr, from, to string, msg []byte, auth smtp.Auth) error {
	// Connect to the SMTP server
	conn, err := tls.Dial("tcp", addr, &tls.Config{
		ServerName: s.cfg.SMTPHost,
	})
	if err != nil {
		// Try STARTTLS if direct TLS fails
		return s.sendEmailSTARTTLS(addr, from, to, msg, auth)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, s.cfg.SMTPHost)
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}
	defer client.Close()

	if auth != nil {
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("SMTP authentication failed: %w", err)
		}
	}

	if err := client.Mail(from); err != nil {
		return fmt.Errorf("MAIL FROM failed: %w", err)
	}

	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("RCPT TO failed: %w", err)
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("DATA command failed: %w", err)
	}

	if _, err := w.Write(msg); err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	if err := w.Close(); err != nil {
		return fmt.Errorf("failed to close message writer: %w", err)
	}

	return client.Quit()
}

// sendEmailSTARTTLS sends email using STARTTLS
func (s *EmailService) sendEmailSTARTTLS(addr, from, to string, msg []byte, auth smtp.Auth) error {
	client, err := smtp.Dial(addr)
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server: %w", err)
	}
	defer client.Close()

	// Try STARTTLS
	if ok, _ := client.Extension("STARTTLS"); ok {
		config := &tls.Config{ServerName: s.cfg.SMTPHost}
		if err := client.StartTLS(config); err != nil {
			return fmt.Errorf("STARTTLS failed: %w", err)
		}
	}

	if auth != nil {
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("SMTP authentication failed: %w", err)
		}
	}

	if err := client.Mail(from); err != nil {
		return fmt.Errorf("MAIL FROM failed: %w", err)
	}

	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("RCPT TO failed: %w", err)
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("DATA command failed: %w", err)
	}

	if _, err := w.Write(msg); err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	if err := w.Close(); err != nil {
		return fmt.Errorf("failed to close message writer: %w", err)
	}

	return client.Quit()
}
