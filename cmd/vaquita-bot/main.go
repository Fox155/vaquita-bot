package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"vaquita-bot/src/expense"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/spf13/viper"
)

const (
	NewTicketPrefix  string = "ticket"
	NewExpensePrefix string = "expense"
)

type envConfigs struct {
	TelegramToken string `mapstructure:"TELEGRAM_BOT_TOKEN"`
}

func main() {
	viper.AddConfigPath(".")
	viper.SetConfigName("vaquita-bot")
	viper.SetConfigType("env")

	if err := viper.ReadInConfig(); err != nil {
		log.Panic("Error reading env file", err)
	}

	var config envConfigs
	if err := viper.Unmarshal(&config); err != nil {
		log.Panic(err)
	}

	if config.TelegramToken == "" {
		log.Panic("Empty token")
	}

	bot, err := tgbotapi.NewBotAPI(config.TelegramToken)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil { // ignore any non-Message updates
			continue
		}

		if !update.Message.IsCommand() { // ignore any non-command Messages
			continue
		}

		// Create a new MessageConfig. We don't have text yet,
		// so we leave it empty.
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")

		// Extract the command from the Message.
		switch update.Message.Command() {
		case "help":
			msg.Text = "I understand /clean, /expense, /balance, /total or /summary"
		case "balance":
			msg.Text = Balance(update.FromChat().ID)
		case "total":
			msg.Text = Total(update.FromChat().ID)
		case "summary":
			msg.Text = Summary(update.FromChat().ID)
		case "table":
			msg.Text = Table(update.FromChat().ID)
		case "expense":
			msg.Text = NewExpense(update.FromChat().ID, update.Message.CommandArguments(), update.Message.From.UserName)
		case "clean":
			expense.CleanGroup(update.FromChat().ID)

			msg.Text = "Cleaned"
		default:
			msg.Text = "I don't know that command"
		}

		if msg.Text == "" {
			msg.Text = "Piols"
		}

		if _, err := bot.Send(msg); err != nil {
			log.Panic(err)
		}
	}
}

func NewExpense(chatId int64, commargs string, payerName string) string {
	params := strings.SplitN(commargs, " ", 2)
	if len(params) != 2 {
		return "Try writing the positive amount of the expense followed by its description. For example */expense 1200 Ice*"
	}

	amount, err := strconv.ParseFloat(params[0], 64)
	if err != nil {
		log.Println(err)
		return "Try writing the positive amount of the expense followed by its description. For example */expense 1200 Ice*"
	}

	description := params[1]
	if description == "" || strings.Contains(strings.ToLower(params[0]), "e") || strings.Contains(strings.ToLower(params[0]), "nan") || strings.Contains(strings.ToLower(params[0]), "inf") {
		return "Try writing the positive amount of the expense followed by its description. For example */expense 1200 Ice*"
	}

	if amount < 0 {
		return "Try writing the positive amount of the expense followed by its description. For example */expense 1200 Ice*"
	}

	err = expense.NewExpense(chatId, amount, payerName, description)
	if err != nil {
		return "Piols"
	}

	return payerName + "'s new " + description + " expense added"
}

func Total(chatId int64) string {
	result := expense.GetTotalBalance(chatId)

	return fmt.Sprintf("Total: $%.2f", result)
}

func Balance(chatId int64) string {
	balances, err := expense.GetFullBalance(chatId)
	if err != nil {
		return "Piols"
	}

	output := ""
	for payerName, balance := range balances {
		output += fmt.Sprintf("%s expense %.2f\n", payerName, balance)
	}

	return output
}

func Summary(chatId int64) string {
	result, err := expense.GetDebts(chatId)
	if err != nil {
		return "Piols"
	}

	output := ""
	for _, deb := range result {
		output += fmt.Sprintf("%s owes %s $%.2f\n", deb.Debtor, deb.Creditor, deb.Amount)
	}

	return output
}

func Table(chatId int64) string {
	result := expense.GetExpenses(chatId)

	output := "User\t\tAmount\tDetail\n"
	for _, exp := range result {
		output += fmt.Sprintf("%s\t\t%.2f\t%s\n", exp.PayerName, exp.Amount, exp.Name)
	}

	return output
}
