package email

import (
	"context"
	"time"
)

// Sender envia e-mails (ex.: recuperação de senha).
// Implementações podem usar Resend, SMTP, etc.
type Sender interface {
	SendPasswordReset(ctx context.Context, to, resetToken string, expiresAt time.Time) error
}
