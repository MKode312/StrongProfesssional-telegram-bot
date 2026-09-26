package handlers

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"sync"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"str-prof-bot/internal/domain"
	"str-prof-bot/internal/lib/logger/sl"
	"str-prof-bot/internal/services/orders"
)

const (
	callbackMenu     = "menu"
	callbackCatalog  = "catalog"
	callbackSearch   = "search"
	callbackCart     = "cart"
	callbackCheckout = "checkout"
	callbackContacts = "contacts"
)

var capacities = []string{"0.5", "1", "3", "5", "10", "20"}
var quantities = []int{1, 2, 5, 10}

var catalogCategories = []string{"degreaser", "solvent"}

var categoryLabels = map[string]string{
	"degreaser": "Обезжириватель",
	"solvent":   "Растворитель",
}

type profileDraft struct {
	step     int
	user     domain.User
	checkout bool
	field    string
}

type quantityDraft struct {
	productID int64
	capacity  string
}

type Handler struct {
	log        *slog.Logger
	orders     *orders.Service
	mu         sync.Mutex
	drafts     map[int64]profileDraft
	searches   map[int64]bool
	quantities map[int64]quantityDraft
}

func New(log *slog.Logger, service *orders.Service) *Handler {
	return &Handler{
		log:        log,
		orders:     service,
		drafts:     make(map[int64]profileDraft),
		searches:   make(map[int64]bool),
		quantities: make(map[int64]quantityDraft),
	}
}

func (h *Handler) Handle(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.CallbackQuery != nil {
		h.handleCallback(ctx, b, update.CallbackQuery)
		return
	}
	if update.Message == nil || update.Message.From == nil || update.Message.Text == "" {
		return
	}

	chatID, userID := update.Message.Chat.ID, update.Message.From.ID
	if h.handleProfile(ctx, b, chatID, userID, update.Message.Text) {
		return
	}
	if h.handleSearch(ctx, b, chatID, userID, update.Message.Text) {
		return
	}
	if h.handleQuantity(ctx, b, chatID, userID, update.Message.Text) {
		return
	}

	switch update.Message.Text {
	case "/start":
		h.showMainMenu(ctx, b, chatID, true)
	case "/menu":
		h.showMainMenu(ctx, b, chatID, false)
	case "/catalog":
		h.showCatalog(ctx, b, chatID)
	case "/cart":
		h.showCart(ctx, b, chatID, userID)
	case "/help":
		h.send(ctx, b, chatID, "Откройте /menu, чтобы сделать заказ или изменить контактные данные.")
	default:
		h.showMainMenu(ctx, b, chatID, false)
	}
}

func (h *Handler) handleCallback(ctx context.Context, b *bot.Bot, query *models.CallbackQuery) {
	if query.Message.Message == nil {
		return
	}
	_, _ = b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{CallbackQueryID: query.ID})

	chatID, userID, data := query.Message.Message.Chat.ID, query.From.ID, query.Data
	h.clearTextInput(userID)
	switch {
	case data == callbackMenu:
		h.showMainMenu(ctx, b, chatID, false)
	case data == callbackCatalog:
		h.showCatalog(ctx, b, chatID)
	case data == callbackSearch:
		h.startSearch(ctx, b, chatID, userID)
	case data == callbackCart:
		h.showCart(ctx, b, chatID, userID)
	case data == callbackCheckout:
		h.checkout(ctx, b, chatID, userID)
	case data == callbackContacts:
		h.showContacts(ctx, b, chatID, userID)
	case strings.HasPrefix(data, "category:"):
		category := strings.TrimPrefix(data, "category:")
		if categoryLabels[category] != "" {
			h.showCategoryProducts(ctx, b, chatID, category)
		}
	case strings.HasPrefix(data, "product:"):
		productID, err := callbackID(data, "product:")
		if err == nil {
			h.showCapacities(ctx, b, chatID, productID)
		}
	case strings.HasPrefix(data, "capacity:"):
		parts := strings.SplitN(data, ":", 3)
		if len(parts) == 3 {
			productID, err := strconv.ParseInt(parts[1], 10, 64)
			if err == nil {
				h.showQuantities(ctx, b, chatID, productID, parts[2])
			}
		}
	case strings.HasPrefix(data, "quantity:"):
		parts := strings.SplitN(data, ":", 4)
		if len(parts) == 4 {
			productID, err := strconv.ParseInt(parts[1], 10, 64)
			quantity, quantityOK := positiveInteger(parts[3])
			if err == nil && quantityOK {
				h.showConfirmation(ctx, b, chatID, productID, parts[2], quantity)
			}
		}
	case strings.HasPrefix(data, "customqty:"):
		parts := strings.SplitN(data, ":", 3)
		if len(parts) == 3 {
			productID, err := strconv.ParseInt(parts[1], 10, 64)
			if err == nil {
				h.startQuantityInput(ctx, b, chatID, userID, productID, parts[2])
			}
		}
	case strings.HasPrefix(data, "add:"):
		h.addToCart(ctx, b, chatID, userID, data)
	case strings.HasPrefix(data, "cartqty:"):
		h.changeCartItem(ctx, b, chatID, userID, data)
	case strings.HasPrefix(data, "contact:"):
		h.startContactEdit(ctx, b, chatID, userID, strings.TrimPrefix(data, "contact:"))
	}
}

