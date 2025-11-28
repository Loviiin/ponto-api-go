package email

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewEmailService(t *testing.T) {
	// Test data
	host := "smtp.example.com"
	port := "587"
	user := "user@example.com"
	password := "password"
	fromName := "Test Sender"
	fromEmail := "sender@example.com"
	frontendURL := "http://localhost:3000"

	service := NewEmailService(host, port, user, password, fromName, fromEmail, frontendURL)

	assert.NotNil(t, service)
	assert.Equal(t, "smtp.example.com", service.host)
	assert.Equal(t, 587, service.port)
	assert.Equal(t, "user@example.com", service.username)
	assert.Equal(t, "password", service.password)
	assert.Equal(t, "Test Sender", service.fromName)
	assert.Equal(t, "sender@example.com", service.fromEmail)
	assert.Equal(t, "http://localhost:3000", service.frontendURL)
}

func TestGetTemplate(t *testing.T) {
	tests := []struct {
		name     string
		tmplName string
		want     string
	}{
		{
			name:     "Password Reset Template",
			tmplName: TemplatePasswordReset,
			want:     "Recuperação de Senha",
		},
		{
			name:     "Welcome Template",
			tmplName: TemplateWelcome,
			want:     "Bem-vindo ao Nexora",
		},
		{
			name:     "Unknown Template",
			tmplName: "unknown",
			want:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetTemplate(tt.tmplName)
			if tt.want == "" {
				assert.Empty(t, got)
			} else {
				assert.Contains(t, got, tt.want)
			}
		})
	}
}

func TestEmailService_SendEmail_InvalidTemplate(t *testing.T) {
	service := NewEmailService("smtp.example.com", "587", "user@example.com", "password", "Test", "test@example.com", "http://localhost:3000")
	err := service.SendEmail([]string{"test@example.com"}, "Subject", "invalid_template", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse template")
}
