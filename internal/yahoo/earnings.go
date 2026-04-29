package yahoo

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"time"
)

var earningsDateRe = regexp.MustCompile(`earningsDate.{0,15}raw.{0,15}:(\d{10})`)

// GetNextEarningsDate fetches the next upcoming earnings date for a ticker
// by parsing the Yahoo Finance quote page HTML.
// Returns nil, nil if no upcoming date is found.
func (c *Client) GetNextEarningsDate(ctx context.Context, ticker string) (*time.Time, error) {
	url := fmt.Sprintf("https://finance.yahoo.com/quote/%s", ticker)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/87.0.4280.88 Safari/537.36")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	matches := earningsDateRe.FindAllSubmatch(body, -1)
	if len(matches) == 0 {
		return nil, nil
	}

	now := time.Now()
	for _, m := range matches {
		ts, err := strconv.ParseInt(string(m[1]), 10, 64)
		if err != nil {
			continue
		}
		t := time.Unix(ts, 0)
		if t.After(now) {
			return &t, nil
		}
	}

	return nil, nil
}
