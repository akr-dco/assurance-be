package utils

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"log"
	"net"
	"net/smtp"
	"os"
)

func SendEmail(to, subject, body string) error {

	//from := "appnotif@phoenixsolusi.com"
	//password := "Phoenix@Kebagusan20@%"
	//host := "smtp.office365.com"
	//port := "587"

	from := os.Getenv("SMTP_FROM")
	password := os.Getenv("SMTP_PASSWORD")
	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")

	log.Println("[EMAIL] Connecting TCP...")
	conn, err := net.Dial("tcp", host+":"+port)
	if err != nil {
		return fmt.Errorf("TCP CONNECT ERROR: %w", err)
	}

	// Create SMTP client
	c, err := smtp.NewClient(conn, host)
	if err != nil {
		return fmt.Errorf("SMTP CLIENT ERROR: %w", err)
	}

	// STARTTLS
	log.Println("[EMAIL] Starting STARTTLS...")
	if err = c.StartTLS(&tls.Config{ServerName: host}); err != nil {
		return fmt.Errorf("STARTTLS ERROR: %w", err)
	}

	// AUTH LOGIN
	log.Println("[EMAIL] AUTH LOGIN...")
	if err := authLogin(c, from, password); err != nil {
		return fmt.Errorf("AUTH LOGIN ERROR: %w", err)
	}
	log.Println("[EMAIL] AUTH SUCCESS")

	// MAIL FROM
	if err = c.Mail(from); err != nil {
		return fmt.Errorf("MAIL FROM ERROR: %w", err)
	}

	// RCPT TO
	if err = c.Rcpt(to); err != nil {
		return fmt.Errorf("RCPT TO ERROR: %w", err)
	}

	// DATA
	w, err := c.Data()
	if err != nil {
		return fmt.Errorf("DATA ERROR: %w", err)
	}

	msg := []byte(
		"From: " + from + "\r\n" +
			"To: " + to + "\r\n" +
			"Subject: " + subject + "\r\n" +
			"Content-Type: text/html; charset=UTF-8\r\n\r\n" +
			body,
	)

	_, err = w.Write(msg)
	if err != nil {
		return fmt.Errorf("WRITE ERROR: %w", err)
	}

	w.Close()
	c.Quit()

	log.Println("[EMAIL] EMAIL SENT SUCCESSFULLY!")
	return nil
}

func authLogin(c *smtp.Client, username, password string) error {

	// Step 1: AUTH LOGIN
	if err := c.Text.PrintfLine("AUTH LOGIN"); err != nil {
		return fmt.Errorf("AUTH LOGIN INIT: %w", err)
	}

	// Expect: 334 VXNlcm5hbWU6
	if _, _, err := c.Text.ReadResponse(334); err != nil {
		return fmt.Errorf("USERNAME PROMPT ERROR: %w", err)
	}

	// Send username (base64)
	if err := c.Text.PrintfLine(base64.StdEncoding.EncodeToString([]byte(username))); err != nil {
		return fmt.Errorf("SEND USERNAME ERROR: %w", err)
	}

	// Expect: 334 UGFzc3dvcmQ6
	if _, _, err := c.Text.ReadResponse(334); err != nil {
		return fmt.Errorf("PASSWORD PROMPT ERROR: %w", err)
	}

	// Send password (base64)
	if err := c.Text.PrintfLine(base64.StdEncoding.EncodeToString([]byte(password))); err != nil {
		return fmt.Errorf("SEND PASSWORD ERROR: %w", err)
	}

	// Expect: 235 Authentication successful
	if _, _, err := c.Text.ReadResponse(235); err != nil {
		return fmt.Errorf("AUTH FAILED: %w", err)
	}

	return nil
}

func BuildResetPasswordEmail(name, resetLink string) string {
	return fmt.Sprintf(`
  <div style="font-family: Arial, sans-serif; background: #f6f7fb; padding: 40px;">
    <div style="
      max-width: 520px;
      margin: auto;
      background: white;
      padding: 30px 40px;
      border-radius: 12px;
      box-shadow: 0 4px 20px rgba(0,0,0,0.08);
    ">
      <h2 style="color:#3f51b5; margin-bottom: 10px; text-align:center;">
        Reset Your Password
      </h2>

      <p style="font-size: 14px; color:#444;">
        Hello <strong>%s</strong>,
      </p>

      <p style="font-size: 14px; color:#444; line-height: 1.6;">
        We received a request to reset your password.  
        Click the button below to proceed:
      </p>

      <div style="text-align:center; margin: 28px 0;">
        <a href="%s" 
            style="
              background:#3f51b5;
              color:white;
              padding:12px 28px;
              border-radius:6px;
              text-decoration:none;
              font-weight:bold;
              display:inline-block;
            ">
          Reset Password
        </a>
      </div>

      <p style="font-size: 13px; color:#666; line-height:1.5;">
        If the button above doesn't work, copy and paste this link into your browser:
      </p>

      <p style="font-size: 12px; color:#3f51b5; word-break:break-all;">
        %s
      </p>

      <hr style="border:none; border-top:1px solid #eee; margin: 25px 0;">

      <p style="font-size: 12px; color:#999; text-align:center;">
        This link will expire in <strong>60 minutes</strong>.  
        If you did not request a password reset, please ignore this email.
      </p>
    </div>

    <p style="text-align:center; font-size:11px; color:#aaa; margin-top:25px;">
      © 2025 Phoenix Solusi • All Rights Reserved
    </p>
  </div>
  `, name, resetLink, resetLink)
}
