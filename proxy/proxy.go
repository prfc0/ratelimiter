package proxy

import (
	"io"
	"net"
	"net/http"
	"net/url"
	"ratelimiter/config"
	"ratelimiter/limiter"
)

type Proxy struct {
	backendURL  *url.URL
	rule        config.Rule
	rateLimiter limiter.RateLimiter
}

func NewProxy(backend string, rule config.Rule, rl limiter.RateLimiter) (*Proxy, error) {
	parsed, err := url.Parse(backend)
	if err != nil {
		return nil, err
	}
	return &Proxy{backendURL: parsed, rule: rule, rateLimiter: rl}, nil
}

func (p *Proxy) extractKey(r *http.Request) string {
	switch p.rule.KeySource {
	case "ip":
		host, _, _ := net.SplitHostPort(r.RemoteAddr)
		return host
	case "header":
		return r.Header.Get(p.rule.HeaderName)
	case "path":
		return r.URL.Path
	default:
		return "default"
	}
}

func (p *Proxy) Handler(w http.ResponseWriter, r *http.Request) {
	clientKey := p.extractKey(r)

	// Check rate limit
	allowed, err := p.rateLimiter.Allow(clientKey)
	// fmt.Println(allowed, err, clientKey)
	if err != nil {
		http.Error(w, "Rate limiter error", http.StatusInternalServerError)
		return
	}

	if !allowed {
		http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
		return
	}

	// Forward request
	req, err := http.NewRequest(r.Method, p.backendURL.String()+r.RequestURI, r.Body)
	if err != nil {
		http.Error(w, "Bad Gateway", http.StatusBadGateway)
		return
	}
	req.Header = r.Header.Clone()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		http.Error(w, "Backend unavailable", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Copy response back
	for k, v := range resp.Header {
		for _, vv := range v {
			w.Header().Add(k, vv)
		}
	}
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}
