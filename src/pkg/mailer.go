package pkg

import (
	"context"
	"fmt"
	"log"
	"sama/sama-backend-2025/src/config"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
)

// MailerService encapsulates the AWS SES v2 client and email configuration.
type MailerService struct {
	sesClient   *sesv2.Client
	senderEmail string
	senderName  string
}

// NewMailerService creates and returns a new MailerService instance.
// It requires an AWS region and a verified sender email address.
func NewMailerService(config *config.Config, cfg *aws.Config) *MailerService {

	// Create an SES v2 client
	sesClient := sesv2.NewFromConfig(*cfg)

	return &MailerService{
		sesClient:   sesClient,
		senderEmail: config.Mailer.SenderEmail,
		senderName:  config.Mailer.SenderName,
	}
}

// SendOTPEmail sends an OTP email to a specified recipient using AWS SES v2.
func (s *MailerService) SendOTPEmail(ctx context.Context, recipientName, recipientEmail, otpCode string) error {
	// Use a context with a timeout for the API call
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Construct the email subject
	subject := "Your One-Time Password (OTP)"

	// Construct the HTML and plain text bodies of the email
	htmlBody := fmt.Sprintf(`
		<html>
		<body>
			<h1>Hello %s,</h1>
			<p>Your one-time password is: <strong>%s</strong></p>
			<p>This code will expire in 5 minutes.</p>
			<p>If you did not request this, please ignore this email.</p>
		</body>
		</html>
	`, recipientName, otpCode)

	textBody := fmt.Sprintf("Hello %s,\n\nYour one-time password is: %s\n\nThis code will expire in 5 minutes. If you did not request this, please ignore this email.", recipientName, otpCode)

	// Build the email input
	input := &sesv2.SendEmailInput{
		Destination: &types.Destination{
			ToAddresses: []string{recipientEmail},
		},
		Content: &types.EmailContent{
			Simple: &types.Message{
				Subject: &types.Content{
					Data: aws.String(subject),
				},
				Body: &types.Body{
					Html: &types.Content{
						Data: aws.String(htmlBody),
					},
					Text: &types.Content{
						Data: aws.String(textBody),
					},
				},
			},
		},
		FromEmailAddress: aws.String(fmt.Sprintf("%s <%s>", s.senderName, s.senderEmail)),
	}

	// Send the email
	result, err := s.sesClient.SendEmail(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to send OTP email via SES: %w", err)
	}

	log.Printf("OTP email sent successfully. Message ID: %s", *result.MessageId)

	return nil
}
