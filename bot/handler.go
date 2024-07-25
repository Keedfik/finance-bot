package bot

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"finance-bot/config"
	"finance-bot/db"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	StateIdle = iota
	StateAddingExpenseCategory
	StateAddingExpenseAmount
	StateAddingExpenseNote
	StateAddingCategoryName
	StateAddingCategoryLimit
	StateSettingCategoryLimit
	StateSettingCategoryLimitAmount
	StateDeletingLastExpense
)

type UserState struct {
	State      int
	Category   string
	Amount     float64
	CategoryID primitive.ObjectID
}

var userStates = make(map[int64]*UserState)

func getUserState(userID int64) *UserState {
	if state, exists := userStates[userID]; exists {
		return state
	}
	state := &UserState{State: StateIdle}
	userStates[userID] = state
	return state
}

type BotHandler struct {
	Bot *tgbotapi.BotAPI
	DB  *db.MongoDB
	Cfg *config.Config
}

func NewBotHandler(botToken string, db *db.MongoDB, cfg *config.Config) *BotHandler {
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Panic(err)
	}

	log.Println("Bot created successfully")
	return &BotHandler{
		Bot: bot,
		DB:  db,
		Cfg: cfg,
	}
}

func (h *BotHandler) createKeyboard() tgbotapi.ReplyKeyboardMarkup {
	log.Println("Creating keyboard")
	return tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🏠 Start"),
			tgbotapi.NewKeyboardButton("➕ Add Expense"),
			tgbotapi.NewKeyboardButton("📋 Get Expenses"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("❓ Help"),
		),
	)
}

func (h *BotHandler) extractCommand(text string) string {
	text = strings.TrimSpace(text)
	words := strings.Fields(text)
	if len(words) > 0 {
		switch words[0] {
		case "🏠":
			return "/start"
		case "➕":
			return "/addexpense"
		case "📋":
			return "/getexpenses"
		case "❓":
			return "/help"
		}
	}
	return text
}

func (h *BotHandler) HandleUpdate(update tgbotapi.Update) {
	if update.Message != nil {
		log.Printf("Received message: %s", update.Message.Text)
		userID := update.Message.Chat.ID
		userState := getUserState(userID)

		switch userState.State {
		case StateIdle:
			h.handleCommand(update.Message)
		case StateAddingExpenseCategory:
			h.handleAddExpenseCategoryInput(update.Message)
		case StateAddingExpenseAmount:
			h.handleAddExpenseAmountInput(update.Message)
		case StateAddingExpenseNote:
			h.handleAddExpenseNoteInput(update.Message)
		case StateAddingCategoryName:
			h.handleAddCategoryName(update.Message)
		case StateAddingCategoryLimit:
			h.handleAddCategoryLimit(update.Message)
		case StateSettingCategoryLimit:
			h.handleSetCategoryLimit(update.Message)
		case StateSettingCategoryLimitAmount:
			h.handleSetCategoryLimitAmount(update.Message)
		case StateDeletingLastExpense:
			if strings.ToLower(update.Message.Text) == "да" {
				h.handleDeleteLastExpense(update.Message)
			} else {
				h.Bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Отмена удаления."))
				userState.State = StateIdle
			}
		}
	}
}

func (h *BotHandler) handleCommand(msg *tgbotapi.Message) {
	command := h.extractCommand(msg.Text)
	switch {
	case strings.HasPrefix(command, "/start"):
		h.handleStart(msg)
	case strings.HasPrefix(command, "/addexpense"):
		h.promptAddExpenseCategory(msg)
	case strings.HasPrefix(command, "/addcategory"):
		h.promptAddCategoryName(msg)
	case strings.HasPrefix(command, "/setlimit"):
		h.promptSetCategoryLimit(msg)
	case strings.HasPrefix(command, "/getcategories"):
		h.handleGetCategories(msg)
	case strings.HasPrefix(command, "/getexpenses"):
		h.handleGetExpenses(msg)
	case strings.HasPrefix(command, "/deletelastexpense"):
		h.promptDeleteLastExpense(msg)
	case strings.HasPrefix(command, "/help"):
		h.handleHelp(msg)
	default:
		h.handleUnknownCommand(msg)
	}
}