func (h *Handler) showMainMenu(ctx context.Context, b *bot.Bot, chatID int64, greeting bool) {
	text := "Главное меню"
	if greeting {
		text = "Вас приветствует бот StrongProfessional. Чтобы сделать заказ нажмите на «Сделать заказ». Чтобы изменить контактные данные, нажмите на «Изменить контактные данные»."
	}
	h.sendMarkup(ctx, b, chatID, text, [][]models.InlineKeyboardButton{
		{{Text: "Сделать заказ", CallbackData: callbackCatalog}},
		{{Text: "Изменить контактные данные", CallbackData: callbackContacts}},
	})
}

func (h *Handler) showCatalog(ctx context.Context, b *bot.Bot, chatID int64) {
	rows := make([][]models.InlineKeyboardButton, 0, len(catalogCategories)+3)
	rows = append(rows, []models.InlineKeyboardButton{{Text: "🔎 Найти товар", CallbackData: callbackSearch}})
	for _, category := range catalogCategories {
		rows = append(rows, []models.InlineKeyboardButton{{Text: categoryLabels[category], CallbackData: fmt.Sprintf("category:%s", category)}})
	}
	rows = append(rows, []models.InlineKeyboardButton{{Text: "🛒 Корзина", CallbackData: callbackCart}})
	rows = append(rows, []models.InlineKeyboardButton{{Text: "Главное меню", CallbackData: callbackMenu}})
	h.sendMarkup(ctx, b, chatID, "Каталог товаров\nВыберите категорию.", rows)
}

func (h *Handler) startSearch(ctx context.Context, b *bot.Bot, chatID, userID int64) {
	h.mu.Lock()
	h.searches[userID] = true
	h.mu.Unlock()
	h.send(ctx, b, chatID, "Введите название желаемого товара или его часть.")
}

func (h *Handler) handleSearch(ctx context.Context, b *bot.Bot, chatID, userID int64, value string) bool {
	h.mu.Lock()
	waiting := h.searches[userID]
	h.mu.Unlock()
	if !waiting {
		return false
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/") {
		h.mu.Lock()
		delete(h.searches, userID)
		h.mu.Unlock()
		return false
	}

	keyword := strings.TrimSpace(value)
	if keyword == "" {
		h.send(ctx, b, chatID, "Название не должно быть пустым. Введите ключевое слово ещё раз.")
		return true
	}

	products, err := h.orders.SearchProducts(ctx, keyword)
	if err != nil {
		h.fail(ctx, b, chatID, err)
		return true
	}
	h.mu.Lock()
	delete(h.searches, userID)
	h.mu.Unlock()

	if len(products) == 0 {
		h.sendMarkup(ctx, b, chatID, fmt.Sprintf("По запросу «%s» ничего не найдено.", keyword), [][]models.InlineKeyboardButton{
			{{Text: "Искать ещё", CallbackData: callbackSearch}},
			{{Text: "Назад к каталогу", CallbackData: callbackCatalog}},
		})
		return true
	}

	rows := make([][]models.InlineKeyboardButton, 0, len(products)+2)
	for _, product := range products {
		rows = append(rows, []models.InlineKeyboardButton{{Text: product.Name, CallbackData: fmt.Sprintf("product:%d", product.ID)}})
	}
	rows = append(rows, []models.InlineKeyboardButton{{Text: "Искать ещё", CallbackData: callbackSearch}})
	rows = append(rows, []models.InlineKeyboardButton{{Text: "Назад к каталогу", CallbackData: callbackCatalog}})
	h.sendMarkup(ctx, b, chatID, fmt.Sprintf("Результаты поиска по запросу «%s». Выберите товар.", keyword), rows)
	return true
}

