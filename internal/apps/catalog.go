package apps

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const DefaultCatalogURL = "https://raw.githubusercontent.com/ModernSelfHost/openappstore/main/index.json"

type App struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Version     string `json:"version"`
	Category    string `json:"category"`
	Description string `json:"description"`
	Runtime     string `json:"runtime"`
	Icon        string `json:"icon"`
	Manifest    string `json:"manifest"`
}

type catalogDoc struct {
	SchemaVersion int   `json:"schemaVersion"`
	Apps          []App `json:"apps"`
}

// FetchCatalog retrieves the app catalog, cached for 1 hour in memory.
type Catalog struct {
	URL      string
	cache    []App
	cachedAt time.Time
}

func NewCatalog(url string) *Catalog {
	if url == "" {
		url = DefaultCatalogURL
	}
	return &Catalog{URL: url}
}

func (c *Catalog) List() ([]App, error) {
	if len(c.cache) > 0 && time.Since(c.cachedAt) < time.Hour {
		return c.cache, nil
	}
	resp, err := http.Get(c.URL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("catalog http %d", resp.StatusCode)
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var doc catalogDoc
	if err := json.Unmarshal(b, &doc); err != nil {
		return nil, err
	}
	c.cache = doc.Apps
	c.cachedAt = time.Now()
	return c.cache, nil
}
