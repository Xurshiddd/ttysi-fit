package hemis

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/ttysi-fit/backend/config"
	"golang.org/x/time/rate"
)

// Client — HEMIS REST API klienti.
// HEMIS sekundiga 10 ta so'rovdan oshganda bloklaydi — har bir so'rov
// rate limiter orqali o'tadi.
type Client struct {
	cfg     config.HEMISConfig
	http    *http.Client
	limiter *rate.Limiter
}

// NewClient — config asosida HEMIS klientini yaratadi.
func NewClient(cfg config.HEMISConfig) *Client {
	if cfg.Timeout <= 0 {
		cfg.Timeout = 30 * time.Second
	}
	if cfg.PageLimit <= 0 {
		cfg.PageLimit = 200
	}
	rps := cfg.RateLimit
	if rps <= 0 {
		rps = 10
	}
	return &Client{
		cfg:  cfg,
		http: &http.Client{Timeout: cfg.Timeout},
		// Sekundiga rps ta so'rov, teng oraliqda (burst=1) — HEMIS blokini oldini oladi.
		limiter: rate.NewLimiter(rate.Limit(rps), 1),
	}
}

// ── HEMIS javob qobig'i ────────────────────────────────

type apiResponse struct {
	Success bool            `json:"success"`
	Error   json.RawMessage `json:"error"`
	Data    struct {
		Items      []json.RawMessage `json:"items"`
		Pagination *struct {
			PageCount int `json:"pageCount"`
			Page      int `json:"page"`
		} `json:"pagination"`
	} `json:"data"`
}

// fetchAll — berilgan endpointdan barcha elementlarni (paginatsiya bo'ylab) yig'adi.
func (c *Client) fetchAll(ctx context.Context, path string, extra url.Values) ([]json.RawMessage, error) {
	if c.cfg.BaseURL == "" || c.cfg.Token == "" {
		return nil, fmt.Errorf("hemis: BaseURL yoki Token sozlanmagan")
	}

	var all []json.RawMessage
	page := 1
	for {
		resp, err := c.fetchPage(ctx, path, page, extra)
		if err != nil {
			return nil, err
		}
		all = append(all, resp.Data.Items...)

		if resp.Data.Pagination == nil ||
			page >= resp.Data.Pagination.PageCount ||
			len(resp.Data.Items) == 0 {
			break
		}
		page++
	}
	return all, nil
}

func (c *Client) fetchPage(ctx context.Context, path string, page int, extra url.Values) (*apiResponse, error) {
	u, err := url.Parse(c.cfg.BaseURL + path)
	if err != nil {
		return nil, fmt.Errorf("hemis: url xato: %w", err)
	}
	q := u.Query()
	q.Set("limit", strconv.Itoa(c.cfg.PageLimit))
	q.Set("page", strconv.Itoa(page))
	for k, vs := range extra {
		for _, v := range vs {
			q.Set(k, v)
		}
	}
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.cfg.Token)
	req.Header.Set("Accept", "application/json")

	// Rate limit: HEMIS 10 req/sek dan oshganda bloklaydi.
	if err := c.limiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("hemis: rate limit kutish bekor qilindi: %w", err)
	}

	res, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("hemis: so'rov xatosi: %w", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(io.LimitReader(res.Body, 50<<20)) // 50MB limit
	if err != nil {
		return nil, fmt.Errorf("hemis: javob o'qilmadi: %w", err)
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("hemis: status %d: %s", res.StatusCode, string(body))
	}

	var parsed apiResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, fmt.Errorf("hemis: javob tahlil xatosi: %w", err)
	}
	if !parsed.Success {
		return nil, fmt.Errorf("hemis: success=false: %s", string(parsed.Error))
	}
	return &parsed, nil
}

// ── Umumiy yordamchi tiplar ────────────────────────────

type typeRef struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

// departmentRef — HEMIS yozuvlaridagi `department` obyekti (fakultet yoki kafedra).
type departmentRef struct {
	ID            int64    `json:"id"`
	Name          string   `json:"name"`
	Code          string   `json:"code"`
	StructureType *typeRef `json:"structureType"`
}
