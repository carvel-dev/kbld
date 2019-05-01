package image

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"net/http"
	"time"

	regauthn "github.com/google/go-containerregistry/pkg/authn"
	regname "github.com/google/go-containerregistry/pkg/name"
	regv1 "github.com/google/go-containerregistry/pkg/v1"
	regremote "github.com/google/go-containerregistry/pkg/v1/remote"
)

// ResolvedImage respresents an image that will be resolved into url+digest
type ResolvedImage struct {
	url string
}

func (i ResolvedImage) URL() (string, error) {
	tag, err := regname.NewTag(i.url, regname.WeakValidation)
	if err != nil {
		return "", err
	}

	httpTran, err := i.newHTTPTransport()
	if err != nil {
		return "", err
	}

	opts := []regremote.ImageOption{
		regremote.WithTransport(httpTran),
		regremote.WithAuthFromKeychain(regauthn.DefaultKeychain),
	}

	imgDescriptor, err := regremote.Get(tag, opts...)
	if err != nil {
		return "", err
	}

	return NewDigestedImageFromParts(tag.Repository.String(), imgDescriptor.Digest.String()).URL()
}

func (ResolvedImage) newHTTPTransport() (*http.Transport, error) {
	pool, err := x509.SystemCertPool()
	if err != nil {
		pool = x509.NewCertPool()
	}

	// if crt, err := ioutil.ReadFile(path); err != nil {
	// 	return nil, err
	// } else if ok := pool.AppendCertsFromPEM(crt); !ok {
	// 	return nil, errors.New("failed to append k8s cert bundle to cert pool")
	// }

	// Copied from https://github.com/golang/go/blob/release-branch.go1.12/src/net/http/transport.go#L42-L53
	// We want to use the DefaultTransport but change its TLSClientConfig. There
	// isn't a clean way to do this yet: https://github.com/golang/go/issues/26013
	return &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		// Use the cert pool with k8s cert bundle appended.
		TLSClientConfig: &tls.Config{
			RootCAs: pool,
		},
	}, nil
}

func (i ResolvedImage) Write(ref regname.Reference, img regv1.Image) error {
	httpTran, err := i.newHTTPTransport()
	if err != nil {
		return err
	}

	authz, err := regauthn.DefaultKeychain.Resolve(ref.Context().Registry)
	if err != nil {
		return fmt.Errorf("Getting authz details: %s", err)
	}

	err = i.retry(func() error { return regremote.Write(ref, img, authz, httpTran) })
	if err != nil {
		return fmt.Errorf("Writing image: %s", err)
	}

	return nil
}

func (i ResolvedImage) WriteIndex(ref regname.Reference, idx regv1.ImageIndex) error {
	httpTran, err := i.newHTTPTransport()
	if err != nil {
		return err
	}

	authz, err := regauthn.DefaultKeychain.Resolve(ref.Context().Registry)
	if err != nil {
		return fmt.Errorf("Getting authz details: %s", err)
	}

	err = i.retry(func() error { return regremote.WriteIndex(ref, idx, authz, httpTran) })
	if err != nil {
		return fmt.Errorf("Writing image index: %s", err)
	}

	return nil
}

func (i ResolvedImage) retry(doFunc func() error) error {
	var lastErr error
	for i := 0; i < 5; i++ {
		lastErr = doFunc()
		if lastErr == nil {
			return nil
		}
		time.Sleep(1 * time.Second)
	}
	return fmt.Errorf("Retried 5 times: %s", lastErr)
}
