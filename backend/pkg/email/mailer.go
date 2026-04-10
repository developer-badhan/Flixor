package email

import (
	"crypto/tls"
	"fmt"

	"gopkg.in/gomail.v2"
)

/**
 * Mailer wraps gomail for sending transactional emails via Gmail SMTP.
 * Constructed once at startup and injected into UserService.
*/
type Mailer struct {
	dialer *gomail.Dialer
	from   string
}

/**
 * NewMailer creates a Mailer configured for Gmail's implicit-TLS port (465).
 * For port 587 (STARTTLS), remove the SSL=true line — gomail uses STARTTLS by default.
*/
func NewMailer(host string, port int, user, password string) *Mailer {
	d := gomail.NewDialer(host, port, user, password)

	// Port 465 = implicit SSL — the TLS handshake happens before the SMTP greeting.
	// Port 587 = STARTTLS — negotiated after the SMTP greeting. Different protocol.
	d.SSL = true
	d.TLSConfig = &tls.Config{
		ServerName: host, // must match the certificate CN — prevents MITM
	}

	return &Mailer{dialer: d, from: user}
}

// SendOTP delivers a 6-digit verification code to the given email address.
func (m *Mailer) SendOTP(toEmail, username, otp string) error {
	msg := gomail.NewMessage()
	msg.SetHeader("From", m.from)
	msg.SetHeader("To", toEmail)
	msg.SetHeader("Subject", "Your Flixor verification code")
	msg.SetBody("text/html", buildOTPBody(username, otp))

	if err := m.dialer.DialAndSend(msg); err != nil {
		return fmt.Errorf("failed to send OTP email: %w", err)
	}
	return nil
}

/**
 * buildOTPBody returns a minimal HTML email body.
 * Kept intentionally simple — no external images, no tracking pixels.
*/
func buildOTPBody(username, otp string) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<body style="font-family:sans-serif;max-width:480px;margin:40px auto;color:#1a1a1a">
  <h2 style="font-size:22px;font-weight:500">Hello, %s</h2>
  <p style="font-size:16px">Your Flixor verification code is:</p>
  <div style="font-size:36px;font-weight:700;letter-spacing:12px;padding:20px;
              background:#f4f4f4;text-align:center;border-radius:8px;
              margin:24px 0">%s</div>
  <p style="font-size:14px;color:#666">
    This code expires in <strong>10 minutes</strong>.<br>
    If you did not request this, ignore this email.
  </p>
</body>
</html>`, username, otp)
}