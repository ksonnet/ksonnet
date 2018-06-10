// Copyright 2018 The ksonnet authors
//
//
//    Licensed under the Apache License, Version 2.0 (the "License");
//    you may not use this file except in compliance with the License.
//    You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS,
//    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//    See the License for the specific language governing permissions and
//    limitations under the License.

package github

import (
	"context"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/google/go-github/github"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

var (
	// DefaultClient is the default GitHub client.
	DefaultClient = &defaultGitHub{
		httpClient: defaultHTTPClient(),
		urlParse:   url.Parse,
	}
)

// Repo is a GitHub repo
type Repo struct {
	Org  string
	Repo string
}

// GitHub is an interface for communicating with GitHub.
type GitHub interface {
	ValidateURL(u string) error
	CommitSHA1(ctx context.Context, repo Repo, refSpec string) (string, error)
	Contents(ctx context.Context, repo Repo, path, sha1 string) (*github.RepositoryContent, []*github.RepositoryContent, error)
}

type httpClient interface {
	Head(string) (*http.Response, error)
}

func defaultHTTPClient() *http.Client {
	return &http.Client{
		Timeout: 10 * time.Second,
	}
}

type defaultGitHub struct {
	httpClient httpClient
	urlParse   func(string) (*url.URL, error)
}

var _ GitHub = (*defaultGitHub)(nil)

func (dg *defaultGitHub) ValidateURL(urlStr string) error {
	u, err := dg.urlParse(urlStr)
	if err != nil {
		return errors.Wrap(err, "parsing URL")
	}

	if u.Scheme == "" {
		u.Scheme = "https"
	}

	if !strings.HasPrefix(u.Path, "registry.yaml") {
		u.Path = u.Path + "/registry.yaml"
	}

	resp, err := dg.httpClient.Head(u.String())
	if err != nil {
		return errors.Wrapf(err, "verifying %q", u.String())
	}

	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("%q actual %d; expected %d", u.String(), resp.StatusCode, http.StatusOK)
	}

	return nil
}

func (dg *defaultGitHub) CommitSHA1(ctx context.Context, repo Repo, refSpec string) (string, error) {
	if refSpec == "" {
		refSpec = "master"
	}

	logrus.Debugf("github: fetching SHA1 for %s/%s - %s", repo.Org, repo.Repo, refSpec)
	sha, _, err := dg.client().Repositories.GetCommitSHA1(ctx, repo.Org, repo.Repo, refSpec, "")
	return sha, err
}

func (dg *defaultGitHub) Contents(ctx context.Context, repo Repo, path, sha1 string) (*github.RepositoryContent, []*github.RepositoryContent, error) {
	logrus.Debugf("github: fetching contents for %s/%s/%s - %s", repo.Org, repo.Repo, path, sha1)
	opts := &github.RepositoryContentGetOptions{Ref: sha1}

	file, dir, _, err := dg.client().Repositories.GetContents(ctx, repo.Org, repo.Repo, path, opts)
	return file, dir, err
}

func (dg *defaultGitHub) client() *github.Client {
	var hc *http.Client

	ght := os.Getenv("GITHUB_TOKEN")
	if len(ght) > 0 {
		ctx := context.Background()
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: ght},
		)
		hc = oauth2.NewClient(ctx, ts)
	}

	return github.NewClient(hc)
}