func (h *BotHandler) handleStart(msg *tgbotapi.Message) {
	log.Printf("Handling /start command for user %d", msg.Chat.ID)
	msgText := StartMessage
	msgWithKeyboard := tgbotapi.NewMessage(msg.Chat.ID, msgText)
	msgWithKeyboard.ReplyMarkup = h.createKeyboard()
	_, err := h.Bot.Send(msgWithKeyboard)
	if err != nil {
		log.Printf("Error sending start message: %v", err)
	}
}

func (h *BotHandler) handleHelp(msg *tgbotapi.Message) {
	log.Printf("Handling /help command for user %d", msg.Chat.ID)
	msgText := HelpMessage
	msgWithKeyboard := tgbotapi.NewMessage(msg.Chat.ID, msgText)
	msgWithKeyboard.ReplyMarkup = h.createKeyboard()
	_, err := h.Bot.Send(msgWithKeyboard)
	if err != nil {
		log.Printf("Error sending help message: %v", err)
	}
}

func (h *BotHandler) promptAddExpenseCategory(msg *tgbotapi.Message) {
	log.Printf("Prompting for expense input for user %d", msg.Chat.ID)
	userState := getUserState(msg.Chat.ID)
	userState.State = StateAddingExpenseCategory
	h.Bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Введите категорию:"))
}

func (h *BotHandler) handleAddExpenseCategoryInput(msg *tgbotapi.Message) {
	log.Printf("Handling expense category input for user %d", msg.Chat.ID)
	userState := getUserState(msg.Chat.ID)
	categoryName := msg.Text

	var category db.Category
	err := h.DB.DB.Collection("categories").FindOne(context.Background(), bson.M{"user_id": msg.Chat.ID, "name": categoryName}).Decode(&category)
	if err != nil {
		// Если категория не найдена, используем "Общая"
		defaultCategory := h.getDefaultCategory()
		userState.CategoryID = defaultCategory.ID
		h.Bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Категория не найдена. Запись добавлена в категорию Общая."))
	} else {
		userState.CategoryID = category.ID
	}

	userState.State = StateAddingExpenseAmount
	h.Bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Введите сумму:"))
}

func (h *BotHandler) getDefaultCategory() db.Category {
	var category db.Category
	err := h.DB.DB.Collection("categories").FindOne(context.Background(), bson.M{"name": "Общая"}).Decode(&category)
	if err != nil {
		log.Printf("Error finding default category: %v", err)
	}
	return category
}

func (h *BotHandler) checkCategoryLimit(userID int64, categoryID primitive.ObjectID, amount float64) bool {
	var category db.Category
	err := h.DB.DB.Collection("categories").FindOne(context.Background(), bson.M{"_id": categoryID}).Decode(&category)
	if err != nil {
		log.Printf("Error finding category: %v", err)
		return false
	}

	// Подсчитываем сумму расходов по этой категории
	filter := bson.M{
		"user_id":     userID,
		"category_id": categoryID,
	}
	cursor, err := h.DB.DB.Collection("expenses").Find(context.Background(), filter)
	if err != nil {
		log.Printf("Error getting expenses: %v", err)
		return false
	}
	defer cursor.Close(context.Background())

	total := 0.0
	for cursor.Next(context.Background()) {
		var expense db.Expense
		if err = cursor.Decode(&expense); err != nil {
			continue
		}
		total += expense.Amount
	}

	// Проверяем, не превышает ли новый расход лимит
	if total+amount > category.Limit {
		return false
	}
	return true
}

