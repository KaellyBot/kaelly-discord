package constants

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	amqp "github.com/kaellybot/kaelly-amqp"
)

type Language struct {
	Locale          discordgo.Locale
	AmqpLocale      amqp.RabbitMQMessage_Language
	TranslationFile string
}

const (
	i18nFolder = "i18n"

	frenchFile  = "fr.json"
	englishFile = "en.json"

	DefaultLocale = discordgo.EnglishUS
)

var (
	Languages = []Language{
		{
			Locale:          discordgo.French,
			TranslationFile: fmt.Sprintf("%s/%s", i18nFolder, frenchFile),
			AmqpLocale:      amqp.RabbitMQMessage_FR,
		},
		{
			Locale:          discordgo.EnglishGB,
			TranslationFile: fmt.Sprintf("%s/%s", i18nFolder, englishFile),
			AmqpLocale:      amqp.RabbitMQMessage_EN,
		},
		{
			Locale:          discordgo.EnglishUS,
			TranslationFile: fmt.Sprintf("%s/%s", i18nFolder, englishFile),
			AmqpLocale:      amqp.RabbitMQMessage_EN,
		},
	}
)

func MapDiscordLocale(locale discordgo.Locale) amqp.RabbitMQMessage_Language {
	for _, language := range Languages {
		if language.Locale == locale {
			return language.AmqpLocale
		}
	}

	return amqp.RabbitMQMessage_ANY
}

func MapAmqpLocale(locale amqp.RabbitMQMessage_Language) discordgo.Locale {
	for _, language := range Languages {
		if language.AmqpLocale == locale {
			return language.Locale
		}
	}

	return DefaultLocale
}
