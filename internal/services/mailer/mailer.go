package mailer

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/wneessen/go-mail"
	"str-prof-bot/internal/config"
	"str-prof-bot/internal/domain"
	"str-prof-bot/internal/lib/logger/sl"
)

type Service struct {
	client *mail.Client
	from   string
	to     string
	log    *slog.Logger
}

func New(log *slog.Logger, cfg config.SMTPConfig) (*Service, error) {
	options := []mail.Option{
		mail.WithPort(cfg.Port),
		mail.WithSMTPAuth(mail.SMTPAuthAutoDiscover),
		mail.WithUsername(cfg.Username),
		mail.WithPassword(cfg.Password),
		mail.WithTimeout(10 * time.Second),
	}
	if cfg.Port == 465 {
		options = append(options, mail.WithSSL())
	} else {
		options = append(options, mail.WithTLSPolicy(mail.TLSMandatory))
	}

	client, err := mail.NewClient(cfg.Host, options...)
	if err != nil {
		return nil, fmt.Errorf("create SMTP client: %w", err)
	}
	return &Service{client: client, from: cfg.From, to: cfg.CompanyEmail, log: log}, nil
}

func (s *Service) SendOrder(ctx context.Context, order domain.Order) error {
	message := mail.NewMsg()
	if err := message.From(s.from); err != nil {
		return fmt.Errorf("set sender: %w", err)
	}
	if err := message.To(s.to); err != nil {
		return fmt.Errorf("set recipient: %w", err)
	}
	message.Subject("Заявка из Telegram")
	message.SetBodyString(mail.TypeTextPlain, orderText(order))
	if err := s.client.DialAndSendWithContext(ctx, message); err != nil {
		s.log.Error("failed to send order email", sl.Err(err), "order_id", order.ID)
		return fmt.Errorf("send order email: %w", err)
	}
	s.log.Info("order email sent", "order_id", order.ID, "recipient", s.to)
	return nil
}

func orderText(order domain.Order) string {
	var builder strings.Builder
	fmt.Fprintf(&builder, "Заявка\nДата: %s\n\n", order.CreatedAt.Format("02.01.2006 15:04"))
	fmt.Fprintf(&builder, "Клиент: %s %s\nТелефон: %s\nEmail: %s\n\nТовары:\n", order.User.LastName, order.User.FirstName, order.User.Phone, order.User.Email)
	for _, item := range order.Items {
		fmt.Fprintf(&builder, "- %s, ёмкость %s — %d шт. × %s\n", item.Name, domain.CapacityLabel(item.Capacity), item.Quantity, rubles(item.Price))
	}
	fmt.Fprintf(&builder, "\nИтого: %s", rubles(order.Total))
	return builder.String()
}

func rubles(kopecks int64) string { return fmt.Sprintf("%.2f ₽", float64(kopecks)/100) }
