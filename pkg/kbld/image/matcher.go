package image

import (
	"fmt"
	"regexp"

	ctlconf "github.com/k14s/kbld/pkg/kbld/config"
)

var (
	approximateRefRegexp = regexp.MustCompile(`\A(.+?)(:[A-Za-z0-9_\-\.]+)?(@.+:.+)?\z`)
)

type Matcher struct {
	url string
}

func NewMatcher(url string) Matcher { return Matcher{url} }

func (m Matcher) Matches(ref ctlconf.ImageRef) bool {
	switch {
	case len(ref.Image) > 0:
		return ref.Image == m.url

	case len(ref.ImageRepo) > 0:
		return ref.ImageRepo == m.urlRepo()

	default:
		panic(fmt.Errorf("Missing image or imageRepo configuration"))
	}
}

func (m Matcher) urlRepo() string {
	// Not using go-containerregistry library to parse repository because
	// it does not expose "exact" original repository
	// (eg augments dockerhub images with index.docker.io, etc.);
	// hence, would like to be less surprising and match exactly
	matches := approximateRefRegexp.FindStringSubmatch(m.url)
	if len(matches) >= 1 {
		return matches[1]
	}
	return m.url
}
