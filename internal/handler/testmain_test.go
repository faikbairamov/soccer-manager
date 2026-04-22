package handler

import (
	"os"
	"testing"

	"github.com/faikbairamov/soccer-manager/internal/i18n"
)

func TestMain(m *testing.M) {
	i18n.Init("../../locales")
	os.Exit(m.Run())
}
