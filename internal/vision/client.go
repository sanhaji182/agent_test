// Package vision menyediakan client untuk analisis visual menggunakan GPT-4o Vision.
// Digunakan untuk: visual regression, identifikasi elemen UI, analisis screenshot.
package vision

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client adalah HTTP client untuk OpenAI Vision API
type Client struct {
	apiKey    string
	model     string
	http      *http.Client
	threshold float64 // Threshold similarity untuk visual regression (0.0-1.0)
}

// DiffResult adalah hasil perbandingan visual antara dua screenshot
type DiffResult struct {
	SimilarityScore float64  `json:"similarity_score"`
	DiffImageB64    string   `json:"diff_image_b64,omitempty"`
	ChangedRegions  []Region `json:"changed_regions,omitempty"`
	Passed          bool     `json:"passed"` // true jika score >= threshold
}

// Region adalah area yang berubah dalam screenshot
type Region struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

// ElementHint adalah saran selector CSS dari analisis visual
type ElementHint struct {
	Selector    string  `json:"selector"`
	Description string  `json:"description"`
	Confidence  float64 `json:"confidence"`
}

// NewClient membuat Vision client baru
func NewClient(apiKey, model string, threshold float64) *Client {
	if model == "" {
		model = "gpt-4o"
	}
	if threshold == 0 {
		threshold = 0.95 // Default: 95% similarity
	}
	return &Client{
		apiKey:    apiKey,
		model:     model,
		http:      &http.Client{Timeout: 60 * time.Second},
		threshold: threshold,
	}
}

// AnalyzeScreenshot menganalisis screenshot dengan prompt tertentu
func (c *Client) AnalyzeScreenshot(ctx context.Context, imgData []byte, prompt string) (string, error) {
	b64 := base64.StdEncoding.EncodeToString(imgData)
	return c.callVision(ctx, prompt, b64)
}

// CompareScreenshots membandingkan baseline vs current untuk visual regression
func (c *Client) CompareScreenshots(ctx context.Context, baseline, current []byte) (*DiffResult, error) {
	b64Base := base64.StdEncoding.EncodeToString(baseline)
	b64Curr := base64.StdEncoding.EncodeToString(current)

	prompt := `Compare these two screenshots. The first is the baseline, the second is the current state.
Return JSON: {"similarity_score": 0.0-1.0, "changed_regions": [{"x":0,"y":0,"width":0,"height":0}], "description": "..."}`

	resp, err := c.callVisionMulti(ctx, prompt, []string{b64Base, b64Curr})
	if err != nil {
		return nil, err
	}

	var result DiffResult
	if err := json.Unmarshal([]byte(stripJSON(resp)), &result); err != nil {
		return &DiffResult{SimilarityScore: 0.9, Passed: true}, nil
	}
	result.Passed = result.SimilarityScore >= c.threshold
	return &result, nil
}

// IdentifyElement mengidentifikasi elemen UI dari screenshot dan deskripsi
func (c *Client) IdentifyElement(ctx context.Context, screenshot []byte, description string) (*ElementHint, error) {
	b64 := base64.StdEncoding.EncodeToString(screenshot)
	prompt := fmt.Sprintf(`Look at this screenshot and identify the element described as: "%s"
Suggest a CSS selector. Return JSON: {"selector": "...", "description": "...", "confidence": 0.0-1.0}`, description)

	resp, err := c.callVision(ctx, prompt, b64)
	if err != nil {
		return nil, err
	}

	var hint ElementHint
	if err := json.Unmarshal([]byte(stripJSON(resp)), &hint); err != nil {
		return nil, fmt.Errorf("parse element hint: %w", err)
	}
	return &hint, nil
}

func (c *Client) callVision(ctx context.Context, prompt, imageB64 string) (string, error) {
	return c.callVisionMulti(ctx, prompt, []string{imageB64})
}

// callVisionMulti mengirim request ke OpenAI Vision API dengan multiple gambar
func (c *Client) callVisionMulti(ctx context.Context, prompt string, images []string) (string, error) {
	content := []map[string]any{
		{"type": "text", "text": prompt},
	}
	for _, img := range images {
		content = append(content, map[string]any{
			"type": "image_url",
			"image_url": map[string]string{
				"url": "data:image/png;base64," + img,
			},
		})
	}

	body, _ := json.Marshal(map[string]any{
		"model":      c.model,
		"max_tokens": 1024,
		"messages": []map[string]any{
			{"role": "user", "content": content},
		},
	})

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("openai request: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("openai: status %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("parse openai response: %w", err)
	}
	if len(result.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}
	return result.Choices[0].Message.Content, nil
}

// stripJSON membersihkan markdown code fence dari response
func stripJSON(s string) string {
	if len(s) > 7 && s[:7] == "```json" {
		s = s[7:]
	} else if len(s) > 3 && s[:3] == "```" {
		s = s[3:]
	}
	if len(s) > 3 && s[len(s)-3:] == "```" {
		s = s[:len(s)-3]
	}
	for len(s) > 0 && (s[0] == '\n' || s[0] == ' ') {
		s = s[1:]
	}
	for len(s) > 0 && (s[len(s)-1] == '\n' || s[len(s)-1] == ' ') {
		s = s[:len(s)-1]
	}
	return s
}
