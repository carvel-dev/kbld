package image

import (
	"fmt"
	"regexp"

	regname "github.com/google/go-containerregistry/pkg/name"
	ctlconf "github.com/k14s/kbld/pkg/kbld/config"
)

var (
	approximateRefRegexp = regexp.MustCompile("\\A(.+)(:.+)?(@.+:.+)?\\z")
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
		urlRepo, err := m.urlRepo()
		if err != nil {
			return false
		}

		refRepo, err := regname.NewRepository(ref.ImageRepo, regname.WeakValidation)
		if err != nil {
			return false
		}

		// repo.Name() includes registry information;
		// it will also "normalize" dockerhub urls,
		// e.g. index.docker.io == docker.io
		return urlRepo.Name() == refRepo.Name()

	default:
		panic(fmt.Errorf("Missing image or imageRepo configuration"))
	}
}

func (m Matcher) urlRepo() (regname.Repository, error) {
	tag, tagErr := regname.NewTag(m.url, regname.WeakValidation)
	if tagErr == nil {
		return tag.Repository, nil
	}
	digest, digestErr := regname.NewDigest(m.url, regname.WeakValidation)
	if digestErr == nil {
		return digest.Repository, nil
	}
	repo, repoErr := regname.NewRepository(m.url, regname.WeakValidation)
	if repoErr == nil {
		return repo, nil
	}
	return regname.Repository{}, fmt.Errorf("Expected to successfully parse url '%s' as "+
		"tag ref, digest ref, or repository, but could not (errors: %s, %s, %s)",
		m.url, tagErr, digestErr, repoErr)
}

func (m Matcher) urlRepo2() string {
	matches := approximateRefRegexp.FindStringSubmatch(m.url)
	if len(matches) >= 1 {
		return matches[1]
	}
	return m.url
}