func (h *Handler) showCategoryProducts(ctx context.Context, b *bot.Bot, chatID int64, category string) {
	products, err := h.orders.ProductsByCategory(ctx, category)
	if err != nil {
		h.fail(ctx, b, chatID, err)
		return
	}
	label := categoryLabels[category]
	rows := make([][]models.InlineKeyboardButton, 0, len(products)+2)
	for _, product := range products {
		rows = append(rows, []models.InlineKeyboardButton{{Text: product.Name, CallbackData: fmt.Sprintf("product:%d", product.ID)}})
	}
	rows = append(rows, []models.InlineKeyboardButton{{Text: "Назад к категориям", CallbackData: callbackCatalog}})
	rows = append(rows, []models.InlineKeyboardButton{{Text: "🛒 Корзина", CallbackData: callbackCart}, {Text: "Главное меню", CallbackData: callbackMenu}})
	h.sendMarkup(ctx, b, chatID, fmt.Sprintf("%s\nВыберите товар.", label), rows)
}

func (h *Handler) showCapacities(ctx context.Context, b *bot.Bot, chatID, productID int64) {
	product, err := h.product(ctx, productID)
	if err != nil {
		h.fail(ctx, b, chatID, err)
		return
	}
	rows := make([][]models.InlineKeyboardButton, 0, len(capacities)+1)
	for _, capacity := range capacities {
		rows = append(rows, []models.InlineKeyboardButton{{Text: domain.CapacityLabel(capacity), CallbackData: fmt.Sprintf("capacity:%d:%s", product.ID, capacity)}})
	}
	backData := callbackCatalog
	if product.Category != "" {
		backData = fmt.Sprintf("category:%s", product.Category)
	}
	rows = append(rows, []models.InlineKeyboardButton{{Text: "Назад к товарам", CallbackData: backData}})
	h.sendMarkup(ctx, b, chatID, fmt.Sprintf("%s\nВыберите ёмкость.", product.Name), rows)
}

func (h *Handler) showQuantities(ctx context.Context, b *bot.Bot, chatID, productID int64, capacity string) {
	rows := make([][]models.InlineKeyboardButton, 0, len(quantities)+2)
	for _, quantity := range quantities {
		rows = append(rows, []models.InlineKeyboardButton{{Text: fmt.Sprintf("%d шт.", quantity), CallbackData: fmt.Sprintf("quantity:%d:%s:%d", productID, capacity, quantity)}})
	}
	rows = append(rows, []models.InlineKeyboardButton{{Text: "Ввести количество", CallbackData: fmt.Sprintf("customqty:%d:%s", productID, capacity)}})
	rows = append(rows, []models.InlineKeyboardButton{{Text: "Назад к ёмкостям", CallbackData: fmt.Sprintf("product:%d", productID)}})
	h.sendMarkup(ctx, b, chatID, "Выберите количество штук.", rows)
}

func (h *Handler) startQuantityInput(ctx context.Context, b *bot.Bot, chatID, userID, productID int64, capacity string) {
	h.mu.Lock()
	h.quantities[userID] = quantityDraft{productID: productID, capacity: capacity}
	h.mu.Unlock()
	h.send(ctx, b, chatID, "Введите количество штук целым положительным числом.")
}

