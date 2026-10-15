package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"portal/internal/models"
)

const maxWeChatResponseBytes = 32 << 20

var weChatAPIBaseURL = "https://api.weixin.qq.com"

var weChatHTTPClient = &http.Client{Timeout: 30 * time.Second}

type weChatTokenEntry struct {
	AppID     string
	Token     string
	ExpiresAt time.Time
}

var weChatTokenCache = struct {
	sync.Mutex
	items map[uint]weChatTokenEntry
}{items: make(map[uint]weChatTokenEntry)}

type WeChatAPIError struct {
	Code    int
	Message string
}

func (e *WeChatAPIError) Error() string {
	return fmt.Sprintf("微信接口返回错误 %d: %s", e.Code, e.Message)
}

type ExternalContentArticle struct {
	ExternalID        string
	ItemIndex         int
	Title             string
	Author            string
	Summary           string
	Content           string
	Cover             string
	URL               string
	ExternalUpdatedAt time.Time
	Deleted           bool
}

type weChatNewsItem struct {
	Title            string     `json:"title"`
	Author           string     `json:"author"`
	Digest           string     `json:"digest"`
	Content          string     `json:"content"`
	ContentSourceURL string     `json:"content_source_url"`
	ThumbURL         string     `json:"thumb_url"`
	URL              string     `json:"url"`
	IsDeleted        weChatBool `json:"is_deleted"`
}

// weChatBool 兼容微信接口历史上同时出现的 true/false 与 0/1 两种布尔编码。
type weChatBool bool

func (b *weChatBool) UnmarshalJSON(data []byte) error {
	raw := strings.TrimSpace(string(data))
	switch raw {
	case "true", "1", `"1"`:
		*b = true
		return nil
	case "false", "0", "null", `"0"`, "":
		*b = false
		return nil
	default:
		return fmt.Errorf("无效的微信布尔值 %s", raw)
	}
}

type weChatContentRecord struct {
	ArticleID  string `json:"article_id"`
	MediaID    string `json:"media_id"`
	UpdateTime int64  `json:"update_time"`
	Content    struct {
		NewsItems []weChatNewsItem `json:"news_item"`
	} `json:"content"`
}

type weChatBatchResponse struct {
	ErrCode    int                   `json:"errcode"`
	ErrMsg     string                `json:"errmsg"`
	TotalCount int                   `json:"total_count"`
	ItemCount  int                   `json:"item_count"`
	Items      []weChatContentRecord `json:"item"`
}

type weChatTokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	ErrCode     int    `json:"errcode"`
	ErrMsg      string `json:"errmsg"`
}

type WeChatContentProvider struct{}

func (p *WeChatContentProvider) InvalidateToken(sourceID uint) {
	weChatTokenCache.Lock()
	delete(weChatTokenCache.items, sourceID)
	weChatTokenCache.Unlock()
}

func (p *WeChatContentProvider) getAccessToken(ctx context.Context, source *models.ContentSource) (string, error) {
	weChatTokenCache.Lock()
	entry, ok := weChatTokenCache.items[source.ID]
	weChatTokenCache.Unlock()
	if ok && entry.AppID == source.AppID && entry.Token != "" && time.Now().Before(entry.ExpiresAt) {
		return entry.Token, nil
	}

	secret := strings.TrimSpace(source.AppSecret)
	if secret == "" {
		return "", fmt.Errorf("未配置公众号 AppSecret，请在公众号同步页面填写")
	}
	values := url.Values{}
	values.Set("grant_type", "client_credential")
	values.Set("appid", source.AppID)
	values.Set("secret", secret)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, weChatAPIBaseURL+"/cgi-bin/token?"+values.Encode(), nil)
	if err != nil {
		return "", err
	}
	resp, err := weChatHTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("连接微信接口失败: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("微信接口返回 HTTP %d", resp.StatusCode)
	}
	var result weChatTokenResponse
	if err := decodeLimitedJSON(resp.Body, &result); err != nil {
		return "", fmt.Errorf("解析微信访问凭证失败: %w", err)
	}
	if result.ErrCode != 0 {
		return "", &WeChatAPIError{Code: result.ErrCode, Message: result.ErrMsg}
	}
	if result.AccessToken == "" {
		return "", fmt.Errorf("微信接口未返回访问凭证")
	}
	expires := result.ExpiresIn
	if expires <= 0 {
		expires = 7200
	}
	// 提前 5 分钟失效，避免同步过程中使用临界过期的 token。
	validFor := time.Duration(expires-300) * time.Second
	if validFor < time.Minute {
		validFor = time.Minute
	}
	weChatTokenCache.Lock()
	weChatTokenCache.items[source.ID] = weChatTokenEntry{
		AppID: source.AppID, Token: result.AccessToken, ExpiresAt: time.Now().Add(validFor),
	}
	weChatTokenCache.Unlock()
	return result.AccessToken, nil
}

