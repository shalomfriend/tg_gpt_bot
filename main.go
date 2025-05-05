package main

import (
	"context"
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
	openai "github.com/sashabaranov/go-openai"
)

func main() {
	// Загружаем .env файл с переменными окружения
	err := godotenv.Load()

	log.Println("Отладка: пытаюсь прочитать переменные окружения")

	log.Println("TELEGRAM_BOT_TOKEN =", os.Getenv("TELEGRAM_BOT_TOKEN"))
	log.Println("OPENAI_API_KEY =", os.Getenv("OPENAI_API_KEY"))

	if err != nil {
		log.Fatal("Ошибка при загрузке .env файла")
	}

	// Получаем токены из окружения
	telegramToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	openaiToken := os.Getenv("OPENAI_API_KEY")

	// Проверка наличия токенов
	if telegramToken == "" || openaiToken == "" {
		log.Fatal("Не найдены переменные TELEGRAM_BOT_TOKEN или OPENAI_API_KEY")
	}

	// Логируем часть токенов для отладки (не выводим полностью)
	log.Printf("Токены загружены. Telegram: %s..., OpenAI: %s...", telegramToken[:10], openaiToken[:10])

	// Инициализация Telegram-бота
	bot, err := tgbotapi.NewBotAPI(telegramToken)
	if err != nil {
		log.Panic("Ошибка инициализации Telegram бота:", err)
	}

	log.Printf("Бот %s запущен", bot.Self.UserName)

	// Настраиваем получение обновлений
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	// Основной цикл обработки сообщений
	for update := range updates {
		if update.Message == nil {
			continue
		}

		userInput := update.Message.Text
		userName := update.Message.From.UserName

		log.Printf("Запрос от @%s: %s", userName, userInput)

		// Получаем ответ от ChatGPT
		replyText, err := getChatGPTResponse(userInput, openaiToken)
		if err != nil {
			log.Println("Ошибка OpenAI:", err)
			replyText = "Произошла ошибка при обращении к ChatGPT."
		}

		// Отправляем ответ пользователю
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, replyText)
		if _, err := bot.Send(msg); err != nil {
			log.Println("Ошибка при отправке сообщения в Telegram:", err)
		}
	}
}

// getChatGPTResponse отправляет запрос к OpenAI и возвращает ответ
func getChatGPTResponse(prompt, apiKey string) (string, error) {
	client := openai.NewClient(apiKey)

	req := openai.ChatCompletionRequest{
		Model: openai.GPT3Dot5Turbo, // Более доступная и быстрая модель
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleUser, Content: prompt},
		},
	}

	resp, err := client.CreateChatCompletion(context.Background(), req)
	if err != nil {
		return "", err
	}

	// Защита от пустого ответа
	if len(resp.Choices) == 0 {
		return "OpenAI не вернул ответ", nil
	}

	return resp.Choices[0].Message.Content, nil
}