func (h *Handler) handleQuantity(ctx context.Context, b *bot.Bot, chatID, userID int64, value string) bool {
	h.mu.Lock()
	draft, waiting := h.quantities[userID]
	h.mu.Unlock()
	if !waiting {
		return false
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/") {
		h.mu.Lock()
		delete(h.quantities, userID)
		h.mu.Unlock()
		return false
	}

	quantity, ok := positiveInteger(value)
	if !ok {
		h.send(ctx, b, chatID, "Количество должно быть целым числом больше нуля. Попробуйте ещё раз.")
		return true
	}
	h.mu.Lock()
	delete(h.quantities, userID)
	h.mu.Unlock()
	h.showConfirmation(ctx, b, chatID, draft.productID, draft.capacity, quantity)
	return true
}

func (h *Handler) showConfirmation(ctx context.Context, b *bot.Bot, chatID, productID int64, capacity string, quantity int) {
	product, err := h.product(ctx, productID)
	if err != nil {
		h.fail(ctx, b, chatID, err)
		return
	}
	text := fmt.Sprintf("Вы выбрали:\n%s\n%s\nЁмкость: %s\nКоличество: %d шт.\nЦена: %s за штуку", product.Name, product.Description, domain.CapacityLabel(capacity), quantity, money(product.Price))
	rows := [][]models.InlineKeyboardButton{
		{{Text: "Добавить в корзину", CallbackData: fmt.Sprintf("add:%d:%s:%d", productID, capacity, quantity)}},
		{{Text: "Отменить", CallbackData: callbackCatalog}},
	}
	h.sendMarkup(ctx, b, chatID, text, rows)
}

func (h *Handler) addToCart(ctx context.Context, b *bot.Bot, chatID, userID int64, data string) {
	parts := strings.SplitN(data, ":", 4)
	if len(parts) != 4 {
		return
	}
	productID, err := strconv.ParseInt(parts[1], 10, 64)
	quantity, quantityOK := positiveInteger(parts[3])
	if err != nil || !quantityOK {
		return
	}
	if err := h.orders.AddToCart(ctx, userID, productID, parts[2], quantity); err != nil {
		h.fail(ctx, b, chatID, err)
		return
	}
	h.log.Info("cart item added", "telegram_id", userID, "product_id", productID, "capacity", parts[2], "quantity", quantity)
	h.sendMarkup(ctx, b, chatID, "Товар добавлен в корзину.", [][]models.InlineKeyboardButton{{{Text: "Продолжить покупки", CallbackData: callbackCatalog}, {Text: "Корзина", CallbackData: callbackCart}}})
}

func (h *Handler) showCart(ctx context.Context, b *bot.Bot, chatID, userID int64) {
	items, err := h.orders.Cart(ctx, userID)
	if err != nil {
		h.fail(ctx, b, chatID, err)
		return
	}
	if len(items) == 0 {
		h.sendMarkup(ctx, b, chatID, "Корзина пока пуста.", [][]models.InlineKeyboardButton{{{Text: "Каталог", CallbackData: callbackCatalog}, {Text: "Главное меню", CallbackData: callbackMenu}}})
		return
	}
	h.log.Debug("cart loaded", "telegram_id", userID, "items", len(items))

	var text strings.Builder
	text.WriteString("Корзина:\n")
	rows := make([][]models.InlineKeyboardButton, 0, len(items)+2)
	total := int64(0)
	for _, item := range items {
		fmt.Fprintf(&text, "• %s, %s — %d шт. × %s\n", item.Name, domain.CapacityLabel(item.Capacity), item.Quantity, money(item.Price))
		total += item.Price * int64(item.Quantity)
		rows = append(rows, []models.InlineKeyboardButton{
			{Text: "−", CallbackData: fmt.Sprintf("cartqty:%d:%s:-1", item.ID, item.Capacity)},
			{Text: "+", CallbackData: fmt.Sprintf("cartqty:%d:%s:1", item.ID, item.Capacity)},
		})
	}
	fmt.Fprintf(&text, "\nИтого: %s", money(total))
	rows = append(rows, []models.InlineKeyboardButton{{Text: "Оформить заявку", CallbackData: callbackCheckout}, {Text: "Каталог", CallbackData: callbackCatalog}})
	h.sendMarkup(ctx, b, chatID, text.String(), rows)
}

