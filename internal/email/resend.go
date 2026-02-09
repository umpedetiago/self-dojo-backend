package email

import (
	"context"
	"fmt"
	"html"
	"time"

	"github.com/resend/resend-go/v2"
)

// ResendSender envia e-mails via Resend.
type ResendSender struct {
	client   *resend.Client
	from     string
	resetURL string // URL base do frontend, ex: https://app.com/reset-password?token=
}

// ResendConfig configura o sender Resend.
type ResendConfig struct {
	APIKey           string
	From             string
	FrontendResetURL string
}

// NewResendSender cria um Sender que usa a API Resend.
func NewResendSender(cfg ResendConfig) (*ResendSender, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("email: RESEND_API_KEY is required")
	}
	client := resend.NewClient(cfg.APIKey)
	from := cfg.From
	if from == "" {
		from = "onboarding@resend.dev"
	}
	return &ResendSender{
		client:   client,
		from:     from,
		resetURL: cfg.FrontendResetURL,
	}, nil
}

// SendPasswordReset envia e-mail com link/token para redefinir a senha.
func (r *ResendSender) SendPasswordReset(ctx context.Context, to, resetToken string, expiresAt time.Time) error {
	subject := "Redefinição de senha"
	expiresStr := expiresAt.Format(time.RFC3339)
	tokenEscaped := html.EscapeString(resetToken)

	var body string
	if r.resetURL != "" {
		link := r.resetURL + resetToken
		linkEscaped := html.EscapeString(link)
		body = fmt.Sprintf(
			`<p>Você solicitou a redefinição de senha.</p>
<p><strong><a href="%s">Clique aqui para redefinir sua senha</a></strong></p>
<p>Ou copie o link e cole no navegador:</p>
<p style="word-break:break-all;"><a href="%s">%s</a></p>
<p>Token (válido até %s), se precisar colar no app:</p>
<pre>%s</pre>
<p>Se não foi você, ignore este e-mail.</p>`,
			linkEscaped,
			linkEscaped, linkEscaped,
			expiresStr,
			tokenEscaped,
		)
	} else {
		body = fmt.Sprintf(
			`<p>Você solicitou a redefinição de senha.</p>
<p>Use o token abaixo no app (válido até %s):</p>
<pre>%s</pre>
<p>Configure <code>FRONTEND_RESET_PASSWORD_URL</code> no servidor para enviar um link clicável.</p>
<p>Se não foi você, ignore este e-mail.</p>`,
			expiresStr,
			tokenEscaped,
		)
	}

	params := &resend.SendEmailRequest{
		From:    r.from,
		To:      []string{to},
		Subject: subject,
		Html:    body,
	}

	_, err := r.client.Emails.Send(params)
	if err != nil {
		return fmt.Errorf("email send: %w", err)
	}
	return nil
}
