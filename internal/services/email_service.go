package services

import (
	"fmt"
	"golang-mongodb-rest-api-starter/internal/models"

	"github.com/mailjet/mailjet-apiv3-go"
)

type EmailServiceWrapper interface {
	SendVerificationCode(email, username, code string)
}

type EmailService struct {
	Client       *mailjet.Client
	ReplyToEmail string
	FromEmail    string
	FromName     string
}

func NewEmailService(client *mailjet.Client, replyToEmail, fromName, fromEmail string) *EmailService {
	return &EmailService{Client: client, ReplyToEmail: replyToEmail, FromEmail: fromEmail, FromName: fromName}
}

func (s *EmailService) SendVerificationCode(email, username, code string) error {
	messagesInfo := []mailjet.InfoMessagesV31{
		{
			From: &mailjet.RecipientV31{
				Email: s.FromEmail,
				Name:  s.FromName,
			},
			ReplyTo: &mailjet.RecipientV31{
				Email: s.ReplyToEmail,
				Name:  s.FromName,
			}, To: &mailjet.RecipientsV31{
				mailjet.RecipientV31{
					Email: email,
					Name:  username,
				},
			},
			TemplateID:       4578365,
			TemplateLanguage: true,
			Subject:          "Verification code for TODO change this",
			Variables: map[string]interface{}{
				"name": username,
				"code": code,
			},
		},
	}

	sendEmailData := mailjet.MessagesV31{Info: messagesInfo}
	_, err := s.Client.SendMailV31(&sendEmailData)
	if err != nil {
		return err
	}

	return nil
}

func (s *EmailService) SendInvite(email, inviterName, inviteeName, password string) error {
	messagesInfo := []mailjet.InfoMessagesV31{
		{
			From: &mailjet.RecipientV31{
				Email: s.FromEmail,
				Name:  s.FromName,
			},
			ReplyTo: &mailjet.RecipientV31{
				Email: s.ReplyToEmail,
				Name:  s.FromName,
			}, To: &mailjet.RecipientsV31{
				mailjet.RecipientV31{
					Email: email,
					Name:  inviteeName,
				},
			},
			TemplateID:       5809696,
			TemplateLanguage: true,
			Subject:          fmt.Sprintf("%s has invited you to TODO", inviterName),
			Variables: map[string]interface{}{
				"inviteeName": inviteeName,
				"inviterName": inviterName,
				"password":    password,
				"email":       email,
			},
		},
	}

	sendEmailData := mailjet.MessagesV31{Info: messagesInfo}
	_, err := s.Client.SendMailV31(&sendEmailData)
	if err != nil {
		return err
	}

	return nil
}

func (s *EmailService) BulkSendInvite(inviterName string, req *models.BulkCreateUsersRequest) error {
	messages := []mailjet.InfoMessagesV31{}
	for _, user := range req.Users {
		message := mailjet.InfoMessagesV31{
			From: &mailjet.RecipientV31{
				Email: s.FromEmail,
				Name:  s.FromName,
			},
			ReplyTo: &mailjet.RecipientV31{
				Email: s.ReplyToEmail,
				Name:  s.FromName,
			}, To: &mailjet.RecipientsV31{
				mailjet.RecipientV31{
					Email: user.Email,
					Name:  user.Name,
				},
			},
			TemplateID:       5809696,
			TemplateLanguage: true,
			Subject:          fmt.Sprintf("%s has invited you to TODO", inviterName),
			Variables: map[string]interface{}{
				"inviteeName": user.Name,
				"inviterName": inviterName,
				"password":    user.LoginOtp,
				"email":       user.Email,
			},
		}
		messages = append(messages, message)
	}
	messagesInfo := messages

	sendEmailData := mailjet.MessagesV31{Info: messagesInfo}
	_, err := s.Client.SendMailV31(&sendEmailData)
	if err != nil {
		return err
	}

	return nil
}

func (s *EmailService) SendPasswordResetCode(email, username, code string) error {
	messagesInfo := []mailjet.InfoMessagesV31{
		{
			From: &mailjet.RecipientV31{
				Email: s.FromEmail,
				Name:  s.FromName,
			},
			ReplyTo: &mailjet.RecipientV31{
				Email: s.ReplyToEmail,
				Name:  s.FromName,
			},
			To: &mailjet.RecipientsV31{
				mailjet.RecipientV31{
					Email: email,
					Name:  username,
				},
			},
			TemplateID:       5977750,
			TemplateLanguage: true,
			Subject:          "Password reset code for TODO",
			Variables: map[string]interface{}{
				"name": username,
				"code": code,
			},
		},
	}

	sendEmailData := mailjet.MessagesV31{Info: messagesInfo}
	_, err := s.Client.SendMailV31(&sendEmailData)
	if err != nil {
		return err
	}

	return nil
}