func (h *Handler) changeCartItem(ctx context.Context, b *bot.Bot, chatID, userID int64, data string) {
	parts := strings.SplitN(data, ":", 4)
	if len(parts) != 4 {
		return
	}
	productID, err := strconv.ParseInt(parts[1], 10, 64)
	delta, deltaErr := strconv.Atoi(parts[3])
	if err != nil || deltaErr != nil {
		return
	}
	if err := h.orders.ChangeCartItem(ctx, userID, productID, parts[2], delta); err != nil {
		h.fail(ctx, b, chatID, err)
		return
	}
	h.showCart(ctx, b, chatID, userID)
}

func (h *Handler) showContacts(ctx context.Context, b *bot.Bot, chatID, userID int64) {
	user, exists, err := h.orders.User(ctx, userID)
	if err != nil {
		h.fail(ctx, b, chatID, err)
		return
	}
	if !exists {
		h.sendMarkup(ctx, b, chatID, "Контактные данные появятся после первого оформления заказа.", [][]models.InlineKeyboardButton{{{Text: "Сделать заказ", CallbackData: callbackCatalog}, {Text: "Главное меню", CallbackData: callbackMenu}}})
		return
	}
	text := fmt.Sprintf("Ваши контактные данные:\nEmail: %s\nТелефон: %s\nФамилия: %s\nИмя: %s\n\nВыберите поле для изменения.", user.Email, user.Phone, user.LastName, user.FirstName)
	h.sendMarkup(ctx, b, chatID, text, [][]models.InlineKeyboardButton{
		{{Text: "Email", CallbackData: "contact:email"}, {Text: "Телефон", CallbackData: "contact:phone"}},
		{{Text: "Фамилия", CallbackData: "contact:last_name"}, {Text: "Имя", CallbackData: "contact:first_name"}},
		{{Text: "Главное меню", CallbackData: callbackMenu}},
	})
}

func (h *Handler) startContactEdit(ctx context.Context, b *bot.Bot, chatID, userID int64, field string) {
	if field != "email" && field != "phone" && field != "last_name" && field != "first_name" {
		return
	}
	user, exists, err := h.orders.User(ctx, userID)
	if err != nil {
		h.fail(ctx, b, chatID, err)
		return
	}
	if !exists {
		h.showContacts(ctx, b, chatID, userID)
		return
	}
	h.mu.Lock()
	h.drafts[userID] = profileDraft{user: user, field: field}
	h.mu.Unlock()
	prompts := map[string]string{"email": "Введите новый email.", "phone": "Введите новый номер телефона.", "last_name": "Введите новую фамилию.", "first_name": "Введите новое имя."}
	h.send(ctx, b, chatID, prompts[field])
}

func (h *Handler) checkout(ctx context.Context, b *bot.Bot, chatID, userID int64) {
	user, exists, err := h.orders.User(ctx, userID)
	if err != nil {
		h.fail(ctx, b, chatID, err)
		return
	}
	if !exists {
		h.mu.Lock()
		h.drafts[userID] = profileDraft{user: domain.User{TelegramID: userID}, checkout: true}
		h.mu.Unlock()
		h.send(ctx, b, chatID, "Первый заказ: укажите вашу электронную почту.")
		return
	}
	h.sendOrder(ctx, b, chatID, user)
}

