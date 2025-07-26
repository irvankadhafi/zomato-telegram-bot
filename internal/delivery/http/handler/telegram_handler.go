package handler

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/gofiber/fiber/v2"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"telegram-bot/internal/domain"
	"telegram-bot/internal/usecase"
	"telegram-bot/pkg/response"
)

// TelegramHandler handles Telegram webhook endpoints
type TelegramHandler struct {
	bot               *tgbotapi.BotAPI
	restaurantUsecase *usecase.RestaurantUsecase
}

// NewTelegramHandler creates a new Telegram handler
func NewTelegramHandler(bot *tgbotapi.BotAPI, restaurantUsecase *usecase.RestaurantUsecase) *TelegramHandler {
	return &TelegramHandler{
		bot:               bot,
		restaurantUsecase: restaurantUsecase,
	}
}

// HandleWebhook handles incoming Telegram webhook updates
func (h *TelegramHandler) HandleWebhook(c *fiber.Ctx) error {
	var update tgbotapi.Update
	if err := c.BodyParser(&update); err != nil {
		return response.BadRequest(c, "Invalid webhook payload", err.Error())
	}

	// Process the update
	if err := h.processUpdate(update); err != nil {
		log.Printf("Error processing Telegram update: %v", err)
		// Don't return error to Telegram to avoid webhook retries
	}

	return response.Success(c, "Webhook processed successfully", nil)
}

// processUpdate processes a Telegram update
func (h *TelegramHandler) processUpdate(update tgbotapi.Update) error {
	// Handle text messages
	if update.Message != nil {
		return h.handleMessage(update.Message)
	}

	// Handle callback queries (inline keyboard buttons)
	if update.CallbackQuery != nil {
		return h.handleCallbackQuery(update.CallbackQuery)
	}

	return nil
}

// handleMessage handles text messages
func (h *TelegramHandler) handleMessage(message *tgbotapi.Message) error {
	if message.Text == "" {
		return nil
	}

	chatID := message.Chat.ID
	text := strings.TrimSpace(message.Text)

	// Handle commands
	if strings.HasPrefix(text, "/") {
		return h.handleCommand(chatID, text, message.From.UserName)
	}

	// Handle regular text as search query
	return h.handleSearchQuery(chatID, text)
}

// handleCommand handles bot commands
func (h *TelegramHandler) handleCommand(chatID int64, command string, username string) error {
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return nil
	}

	cmd := parts[0]

	switch cmd {
	case "/start":
		return h.sendWelcomeMessage(chatID, username)
	case "/help":
		return h.sendHelpMessage(chatID)
	case "/search":
		if len(parts) < 2 {
			return h.sendMessage(chatID, "Please provide a search query. Example: /search pizza")
		}
		query := strings.Join(parts[1:], " ")
		return h.handleSearchQuery(chatID, query)
	case "/nearby":
		return h.sendNearbyInstructions(chatID)
	default:
		return h.sendMessage(chatID, "Unknown command. Type /help to see available commands.")
	}
}

// handleSearchQuery handles restaurant search queries
func (h *TelegramHandler) handleSearchQuery(chatID int64, query string) error {
	// Send "typing" action
	typingAction := tgbotapi.NewChatAction(chatID, tgbotapi.ChatTyping)
	if _, err := h.bot.Send(typingAction); err != nil {
		log.Printf("Failed to send typing action: %v", err)
	}

	// Search for restaurants
	req := &domain.RestaurantSearchRequest{
		Query: query,
	}

	result, err := h.restaurantUsecase.SearchByName(context.Background(), req)
	if err != nil {
		return h.sendMessage(chatID, "Sorry, I couldn't search for restaurants right now. Please try again later.")
	}

	if len(result.Restaurants) == 0 {
		return h.sendMessage(chatID, fmt.Sprintf("No restaurants found for '%s'. Try a different search term.", query))
	}

	// Send results
	return h.sendRestaurantResults(chatID, result.Restaurants, query)
}

// sendRestaurantResults sends restaurant search results
func (h *TelegramHandler) sendRestaurantResults(chatID int64, restaurants []domain.Restaurant, query string) error {
	// Send header message
	headerMsg := fmt.Sprintf("🔍 Found %d restaurants for '%s':\n\n", len(restaurants), query)
	if err := h.sendMessage(chatID, headerMsg); err != nil {
		return err
	}

	// Send each restaurant (limit to first 5 to avoid spam)
	maxResults := 5
	if len(restaurants) < maxResults {
		maxResults = len(restaurants)
	}

	for i := 0; i < maxResults; i++ {
		restaurant := restaurants[i]
		formattedMsg := h.restaurantUsecase.FormatRestaurantForTelegram(restaurant)

		// Create inline keyboard for restaurant actions
		keyboard := h.createRestaurantKeyboard(restaurant)

		msg := tgbotapi.NewMessage(chatID, formattedMsg)
		msg.ParseMode = "Markdown"
		msg.ReplyMarkup = keyboard

		if _, err := h.bot.Send(msg); err != nil {
			log.Printf("Failed to send restaurant message: %v", err)
			continue
		}
	}

	if len(restaurants) > maxResults {
		footerMsg := fmt.Sprintf("\n... and %d more results. Try a more specific search for better results.", len(restaurants)-maxResults)
		return h.sendMessage(chatID, footerMsg)
	}

	return nil
}

