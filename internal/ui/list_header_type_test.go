package ui

import (
	"strings"
	"testing"

	"github.com/ivantit66/onebase/internal/metadata"
	"github.com/ivantit66/onebase/internal/storage"
)

// TestPageList_HeaderShowsTypePrefix: в заголовке списка показывается тип объекта
// («Справочник: Очередь», «Документ: Назначение»).
func TestPageList_HeaderShowsTypePrefix(t *testing.T) {
	base := map[string]any{
		"Rows":             []map[string]any{{"id": "1", "Наименование": "A"}},
		"Params":           storage.ListParams{},
		"RefFilterOptions": map[string]any{},
		"Lang":             "ru",
		"Feed":             false,
		"Total":            1, "Page": 1, "TotalPages": 1, "HasNext": false,
	}
	base["Entity"] = &metadata.Entity{Name: "Очередь", Kind: metadata.KindCatalog}
	html := renderPageList(t, base)
	if !strings.Contains(html, "<h2>Справочник: Очередь</h2>") {
		t.Errorf("справочник: в заголовке нет типа объекта: %s", html[strings.Index(html, "<h2>"):strings.Index(html, "</h2>")+5])
	}

	base["Entity"] = &metadata.Entity{Name: "Назначение", Kind: metadata.KindDocument}
	html = renderPageList(t, base)
	if !strings.Contains(html, "<h2>Документ: Назначение</h2>") {
		t.Errorf("документ: в заголовке нет типа объекта: %s", html[strings.Index(html, "<h2>"):strings.Index(html, "</h2>")+5])
	}
}