func (h *Handler) handleProfile(ctx context.Context, b *bot.Bot, chatID, userID int64, value string) bool {
	h.mu.Lock()
	draft, ok := h.drafts[userID]
	h.mu.Unlock()
	if !ok {
		return false
	}

	value = strings.TrimSpace(value)
	if value == "" {
		h.send(ctx, b, chatID, "Поле не должно быть пустым. Попробуйте ещё раз.")
		return true
	}
	if draft.field != "" {
		switch draft.field {
		case "email":
			draft.user.Email = value
		case "phone":
			draft.user.Phone = value
		case "last_name":
			draft.user.LastName = value
		case "first_name":
			draft.user.FirstName = value
		}
		if err := h.orders.SaveUser(ctx, draft.user); err != nil {
			h.send(ctx, b, chatID, "Не удалось сохранить значение. Проверьте формат и введите его ещё раз.")
			return true
		}
		h.mu.Lock()
		delete(h.drafts, userID)
		h.mu.Unlock()
		h.send(ctx, b, chatID, "Контактные данные обновлены.")
		h.showContacts(ctx, b, chatID, userID)
		return true
	}

	switch draft.step {
	case 0:
		draft.user.Email = value
		draft.step = 1
		h.send(ctx, b, chatID, "Укажите номер телефона.")
	case 1:
		draft.user.Phone = value
		draft.step = 2
		h.send(ctx, b, chatID, "Укажите фамилию.")
	case 2:
		draft.user.LastName = value
		draft.step = 3
		h.send(ctx, b, chatID, "Укажите имя.")
	case 3:
		draft.user.FirstName = value
		if err := h.orders.SaveUser(ctx, draft.user); err != nil {
			h.send(ctx, b, chatID, "Некорректные данные. Введите электронную почту ещё раз.")
			draft.step = 0
			draft.user.Email = ""
			break
		}
		h.mu.Lock()
		delete(h.drafts, userID)
		h.mu.Unlock()
		if draft.checkout {
			h.sendOrder(ctx, b, chatID, draft.user)
		} else {
			h.showMainMenu(ctx, b, chatID, false)
		}
		return true
	}

	h.mu.Lock()
	h.drafts[userID] = draft
	h.mu.Unlock()
	return true
}

func (h *Handler) product(ctx context.Context, productID int64) (domain.Product, error) {
	products, err := h.orders.Products(ctx)
	if err != nil {
		return domain.Product{}, err
	}
	for _, product := range products {
		if product.ID == productID {
			return product, nil
		}
	}
	return domain.Product{}, fmt.Errorf("product %d not found", productID)
}

func (h *Handler) sendOrder(ctx context.Context, b *bot.Bot, chatID int64, user domain.User) {
	_, err := h.orders.Checkout(ctx, user)
	if err != nil {
		h.fail(ctx, b, chatID, err)
		return
	}
	h.send(ctx, b, chatID, "Заявка отправлена в компанию. Мы свяжемся с вами в ближайшее время.")
}

func (h *Handler) send(ctx context.Context, b *bot.Bot, chatID int64, text string) {
	_, _ = b.SendMessage(ctx, &bot.SendMessageParams{ChatID: chatID, Text: text})
}

func (h *Handler) sendMarkup(ctx context.Context, b *bot.Bot, chatID int64, text string, rows [][]models.InlineKeyboardButton) {
	_, _ = b.SendMessage(ctx, &bot.SendMessageParams{ChatID: chatID, Text: text, ReplyMarkup: &models.InlineKeyboardMarkup{InlineKeyboard: rows}})
}

func (h *Handler) fail(ctx context.Context, b *bot.Bot, chatID int64, err error) {
	h.log.Error("request failed", sl.Err(err), "chat_id", chatID)
	h.send(ctx, b, chatID, "Не удалось выполнить операцию. Попробуйте ещё раз позднее.")
}

func (h *Handler) clearTextInput(userID int64) {
	h.mu.Lock()
	delete(h.drafts, userID)
	delete(h.searches, userID)
	delete(h.quantities, userID)
	h.mu.Unlock()
}

func callbackID(data, prefix string) (int64, error) {
	return strconv.ParseInt(strings.TrimPrefix(data, prefix), 10, 64)
}

func positiveInteger(value string) (int, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, false
	}
	for _, char := range value {
		if char < '0' || char > '9' {
			return 0, false
		}
	}
	parsed, err := strconv.ParseInt(value, 10, 32)
	quantity := int(parsed)
	if err != nil || quantity <= 0 {
		return 0, false
	}
	return quantity, true
}

func money(kopecks int64) string {
	return fmt.Sprintf("%.2f ₽", float64(kopecks)/100)
}
