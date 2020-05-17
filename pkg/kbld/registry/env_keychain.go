package registry

import (
	"fmt"
	"os"
	"strings"
	"sync"

	regauthn "github.com/google/go-containerregistry/pkg/authn"
	regname "github.com/google/go-containerregistry/pkg/name"
)

// Env vars with optional suffix:
//   export KBLD_REGISTRY_HOSTNAME_0=...
//   export KBLD_REGISTRY_USERNAME_0=...
//   export KBLD_REGISTRY_PASSWORD_0=...

type EnvKeychain struct {
	globalPrefix string

	infos       []envKeychainInfo
	collectErr  error
	collected   bool
	collectLock sync.Mutex
}

var _ regauthn.Keychain = &EnvKeychain{}

func NewEnvKeychain(globalPrefix string) *EnvKeychain {
	return &EnvKeychain{globalPrefix: globalPrefix}
}

func (k *EnvKeychain) Resolve(target regauthn.Resource) (regauthn.Authenticator, error) {
	infos, err := k.collect()
	if err != nil {
		return nil, err
	}

	for _, info := range infos {
		if info.Hostname == target.RegistryStr() {
			return regauthn.FromConfig(regauthn.AuthConfig{
				Username:      info.Username,
				Password:      info.Password,
				IdentityToken: info.IdentityToken,
				RegistryToken: info.RegistryToken,
			}), nil
		}
	}
	return regauthn.Anonymous, nil
}

type envKeychainInfo struct {
	Hostname      string
	Username      string
	Password      string
	IdentityToken string
	RegistryToken string
}

func (k *EnvKeychain) collect() ([]envKeychainInfo, error) {
	k.collectLock.Lock()
	defer k.collectLock.Unlock()

	if k.collected || k.collectErr != nil {
		return append([]envKeychainInfo{}, k.infos...), k.collectErr
	}

	const (
		sep = "_"
	)

	funcsMap := map[string]func(*envKeychainInfo, string) error{
		"HOSTNAME": func(info *envKeychainInfo, val string) error {
			registry, err := regname.NewRegistry(val, regname.StrictValidation)
			if err != nil {
				return fmt.Errorf("Parsing registry hostname: %s (e.g. gcr.io, index.docker.io)", err)
			}
			info.Hostname = registry.RegistryStr()
			return nil
		},
		"USERNAME": func(info *envKeychainInfo, val string) error {
			info.Username = val
			return nil
		},
		"PASSWORD": func(info *envKeychainInfo, val string) error {
			info.Password = val
			return nil
		},
		"IDENTITY_TOKEN": func(info *envKeychainInfo, val string) error {
			info.IdentityToken = val
			return nil
		},
		"REGISTRY_TOKEN": func(info *envKeychainInfo, val string) error {
			info.RegistryToken = val
			return nil
		},
	}

	defaultInfo := envKeychainInfo{}
	infos := map[string]envKeychainInfo{}

	for _, env := range os.Environ() {
		pieces := strings.SplitN(env, "=", 2)
		if len(pieces) != 2 {
			continue
		}

		var matched bool

		for key, updateFunc := range funcsMap {
			switch {
			case pieces[0] == k.globalPrefix+sep+key:
				matched = true
				err := updateFunc(&defaultInfo, pieces[1])
				if err != nil {
					k.collectErr = err
					return nil, k.collectErr
				}

			case strings.HasPrefix(pieces[0], k.globalPrefix+sep+key+sep):
				matched = true
				prefix := strings.TrimPrefix(pieces[0], k.globalPrefix+sep+key+sep)
				info := infos[prefix]
				err := updateFunc(&info, pieces[1])
				if err != nil {
					k.collectErr = err
					return nil, k.collectErr
				}
				infos[prefix] = info
			}
		}

		if !matched && strings.HasPrefix(pieces[0], k.globalPrefix+sep) {
			k.collectErr = fmt.Errorf("Unknown env variable '%s'", pieces[0])
			return nil, k.collectErr
		}
	}

	var result []envKeychainInfo

	if defaultInfo != (envKeychainInfo{}) {
		result = append(result, defaultInfo)
	}
	for _, info := range infos {
		result = append(result, info)
	}

	k.infos = result
	k.collected = true

	return append([]envKeychainInfo{}, k.infos...), nil
}