func (p *WeChatContentProvider) TestConnection(ctx context.Context, source *models.ContentSource) error {
	p.InvalidateToken(source.ID)
	_, err := p.getAccessToken(ctx, source)
	return err
}

// FetchArticles 分页读取公众号文章并逐篇回调，避免全量文章正文同时堆在内存里。
func (p *WeChatContentProvider) FetchArticles(
	ctx context.Context,
	source *models.ContentSource,
	yield func(ExternalContentArticle) error,
) (int, error) {
	token, err := p.getAccessToken(ctx, source)
	if err != nil {
		return 0, err
	}

	offset := 0
	scanned := 0
	for page := 0; page < 500; page++ {
		result, err := p.fetchPage(ctx, source.SyncMode, token, offset)
		if apiErr, ok := err.(*WeChatAPIError); ok && (apiErr.Code == 40014 || apiErr.Code == 42001) {
			p.InvalidateToken(source.ID)
			token, err = p.getAccessToken(ctx, source)
			if err == nil {
				result, err = p.fetchPage(ctx, source.SyncMode, token, offset)
			}
		}
		if err != nil {
			return scanned, err
		}
		if len(result.Items) == 0 {
			return scanned, nil
		}
		for _, item := range result.Items {
			externalID := item.ArticleID
			if externalID == "" {
				externalID = item.MediaID
			}
			if externalID == "" {
				continue
			}
			updatedAt := time.Unix(item.UpdateTime, 0)
			for index, article := range item.Content.NewsItems {
				scanned++
				articleURL := article.URL
				if articleURL == "" {
					articleURL = article.ContentSourceURL
				}
				if err := yield(ExternalContentArticle{
					ExternalID:        externalID,
					ItemIndex:         index,
					Title:             article.Title,
					Author:            article.Author,
					Summary:           article.Digest,
					Content:           article.Content,
					Cover:             article.ThumbURL,
					URL:               articleURL,
					ExternalUpdatedAt: updatedAt,
					Deleted:           bool(article.IsDeleted),
				}); err != nil {
					return scanned, err
				}
			}
		}
		offset += len(result.Items)
		if offset >= result.TotalCount || result.ItemCount == 0 {
			return scanned, nil
		}
	}
	return scanned, fmt.Errorf("公众号文章超过单次同步上限，请联系管理员分批处理")
}

func (p *WeChatContentProvider) fetchPage(ctx context.Context, syncMode, token string, offset int) (*weChatBatchResponse, error) {
	endpoint := "/cgi-bin/freepublish/batchget"
	payload := map[string]any{"offset": offset, "count": 20, "no_return_content": 0}
	if syncMode == models.ContentSourceSyncMaterials {
		endpoint = "/cgi-bin/material/batchget_material"
		payload = map[string]any{"type": "news", "offset": offset, "count": 20}
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	reqURL := weChatAPIBaseURL + endpoint + "?access_token=" + url.QueryEscape(token)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := weChatHTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("连接微信接口失败: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("微信接口返回 HTTP %d", resp.StatusCode)
	}
	var result weChatBatchResponse
	if err := decodeLimitedJSON(resp.Body, &result); err != nil {
		return nil, fmt.Errorf("解析微信文章列表失败: %w", err)
	}
	if result.ErrCode != 0 {
		return nil, &WeChatAPIError{Code: result.ErrCode, Message: result.ErrMsg}
	}
	return &result, nil
}

func decodeLimitedJSON(reader io.Reader, target any) error {
	return json.NewDecoder(io.LimitReader(reader, maxWeChatResponseBytes)).Decode(target)
}
