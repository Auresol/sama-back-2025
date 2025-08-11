package pkg

import (
	"context"
	"fmt"
	"sama/sama-backend-2025/src/config"
	"time"

	"github.com/mailersend/mailersend-go"
)

// MailerService encapsulates the Mailersend client and email logic.
type MailerService struct {
	mailersendClient *mailersend.Mailersend
	senderEmail      string
	senderName       string
	templateID       string
}

// NewMailerService creates and returns a new MailerService instance.
func NewMailerService(cfg *config.Config) *MailerService {
	ms := mailersend.NewMailersend(cfg.MailerSend.Key)
	return &MailerService{
		mailersendClient: ms,
		senderEmail:      cfg.MailerSend.SenderEmail,
		senderName:       cfg.MailerSend.SenderName,
		templateID:       cfg.MailerSend.OTPTemplateID,
	}
}

// SendOTPEmail sends an OTP email to a specified recipient.
func (s *MailerService) SendOTPEmail(recipientName, recipientEmail string, otpCode int) error {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	from := mailersend.From{
		Name:  s.senderName,
		Email: s.senderEmail,
	}

	recipients := []mailersend.Recipient{
		{
			Name:  recipientName,
			Email: recipientEmail,
		},
	}

	personalization := []mailersend.Personalization{
		{
			Email: recipientEmail,
			Data: map[string]interface{}{
				"code": otpCode,
			},
		},
	}

	// Create and set up the message
	message := s.mailersendClient.Email.NewMessage()
	message.SetFrom(from)
	message.SetRecipients(recipients)
	message.SetSubject("Your One-Time Password (OTP)")
	message.SetTemplateID(s.templateID)
	message.SetPersonalization(personalization)

	// Send the email
	res, err := s.mailersendClient.Email.Send(ctx, message)
	if err != nil {
		return fmt.Errorf("failed to send OTP email: %w", err)
	}

	fmt.Printf("OTP email sent successfully to %s. Message ID: %s\n", recipientEmail, res.Header.Get("X-Message-Id"))

	return nil
}
