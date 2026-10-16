package app

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Webhooks: every session outcome and delegation is POSTed to the app's endpoint with a
// Stripe-style signature header: Singleton-Signature: t=<unix>,v1=<hex HMAC-SHA256>.

var webhookClient = &http.Client{Timeout: 5 * time.Second}

func SignWebhook(secret string, t int64, body []byte) string {
	m := hmac.New(sha256.New, []byte(secret))
	fmt.Fprintf(m, "%d.", t)
	m.Write(body)
	return fmt.Sprintf("t=%d,v1=%s", t, hex.EncodeToString(m.Sum(nil)))
}

func (s *Service) emit(app *App, event string, data map[string]any) {
	if app.WebhookURL == "" {
		return
	}
	body, _ := json.Marshal(map[string]any{"id": newID("evt"), "type": event, "created": time.Now().Unix(), "app": app.ID, "data": data})
	id := newID("whd")
	_, _ = s.DB.Exec(`INSERT INTO webhook_deliveries(id, app_id, event, payload, status, created_at) VALUES(?,?,?,?,?,?)`, id, app.ID, event, string(body), "pending", Now())
	go s.deliver(id, app.WebhookURL, app.WebhookSecret, body)
}

func (s *Service) deliver(id, url, secret string, body []byte) {
	backoff := time.Second
	for attempt := 1; attempt <= 3; attempt++ {
		req, _ := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("User-Agent", "Singleton-Webhooks/1.0")
		req.Header.Set("Singleton-Signature", SignWebhook(secret, time.Now().Unix(), body))
		resp, err := webhookClient.Do(req)
		code := 0
		if err == nil {
			code = resp.StatusCode
			resp.Body.Close()
		}
		status := "failed"
		if err == nil && code >= 200 && code < 300 {
			status = "delivered"
		}
		_, _ = s.DB.Exec(`UPDATE webhook_deliveries SET status=?, response_code=?, attempts=? WHERE id=?`, status, code, attempt, id)
		if status == "delivered" {
			return
		}
		time.Sleep(backoff)
		backoff *= 3
	}
}

type WebhookDelivery struct {
	ID        string `json:"id"`
	AppID     string `json:"appId"`
	Event     string `json:"event"`
	Payload   string `json:"payload"`
	Status    string `json:"status"`
	Code      int    `json:"code"`
	Attempts  int    `json:"attempts"`
	CreatedAt string `json:"createdAt"`
}

func (s *Service) Deliveries(appIDs []string, limit int) ([]WebhookDelivery, error) {
	if len(appIDs) == 0 {
		return nil, nil
	}
	args := []any{}
	for _, id := range appIDs {
		args = append(args, id)
	}
	args = append(args, limit)
	rows, err := s.DB.Query(`SELECT id, app_id, event, payload, status, COALESCE(response_code,0), attempts, created_at FROM webhook_deliveries WHERE app_id IN (`+placeholders(len(appIDs))+`) ORDER BY created_at DESC LIMIT ?`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []WebhookDelivery
	for rows.Next() {
		var d WebhookDelivery
		if err := rows.Scan(&d.ID, &d.AppID, &d.Event, &d.Payload, &d.Status, &d.Code, &d.Attempts, &d.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}