func (h *BotHandler) handleAddExpenseAmountInput(msg *tgbotapi.Message) {
	log.Printf("Handling expense amount input for user %d", msg.Chat.ID)
	userState := getUserState(msg.Chat.ID)
	amount, err := strconv.ParseFloat(msg.Text, 64)
	if err != nil {
		h.Bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Неверная сумма. Пожалуйста, введите число, например 100.50"))
		return
	}

	userState.Amount = amount

	// После успешного парсинга суммы, проверяем лимит
	if !h.checkCategoryLimit(msg.Chat.ID, userState.CategoryID, amount) {
		var category db.Category
		err := h.DB.DB.Collection("categories").FindOne(context.Background(), bson.M{"_id": userState.CategoryID}).Decode(&category)
		if err != nil {
			log.Printf("Error finding category: %v", err)
			return
		}
		h.Bot.Send(tgbotapi.NewMessage(msg.Chat.ID, fmt.Sprintf("Превышен лимит категории '%s' (%.2f). Расход %.2f не добавлен.", category.Name, category.Limit, amount)))
		userState.State = StateIdle // Завершаем процесс добавления расхода
		return
	}

	userState.State = StateAddingExpenseNote
	h.Bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Введите заметку:"))
}

func (h *BotHandler) handleAddExpenseNoteInput(msg *tgbotapi.Message) {
	log.Printf("Handling expense note input for user %d", msg.Chat.ID)
	userState := getUserState(msg.Chat.ID)
	note := msg.Text

	expense := db.Expense{
		ID:         primitive.NewObjectID(),
		Amount:     userState.Amount,
		Date:       time.Now().Format("2006-01-02"),
		Note:       note,
		CategoryID: userState.CategoryID,
	}

	filter := bson.M{"user_id": msg.Chat.ID}
	update := bson.M{"$push": bson.M{"expenses": expense}}
	opts := options.Update().SetUpsert(true)
	_, err := h.DB.DB.Collection("users").UpdateOne(context.Background(), filter, update, opts)
	if err != nil {
		h.Bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Ошибка при добавлении расхода. Попробуйте снова."))
		return
	}

	userState.State = StateIdle
	h.Bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Расход добавлен!"))
}

func (h *BotHandler) promptAddCategoryName(msg *tgbotapi.Message) {
	userState := getUserState(msg.Chat.ID)
	userState.State = StateAddingCategoryName
	h.Bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Введите имя новой категории:"))
}

func (h *BotHandler) handleAddCategoryName(msg *tgbotapi.Message) {
	userState := getUserState(msg.Chat.ID)
	userState.State = StateAddingCategoryLimit
	userState.Category = msg.Text
	h.Bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Введите лимит для категории:"))
}

func (h *BotHandler) handleAddCategoryLimit(msg *tgbotapi.Message) {
	userState := getUserState(msg.Chat.ID)
	limit, err := strconv.ParseFloat(msg.Text, 64)
	if err != nil {
		h.Bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Неверный лимит. Пожалуйста, введите число."))
		return
	}

	category := db.Category{
		ID:     primitive.NewObjectID(),
		UserID: msg.Chat.ID,
		Name:   userState.Category,
		Limit:  limit,
	}

	_, err = h.DB.DB.Collection("categories").InsertOne(context.Background(), category)
	if err != nil {
		h.Bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Ошибка при добавлении категории. Попробуйте снова."))
		return
	}

	userState.State = StateIdle
	h.Bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Категория добавлена!"))
}

func (h *BotHandler) promptSetCategoryLimit(msg *tgbotapi.Message) {
	userState := getUserState(msg.Chat.ID)
	userState.State = StateSettingCategoryLimit
	h.Bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Введите имя категории, для которой хотите установить лимит:"))
}

func (h *BotHandler) handleSetCategoryLimit(msg *tgbotapi.Message) {
	userState := getUserState(msg.Chat.ID)
	categoryName := msg.Text

	var category db.Category
	err := h.DB.DB.Collection("categories").FindOne(context.Background(), bson.M{"user_id": msg.Chat.ID, "name": categoryName}).Decode(&category)
	if err != nil {
		h.Bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Категория не найдена. Попробуйте снова."))
		userState.State = StateIdle // Завершаем процесс
		return
	}

	userState.CategoryID = category.ID
	userState.State = StateSettingCategoryLimitAmount
	h.Bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Введите новый лимит для категории:"))
}

