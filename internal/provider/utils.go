package provider

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/agilize/terraform-provider-jumpcloud/internal/client"
)

// IsNotFound verifica se o erro indica que o recurso não foi encontrado (HTTP 404)
func IsNotFound(err error) bool {
	if err == nil {
		return false
	}

	// Primeiro, tenta converter para um erro JumpCloud para verificar o código de status HTTP
	if jumpCloudErr, ok := err.(*client.JumpCloudError); ok {
		return jumpCloudErr.StatusCode == http.StatusNotFound
	}

	// Se não for um erro JumpCloud, verifica se a string de erro contém "404"
	return strings.Contains(err.Error(), "404") ||
		strings.Contains(strings.ToLower(err.Error()), "not found")
}

// IsUnauthorized verifica se o erro indica que a requisição não está autorizada (HTTP 401)
func IsUnauthorized(err error) bool {
	if err == nil {
		return false
	}

	if jumpCloudErr, ok := err.(*client.JumpCloudError); ok {
		return jumpCloudErr.StatusCode == http.StatusUnauthorized
	}

	return strings.Contains(strings.ToLower(err.Error()), "unauthorized") ||
		strings.Contains(strings.ToLower(err.Error()), "invalid token")
}

// IsForbidden verifica se o erro indica que a requisição está proibida (HTTP 403)
func IsForbidden(err error) bool {
	if err == nil {
		return false
	}

	if jumpCloudErr, ok := err.(*client.JumpCloudError); ok {
		return jumpCloudErr.StatusCode == http.StatusForbidden
	}

	return strings.Contains(strings.ToLower(err.Error()), "forbidden") ||
		strings.Contains(strings.ToLower(err.Error()), "access denied")
}

// IsConflict verifica se o erro indica um conflito de recursos (HTTP 409)
func IsConflict(err error) bool {
	if err == nil {
		return false
	}

	if jumpCloudErr, ok := err.(*client.JumpCloudError); ok {
		return jumpCloudErr.StatusCode == http.StatusConflict
	}

	return strings.Contains(strings.ToLower(err.Error()), "conflict") ||
		strings.Contains(strings.ToLower(err.Error()), "already exists")
}

// IsBadRequest verifica se o erro indica uma requisição inválida (HTTP 400)
func IsBadRequest(err error) bool {
	if err == nil {
		return false
	}

	if jumpCloudErr, ok := err.(*client.JumpCloudError); ok {
		return jumpCloudErr.StatusCode == http.StatusBadRequest
	}

	return strings.Contains(strings.ToLower(err.Error()), "bad request") ||
		strings.Contains(strings.ToLower(err.Error()), "invalid request")
}

// IsRetryable verifica se o erro é temporário e a operação pode ser tentada novamente
func IsRetryable(err error) bool {
	if err == nil {
		return false
	}

	if jumpCloudErr, ok := err.(*client.JumpCloudError); ok {
		// Códigos 5xx indicam erro do servidor e podem ser tentados novamente
		return jumpCloudErr.StatusCode >= 500 && jumpCloudErr.StatusCode < 600
	}

	return strings.Contains(strings.ToLower(err.Error()), "timeout") ||
		strings.Contains(strings.ToLower(err.Error()), "temporary") ||
		strings.Contains(strings.ToLower(err.Error()), "retry")
}

// FormatISOTime converte um timestamp ISO8601 para o formato RFC3339 usado pelo Terraform
func FormatISOTime(isoTime string) string {
	if isoTime == "" {
		return ""
	}

	// Tentar vários formatos de data conhecidos
	formats := []string{
		time.RFC3339,
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05-07:00",
		"2006-01-02T15:04:05.999Z",
		"2006-01-02T15:04:05.999-07:00",
	}

	for _, format := range formats {
		t, err := time.Parse(format, isoTime)
		if err == nil {
			return t.Format(time.RFC3339)
		}
	}

	// Se não conseguir parsear, retornar a string original
	return isoTime
}

// ValidateRFC3339 verifica se uma string está no formato RFC3339 e é uma data futura
func ValidateRFC3339FutureDate(v string) error {
	if v == "" {
		return nil
	}

	t, err := time.Parse(time.RFC3339, v)
	if err != nil {
		return fmt.Errorf("formato de data inválido: %v (deve estar no formato RFC3339)", err)
	}

	if t.Before(time.Now()) {
		return fmt.Errorf("a data deve estar no futuro")
	}

	return nil
}

// ExtractErrorDetails extrai informações detalhadas do erro
func ExtractErrorDetails(err error) string {
	if err == nil {
		return ""
	}

	// Se for um erro JumpCloud, extrair campos específicos
	if jumpCloudErr, ok := err.(*client.JumpCloudError); ok {
		if jumpCloudErr.Message != "" {
			return fmt.Sprintf("%s (Código: %s, Status: %d)",
				jumpCloudErr.Message,
				jumpCloudErr.Code,
				jumpCloudErr.StatusCode)
		}
	}

	// Caso contrário, retornar a mensagem de erro completa
	return err.Error()
}
