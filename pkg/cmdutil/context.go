package cmdutil

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/atlanticbt/magecli/internal/config"
	"github.com/atlanticbt/magecli/internal/secret"
	"github.com/atlanticbt/magecli/pkg/httpx"
	"github.com/atlanticbt/magecli/pkg/magento"
)

// ResolveContext fetches the context and host configuration given an optional
// override name (typically provided via --context).
func ResolveContext(f *Factory, cmd *cobra.Command, override string) (string, *config.Context, *config.Host, error) {
	cfg, err := f.ResolveConfig()
	if err != nil {
		return "", nil, nil, err
	}

	contextName := override
	if contextName == "" {
		contextName = cfg.ActiveContext
	}
	if contextName == "" {
		return "", nil, nil, fmt.Errorf("no active context; run `%s context use <name>`", f.ExecutableName)
	}

	ctx, err := cfg.Context(contextName)
	if err != nil {
		return "", nil, nil, err
	}
	if ctx.Host == "" {
		return "", nil, nil, fmt.Errorf("context %q has no host configured", contextName)
	}

	host, err := cfg.Host(ctx.Host)
	if err != nil {
		return "", nil, nil, err
	}

	if err := loadHostToken(f.ExecutableName, ctx.Host, host); err != nil {
		return "", nil, nil, err
	}

	return contextName, ctx, host, nil
}

// ResolveHost locates a host configuration using optional context or host overrides.
func ResolveHost(f *Factory, contextOverride, hostOverride string) (string, *config.Host, error) {
	cfg, err := f.ResolveConfig()
	if err != nil {
		return "", nil, err
	}

	hostIdentifier := strings.TrimSpace(hostOverride)
	if hostIdentifier != "" {
		if host, ok := cfg.Hosts[hostIdentifier]; ok {
			if err := loadHostToken(f.ExecutableName, hostIdentifier, host); err != nil {
				return "", nil, err
			}
			return hostIdentifier, host, nil
		}
		baseURL, err := NormalizeBaseURL(hostIdentifier)
		if err == nil {
			if key, err := HostKeyFromURL(baseURL); err == nil {
				if host, ok := cfg.Hosts[key]; ok {
					if err := loadHostToken(f.ExecutableName, key, host); err != nil {
						return "", nil, err
					}
					return key, host, nil
				}
			}
		}
		return "", nil, fmt.Errorf("host %q not found; run `%s auth login` first", hostIdentifier, f.ExecutableName)
	}

	contextName := strings.TrimSpace(contextOverride)
	if contextName == "" {
		contextName = cfg.ActiveContext
	}
	if contextName != "" {
		ctx, err := cfg.Context(contextName)
		if err != nil {
			return "", nil, err
		}
		if ctx.Host == "" {
			return "", nil, fmt.Errorf("context %q has no host configured", contextName)
		}
		host, err := cfg.Host(ctx.Host)
		if err != nil {
			return "", nil, err
		}
		if err := loadHostToken(f.ExecutableName, ctx.Host, host); err != nil {
			return "", nil, err
		}
		return ctx.Host, host, nil
	}

	switch len(cfg.Hosts) {
	case 0:
		return "", nil, fmt.Errorf("no hosts configured; run `%s auth login` first", f.ExecutableName)
	case 1:
		for key, host := range cfg.Hosts {
			if err := loadHostToken(f.ExecutableName, key, host); err != nil {
				return "", nil, err
			}
			return key, host, nil
		}
	default:
		var keys []string
		for key := range cfg.Hosts {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		return "", nil, fmt.Errorf("multiple hosts configured (%s); specify --context", strings.Join(keys, ", "))
	}

	return "", nil, fmt.Errorf("failed to resolve host configuration")
}

// FlagValue returns the value for the named flag if it exists.
func FlagValue(cmd *cobra.Command, name string) string {
	flag := cmd.Flags().Lookup(name)
	if flag == nil {
		return ""
	}
	return flag.Value.String()
}

// NewMagentoClient constructs a Magento API client from the resolved context.
func NewMagentoClient(host *config.Host, storeCode string) (*magento.Client, error) {
	if host == nil {
		return nil, fmt.Errorf("missing host configuration")
	}
	if host.BaseURL == "" {
		return nil, fmt.Errorf("host has no base URL configured")
	}

	sc := storeCode
	if sc == "" {
		sc = host.StoreCode
	}

	return magento.New(magento.ClientOptions{
		BaseURL:     host.BaseURL,
		Token:       host.Token,
		StoreCode:   sc,
		EnableCache: true,
		Retry: httpx.RetryPolicy{
			MaxAttempts:    4,
			InitialBackoff: 250 * time.Millisecond,
			MaxBackoff:     2 * time.Second,
		},
	})
}

// NewHTTPClient constructs a raw HTTP client for the configured host.
func NewHTTPClient(host *config.Host, storeCode string) (*httpx.Client, error) {
	client, err := NewMagentoClient(host, storeCode)
	if err != nil {
		return nil, err
	}
	return client.HTTP(), nil
}

func loadHostToken(executable, hostKey string, host *config.Host) error {
	if host == nil {
		return fmt.Errorf("host %q not configured", hostKey)
	}

	if envToken := secret.TokenFromEnv(); envToken != "" {
		host.Token = envToken
		return nil
	}

	if host.Token != "" {
		return nil
	}

	opts := []secret.Option{}
	if host.AllowInsecureStore {
		opts = append(opts, secret.WithAllowFileFallback(true))
	}

	store, err := secret.Open(opts...)
	if err != nil {
		if secret.IsNoKeyringError(err) {
			return fmt.Errorf("no OS keychain backend available for host %q; rerun `%s auth login %s --allow-insecure-store` or set MAGECLI_ALLOW_INSECURE_STORE=1: %w", hostKey, executable, hostKey, err)
		}
		return err
	}

	token, err := store.Get(secret.TokenKey(hostKey))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			target := host.BaseURL
			if target == "" {
				target = hostKey
			}
			return fmt.Errorf("credentials for host %q not found; run `%s auth login %s`", hostKey, executable, target)
		}
		return err
	}

	host.Token = token
	return nil
}