// createRestaurantKeyboard creates inline keyboard for restaurant actions
func (h *TelegramHandler) createRestaurantKeyboard(restaurant domain.Restaurant) tgbotapi.InlineKeyboardMarkup {
	var buttons [][]tgbotapi.InlineKeyboardButton

	// Add "Get Directions" button if we have location
	if restaurant.Geometry != nil && restaurant.Geometry.Location != nil {
		mapsURL := fmt.Sprintf("https://www.google.com/maps/search/?api=1&query=%f,%f",
			restaurant.Geometry.Location.Lat, restaurant.Geometry.Location.Lng)
		directionsBtn := tgbotapi.NewInlineKeyboardButtonURL("📍 Get Directions", mapsURL)
		buttons = append(buttons, []tgbotapi.InlineKeyboardButton{directionsBtn})
	}

	// Add "More Info" button
	moreInfoBtn := tgbotapi.NewInlineKeyboardButtonData("ℹ️ More Info", fmt.Sprintf("info_%s", restaurant.PlaceID))
	shareBtn := tgbotapi.NewInlineKeyboardButtonData("📤 Share", fmt.Sprintf("share_%s", restaurant.PlaceID))
	buttons = append(buttons, []tgbotapi.InlineKeyboardButton{moreInfoBtn, shareBtn})

	return tgbotapi.NewInlineKeyboardMarkup(buttons...)
}

// handleCallbackQuery handles inline keyboard button presses
func (h *TelegramHandler) handleCallbackQuery(callbackQuery *tgbotapi.CallbackQuery) error {
	// Answer the callback query to remove loading state
	callback := tgbotapi.NewCallback(callbackQuery.ID, "")
	if _, err := h.bot.Request(callback); err != nil {
		log.Printf("Failed to answer callback query: %v", err)
	}

	data := callbackQuery.Data
	chatID := callbackQuery.Message.Chat.ID

	if strings.HasPrefix(data, "info_") {
		return h.sendMessage(chatID, "ℹ️ For more detailed information, please visit the restaurant's Google Maps page using the directions button.")
	}

	if strings.HasPrefix(data, "share_") {
		return h.sendMessage(chatID, "📤 To share this restaurant, forward this message to your friends!")
	}

	return nil
}

// sendWelcomeMessage sends welcome message
func (h *TelegramHandler) sendWelcomeMessage(chatID int64, username string) error {
	name := "there"
	if username != "" {
		name = username
	}

	message := fmt.Sprintf(`👋 Hello %s! Welcome to the Restaurant Finder Bot!

🍽️ I can help you find great restaurants. Here's what you can do:

🔍 **Search for restaurants:**
• Just type what you're looking for (e.g., "pizza", "sushi", "italian food")
• Use /search followed by your query

📍 **Find nearby restaurants:**
• Use /nearby for instructions on location-based search

❓ **Need help?**
• Type /help to see all available commands

Try searching for something delicious! 🍕🍜🍔`, name)

	return h.sendMessage(chatID, message)
}

// sendHelpMessage sends help message
func (h *TelegramHandler) sendHelpMessage(chatID int64) error {
	message := `🤖 **Restaurant Finder Bot Help**

**Available Commands:**

/start - Show welcome message
/help - Show this help message
/search <query> - Search for restaurants
/nearby - Get instructions for nearby search

**How to search:**
• Type any food or restaurant name
• Examples: "pizza", "chinese food", "McDonald's"
• Be specific for better results

**Features:**
• 🔍 Search restaurants by name or cuisine
• ⭐ See ratings and reviews
• 📍 Get directions to restaurants
• 💰 View price levels
• 🕒 Check opening hours

**Tips:**
• Try different search terms if you don't find what you're looking for
• Use specific cuisine types like "italian", "thai", "mexican"
• Include location in your search for better results

Happy dining! 🍽️`

	return h.sendMessage(chatID, message)
}

// sendNearbyInstructions sends nearby search instructions
func (h *TelegramHandler) sendNearbyInstructions(chatID int64) error {
	message := `📍 **Find Nearby Restaurants**

To find restaurants near you:

1. Share your location with me by:
   • Tap the attachment button (📎)
   • Select "Location"
   • Choose "Send My Current Location"

2. Or search with a specific location:
   • Example: "pizza in New York"
   • Example: "sushi near Times Square"

🔍 You can also use our web API for coordinate-based search at:
/restaurants/nearby?lat=YOUR_LAT&lng=YOUR_LNG

Try it now! 📱`

	return h.sendMessage(chatID, message)
}

// sendMessage sends a text message
func (h *TelegramHandler) sendMessage(chatID int64, text string) error {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"
	_, err := h.bot.Send(msg)
	return err
}