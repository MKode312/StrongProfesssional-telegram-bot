package orders

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"str-prof-bot/internal/domain"
	"str-prof-bot/internal/services/mailer"
	"str-prof-bot/internal/storage/postgres"
)

var emailPattern = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)

type Service struct {
	storage *postgres.Storage
	mailer  *mailer.Service
}

func New(storage *postgres.Storage, sender *mailer.Service) *Service {
	return &Service{storage: storage, mailer: sender}
}
func (s *Service) Products(ctx context.Context) ([]domain.Product, error) {
	return s.storage.Products(ctx)
}

func (s *Service) ProductsByCategory(ctx context.Context, category string) ([]domain.Product, error) {
	return s.storage.ProductsByCategory(ctx, category)
}
func (s *Service) SearchProducts(ctx context.Context, keyword string) ([]domain.Product, error) {
	keyword = strings.TrimSpace(keyword)
	if keyword == "" {
		return []domain.Product{}, nil
	}
	return s.storage.SearchProducts(ctx, keyword)
}
func (s *Service) AddToCart(ctx context.Context, userID, productID int64, capacity string, quantity int) error {
	return s.storage.AddToCart(ctx, userID, productID, capacity, quantity)
}
func (s *Service) ChangeCartItem(ctx context.Context, userID, productID int64, capacity string, delta int) error {
	return s.storage.ChangeCartItem(ctx, userID, productID, capacity, delta)
}
func (s *Service) Cart(ctx context.Context, userID int64) ([]domain.CartItem, error) {
	return s.storage.Cart(ctx, userID)
}
func (s *Service) User(ctx context.Context, userID int64) (domain.User, bool, error) {
	return s.storage.User(ctx, userID)
}

func (s *Service) SaveUser(ctx context.Context, user domain.User) error {
	if !emailPattern.MatchString(user.Email) {
		return fmt.Errorf("invalid email")
	}
	if user.Phone == "" || user.FirstName == "" || user.LastName == "" {
		return fmt.Errorf("profile fields cannot be empty")
	}
	return s.storage.SaveUser(ctx, user)
}

func (s *Service) Checkout(ctx context.Context, user domain.User) (domain.Order, error) {
	order, err := s.storage.CreateOrder(ctx, user)
	if err != nil {
		return domain.Order{}, err
	}
	if err := s.mailer.SendOrder(ctx, order); err != nil {
		return domain.Order{}, err
	}
	return order, nil
}
