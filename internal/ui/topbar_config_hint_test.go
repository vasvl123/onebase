package ui

import (
	"bytes"
	"strings"
	"testing"
)

func renderHead(t *testing.T, cfg Config, isAdmin bool) string {
	t.Helper()
	var buf bytes.Buffer
	data := map[string]any{
		"Cfg":         cfg,
		"IsAdmin":     isAdmin,
		"Lang":        "ru",
		"SearchQuery": "",
	}
	if err := tmpl.ExecuteTemplate(&buf, "nav", data); err != nil {
		t.Fatalf("render nav: %v", err)
	}
	return buf.String()
}

func TestTopbar_AdminTitleShowsConfigLocation(t *testing.T) {
	html := renderHead(t, Config{
		AppName:          "Торговля",
		AppTitle:         "Система проектирования программных решений",
		ConfigSource:     "file",
		ConfigLocation:   `/srv/configs/<trade>`,
		DatabaseType:     "sqlite",
		DatabaseLocation: `/srv/data/<trade>.db`,
	}, true)

	for _, want := range []string{
		`Конфигурация: file · /srv/configs/&lt;trade&gt;`,
		"SQLite · /srv/data/&lt;trade&gt;.db",
		"Система проектирования программных решений",
		"/srv/configs/&lt;trade&gt;",
	} {
		if !strings.Contains(html, want) {
			t.Errorf("в подсказке шапки нет %q", want)
		}
	}
	if strings.Contains(html, "<trade>") {
		t.Error("пути в подсказке должны быть HTML-экранированы")
	}
}

func TestTopbar_RegularUserSeesFolderNotPath(t *testing.T) {
	html := renderHead(t, Config{
		AppTitle:         "Система проектирования",
		AppName:          "Торговля",
		ConfigSource:     "file",
		ConfigLocation:   `/srv/configs/trade/base`,
		DatabaseType:     "sqlite",
		DatabaseLocation: `/srv/data/trade.db`,
	}, false)

	if !strings.Contains(html, `<span class="topbar-path">trade</span>`) {
		t.Errorf("обычный пользователь должен видеть имя папки базы: %s", html)
	}
	for _, hidden := range []string{"/srv/configs", "/srv/data/trade.db", "Система проектирования"} {
		if strings.Contains(html, hidden) {
			t.Errorf("обычному пользователю раскрыто %q: %s", hidden, html)
		}
	}
	if !strings.Contains(html, `title="Главная"`) {
		t.Error("у обычного пользователя подсказка осталась «Главная»")
	}
}

func TestTopbar_AdminShowsFolderTitleFullPath(t *testing.T) {
	html := renderHead(t, Config{
		AppTitle:         "Система проектирования",
		AppName:          "Торговля",
		ConfigSource:     "file",
		ConfigLocation:   `/srv/configs/trade/base`,
		DatabaseType:     "sqlite",
		DatabaseLocation: `/srv/data/trade.db`,
	}, true)

	order := []string{
		`<span class="topbar-path">trade</span> ⚡`,
		`⚡ Система проектирования <span class="topbar-path">/srv/configs/trade/base</span>`,
	}
	prev := -1
	for _, want := range order {
		idx := strings.Index(html, want)
		if idx < 0 || prev >= idx {
			t.Errorf("шапка админа нарушает порядок (папка, молния, название, полный путь): нет %q", want)
			break
		}
		prev = idx
	}
}
