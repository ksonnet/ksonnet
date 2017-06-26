package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/golang/glog"
)

var (
	commaRegexp = regexp.MustCompile(", *")
)

// Registry is a *crazy limited* Docker registry client.
type Registry struct {
	URL    string
	Client *http.Client
}

// NewRegistryClient creates a new Registry client using the given
// http client and base URL.
func NewRegistryClient(client *http.Client, url string) *Registry {
	return &Registry{
		URL:    strings.TrimSuffix(url, "/"),
		Client: client,
	}
}

// ManifestDigest fetches the manifest digest for a given reponame and tag.
func (r *Registry) ManifestDigest(reponame, tag string) (string, error) {
	url := fmt.Sprintf("%s/v2/%s/manifests/%s", r.URL, reponame, tag)

	glog.V(1).Infof("HEAD %s", url)

	req, err := http.NewRequest(http.MethodHead, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Add("Accept", "application/vnd.docker.distribution.manifest.v2+json")
	resp, err := r.Client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Request failed: %s", resp.Status)
	}

	digest := resp.Header.Get("Docker-Content-Digest")
	if digest == "" {
		return "", errors.New("No digest in response")
	}

	glog.V(1).Infof("Found digest %s", digest)
	return digest, nil
}

// stolen from golang 1.8
func stripPort(hostport string) string {
	colon := strings.IndexByte(hostport, ':')
	if colon == -1 {
		return hostport
	}
	if i := strings.IndexByte(hostport, ']'); i != -1 {
		return strings.TrimPrefix(hostport[:i], "[")
	}
	return hostport[:colon]
}

func matchesDomain(url *url.URL, domain string) bool {
	host := stripPort(url.Host)
	return strings.HasSuffix(host, domain)
}

type authChallenge struct {
	Scheme string
	Params map[string]string
}

func parseAuthHeader(header http.Header) []*authChallenge {
	authHeaders := header[http.CanonicalHeaderKey("WWW-Authenticate")]
	ret := make([]*authChallenge, 0, len(authHeaders))
	for _, h := range authHeaders {
		var scheme string
		params := map[string]string{}
		parts := strings.SplitN(h, " ", 2)
		if len(parts) < 1 || parts[0] == "" {
			continue
		}
		scheme = strings.ToLower(parts[0])
		if len(parts) == 2 {
			args := commaRegexp.Split(parts[1], -1)
			for _, arg := range args {
				if parts := strings.SplitN(arg, "=", 2); len(parts) == 2 {

					params[parts[0]] = strings.Trim(parts[1], `"`)
				} else if len(parts) == 1 {
					params[parts[0]] = ""
				}
			}

		}
		auth := authChallenge{
			Scheme: scheme,
			Params: params,
		}
		ret = append(ret, &auth)
	}
	return ret
}

// NewAuthTransport returns a roundtripper that does bearer/etc authentication
func NewAuthTransport(inner http.RoundTripper) http.RoundTripper {
	return &authTransport{
		Transport:  inner,
		Client:     &http.Client{Transport: inner},
		tokenCache: map[string]string{},
	}
}

type authTransport struct {
	Client     *http.Client
	Transport  http.RoundTripper
	tokenCache map[string]string
	HostDomain string
	Username   string
	Password   string
}

// RoundTrip is required for the http.RoundTripper interface
func (t *authTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	glog.V(1).Infof("=> %v", req)
	resp, err := t.Transport.RoundTrip(req)
	glog.V(1).Infof("<= err=%v resp=%v", err, resp)
	if err == nil && resp.StatusCode == http.StatusUnauthorized && matchesDomain(req.URL, t.HostDomain) {
		schemes := parseAuthHeader(resp.Header)
		for _, scheme := range schemes {
			if scheme.Scheme == "basic" {
				glog.V(2).Infof("Retrying with basic auth")
				req.SetBasicAuth(t.Username, t.Password)
				glog.V(1).Infof("=> %v", req)
				return t.Transport.RoundTrip(req)
			}
			if scheme.Scheme == "bearer" {
				token, err := t.bearerAuth(scheme.Params["realm"], scheme.Params["service"], scheme.Params["scope"])
				if err != nil {
					return resp, err
				}
				glog.V(2).Infof("Retrying with bearer auth")
				req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
				glog.V(1).Infof("=> %v", req)
				return t.Transport.RoundTrip(req)
			}
		}
		// No recognised auth schemes, return 401 failure
	}

	return resp, err
}

func (t *authTransport) bearerAuth(realm, service, scope string) (string, error) {
	cacheKey := fmt.Sprintf("%s!%s!%s", realm, service, scope)
	if token := t.tokenCache[cacheKey]; token != "" {
		return token, nil
	}

	url, err := url.Parse(realm)
	if err != nil {
		return "", err
	}

	q := url.Query()
	q.Set("service", service)
	if scope != "" {
		q.Set("scope", scope)
	}
	url.RawQuery = q.Encode()

	req, err := http.NewRequest(http.MethodGet, url.String(), nil)
	if t.Username != "" || t.Password != "" {
		req.SetBasicAuth(t.Username, t.Password)
	}

	glog.V(3).Infof("Performing oauth request to %s", req.URL)
	resp, err := t.Client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Auth request returned %s", resp.Status)
	}

	type authToken struct {
		Token string `json:"token"`
	}
	var token authToken
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return "", err
	}
	glog.V(4).Infof("Got oauth token %q", token.Token)
	t.tokenCache[cacheKey] = token.Token
	return token.Token, err
}