func (h *BotHandler) handleSetCategoryLimitAmount(msg *tgbotapi.Message) {
	userState := getUserState(msg.Chat.ID)
	limit, err := strconv.ParseFloat(msg.Text, 64)
	if err != nil {
		h.Bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Неверный лимит. Пожалуйста, введите число."))
		return
	}

	filter := bson.M{"_id": userState.CategoryID}
	update := bson.M{"$set": bson.M{"limit": limit}}
	_, err = h.DB.DB.Collection("categories").UpdateOne(context.Background(), filter, update)
	if err != nil {
		h.Bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Ошибка при установке лимита. Попробуйте снова."))
		return
	}

	userState.State = StateIdle
	h.Bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Лимит для категории установлен!"))
}

func (h *BotHandler) handleGetCategories(msg *tgbotapi.Message) {
	cursor, err := h.DB.DB.Collection("categories").Find(context.Background(), bson.M{"user_id": msg.Chat.ID})
	if err != nil {
		h.Bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Ошибка при получении категорий. Попробуйте снова."))
		return
	}
	defer cursor.Close(context.Background())

	var categories []db.Category
	if err := cursor.All(context.Background(), &categories); err != nil {
		h.Bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Ошибка при обработке категорий. Попробуйте снова."))
		return
	}

	// Если нет ни одной созданной пользователем категории, выводим категорию "Общая"
	if len(categories) == 0 {
		defaultCategory := h.getDefaultCategory()
		categories = append(categories, defaultCategory)
	}

	response := "Ваши категории:\n"
	for _, category := range categories {
		response += fmt.Sprintf("Имя: %s, Лимит: %.2f\n", category.Name, category.Limit)
	}

	h.Bot.Send(tgbotapi.NewMessage(msg.Chat.ID, response))
}

func (h *BotHandler) handleGetExpenses(msg *tgbotapi.Message) {
	log.Printf("Handling /getexpenses command for user %d", msg.Chat.ID)
	filter := bson.M{"user_id": msg.Chat.ID}
	var user db.User
	err := h.DB.DB.Collection("users").FindOne(context.Background(), filter).Decode(&user)
	if err != nil || len(user.Expenses) == 0 {
		log.Printf("No expenses found for user %d", msg.Chat.ID)
		h.Bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Расходы не найдены"))
		return
	}

	response := "Твои расходы:\n"
	for _, expense := range user.Expenses {
		response += fmt.Sprintf("%.2f - %s - %s\n", expense.Amount, expense.Date, expense.Note)
	}

	log.Printf("Sending expenses list to user %d", msg.Chat.ID)
	h.Bot.Send(tgbotapi.NewMessage(msg.Chat.ID, response))
}

func (h *BotHandler) promptDeleteLastExpense(msg *tgbotapi.Message) {
	log.Printf("Prompting for delete last expense for user %d", msg.Chat.ID)
	userState := getUserState(msg.Chat.ID)
	userState.State = StateDeletingLastExpense
	h.Bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Вы уверены, что хотите удалить последний расход? Введите 'Да' для подтверждения."))
}

func (h *BotHandler) handleDeleteLastExpense(msg *tgbotapi.Message) {
	log.Printf("Handling delete last expense for user %d", msg.Chat.ID)
	filter := bson.M{"user_id": msg.Chat.ID}
	updateCmd := bson.M{"$pop": bson.M{"expenses": 1}}
	_, err := h.DB.DB.Collection("users").UpdateOne(context.Background(), filter, updateCmd)
	if err != nil {
		h.Bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Ошибка при удалении последнего расхода. Попробуйте снова."))
		return
	}
	h.Bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Последний расход удален."))

	userState := getUserState(msg.Chat.ID) // Добавляем эту строку для получения состояния пользователя
	userState.State = StateIdle            // Обновляем состояние пользователя на StateIdle
}

func (h *BotHandler) handleUnknownCommand(msg *tgbotapi.Message) {
	log.Printf("Unknown command received from user %d: %s", msg.Chat.ID, msg.Text)
	h.Bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Неизвестная команда. Введите /help для получения списка доступных команд."))
}
