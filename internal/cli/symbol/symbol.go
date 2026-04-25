package symbol

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	c "github.com/achannarasappa/ticker/v5/internal/common"
)

const defaultUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/134.0.0.0 Safari/537.36"

type SymbolSourceMap struct { //nolint:golint,revive
	TickerSymbol string
	SourceSymbol string
	Source       c.QuoteSource
}

type TickerSymbolToSourceSymbol map[string]SymbolSourceMap

func parseQuoteSource(id string) c.QuoteSource {

	if id == "cb" {
		return c.QuoteSourceCoinbase
	}

	return c.QuoteSourceUnknown
}

func parseTickerSymbolToSourceSymbol(body io.ReadCloser) (TickerSymbolToSourceSymbol, error) {

	out := TickerSymbolToSourceSymbol{}
	reader := csv.NewReader(body)
	reader.LazyQuotes = true
	for {

		row, err := reader.Read()

		if errors.Is(err, io.EOF) {
			body.Close()

			break
		}

		if err != nil {
			return nil, err
		}

		if _, exists := out[row[0]]; !exists {
			out[row[0]] = SymbolSourceMap{
				TickerSymbol: row[0],
				SourceSymbol: row[1],
				Source:       parseQuoteSource(row[2]),
			}

		}
	}

	return out, nil
}

// GetTickerSymbols retrieves a list of ticker specific symbols and their data source
func GetTickerSymbols(symbolUrl string) (TickerSymbolToSourceSymbol, error) {
	parsedURL, err := url.Parse(symbolUrl)
	if err != nil {
		return TickerSymbolToSourceSymbol{}, fmt.Errorf("invalid URL: %w", err)
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return TickerSymbolToSourceSymbol{}, errors.New("invalid URL scheme: must be http or https")
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest(http.MethodGet, parsedURL.String(), nil)
	if err != nil {
		return TickerSymbolToSourceSymbol{}, err
	}

	req.Header.Set("User-Agent", defaultUserAgent)

	resp, err := client.Do(req)
	if err != nil {
		return TickerSymbolToSourceSymbol{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return TickerSymbolToSourceSymbol{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	tickerSymbolToSourceSymbol, err := parseTickerSymbolToSourceSymbol(resp.Body)
	if err != nil {
		return TickerSymbolToSourceSymbol{}, err
	}

	return tickerSymbolToSourceSymbol, nil
}
