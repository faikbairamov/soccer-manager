package i18n

import (
	"encoding/json"
	"log/slog"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

type contextKey string

const localeKey contextKey = "locale"

var bundle *i18n.Bundle

func Init(localesDir string) {
	bundle = i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("json", json.Unmarshal)

	for _, lang := range []string{"en", "ka"} {
		path := localesDir + "/" + lang + ".json"
		if _, err := bundle.LoadMessageFile(path); err != nil {
			slog.Error("failed to load locale path", "path", path, "err", err)
			os.Exit(1)
		}
	}
}

func Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		lang := c.GetHeader("Accept-Language")
		c.Set(string(localeKey), lang)
		c.Next()
	}
}

func Translate(c *gin.Context, messageID string) string {
	lang, _ := c.Get(string(localeKey))
	langStr, _ := lang.(string)
	localizer := i18n.NewLocalizer(bundle, langStr, language.English.String())
	msg, err := localizer.Localize(&i18n.LocalizeConfig{MessageID: messageID})
	if err != nil {
		return messageID
	}
	return msg
}
