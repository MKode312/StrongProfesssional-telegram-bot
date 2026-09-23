package domain

import "time"

var capacityLabels = map[string]string{
	"0.5": "0,5 л",
	"1":   "1 л",
	"3":   "3 л",
	"5":   "5 л",
	"10":  "10 л",
	"20":  "20 л",
}

func CapacityLabel(capacity string) string {
	if label, ok := capacityLabels[capacity]; ok {
		return label
	}
	return capacity
}

type Product struct {
	ID          int64
	Name        string
	Description string
	Price       int64 // kopecks
	Category    string
}

type User struct {
	TelegramID int64
	Email      string
	Phone      string
	LastName   string
	FirstName  string
}

type CartItem struct {
	Product
	Capacity string
	Quantity int
}

type Order struct {
	ID        int64
	User      User
	Items     []CartItem
	Total     int64
	CreatedAt time.Time
}
