package secret

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/99designs/keyring"
)

const serviceName = "magecli"

const (
	EnvToken         = "MAGECLI_TOKEN"
	envAllowInsecure = "MAGECLI_ALLOW_INSECURE_STORE"
	envPassphrase    = "MAGECLI_KEYRING_PASSPHRASE"
	envTimeout       = "MAGECLI_KEYRING_TIMEOUT"
	envBackend       = "KEYRING_BACKEND"
)

const (
	keyringTimeoutHeadless    = 3 * time.Second
	keyringTimeoutInteractive = 60 * time.Second
)

var ErrKeyringTimeout = errors.New("keyring operation timed out")

func TokenFromEnv() string {
	return strings.TrimSpace(os.Getenv(EnvToken))
}

func IsHeadless() bool {
	isSSH := os.Getenv("SSH_TTY") != "" || os.Getenv("SSH_CLIENT") != "" || os.Getenv("SSH_CONNECTION") != ""
	if isSSH {
		if runtime.GOOS == "darwin" || runtime.GOOS == "windows" {
			return true
		}
		hasDisplay := os.Getenv("DISPLAY") != "" || os.Getenv("WAYLAND_DISPLAY") != ""
		return !hasDisplay
	}
	if runtime.GOOS == "darwin" || runtime.GOOS == "windows" {
		return envEnabled(os.Getenv("CI"))
	}
	hasDisplay := os.Getenv("DISPLAY") != "" || os.Getenv("WAYLAND_DISPLAY") != ""
	hasDBus := os.Getenv("DBUS_SESSION_BUS_ADDRESS") != ""
	return !hasDisplay && !hasDBus
}

func keyringTimeout() time.Duration {
	if d, ok := parseTimeoutEnv(strings.TrimSpace(os.Getenv(envTimeout))); ok {
		return d
	}
	if IsHeadless() {
		return keyringTimeoutHeadless
	}
	return keyringTimeoutInteractive
}

func parseTimeoutEnv(raw string) (time.Duration, bool) {
	if raw == "" {
		return 0, false
	}
	if d, err := time.ParseDuration(raw); err == nil && d > 0 {
		return d, true
	}
	secs, err := strconv.Atoi(raw)
	if err != nil || secs <= 0 {
		return 0, false
	}
	return time.Duration(secs) * time.Second, true
}

func timeoutHint() string {
	if IsHeadless() {
		return fmt.Sprintf("keyring prompt may be blocked (headless/SSH environment?). Use --allow-insecure-store or set %s=1", envAllowInsecure)
	}
	return fmt.Sprintf("keyring prompt may need more time. Increase timeout via %s (e.g. 60s or 2m)", envTimeout)
}

type Store struct {
	kr keyring.Keyring
}

type openOptions struct {
	allowFile       bool
	passphrase      string
	allowedBackends []keyring.BackendType
	fileDir         string
}

type Option func(*openOptions)

func WithAllowFileFallback(enable bool) Option {
	return func(o *openOptions) { o.allowFile = enable }
}

func WithPassphrase(pass string) Option {
	return func(o *openOptions) {
		if pass != "" {
			o.passphrase = pass
		}
	}
}

func Open(opts ...Option) (*Store, error) {
	cfg := keyring.Config{ServiceName: serviceName}
	settings := openOptions{}

	if envEnabled(os.Getenv(envAllowInsecure)) {
		settings.allowFile = true
	}
	if pass := strings.TrimSpace(os.Getenv(envPassphrase)); pass != "" {
		settings.passphrase = pass
	}

	for _, opt := range opts {
		opt(&settings)
	}

	cfg.AllowedBackends = resolveAllowedBackends(settings)

	if usesFileBackend(cfg.AllowedBackends) {
		if err := configureFileBackend(&cfg, settings); err != nil {
			return nil, err
		}
	}

	kr, err := openKeyringWithTimeout(cfg)
	if err != nil {
		if errors.Is(err, ErrKeyringTimeout) {
			return nil, fmt.Errorf("open keyring: %w; %s", err, timeoutHint())
		}
		if errors.Is(err, keyring.ErrNoAvailImpl) && !usesFileBackend(cfg.AllowedBackends) {
			return nil, fmt.Errorf("open keyring: %w (set %s=1 or rerun with --allow-insecure-store to permit encrypted file fallback)", err, envAllowInsecure)
		}
		return nil, fmt.Errorf("open keyring: %w", err)
	}
	return &Store{kr: kr}, nil
}

func openKeyringWithTimeout(cfg keyring.Config) (keyring.Keyring, error) {
	type result struct {
		kr  keyring.Keyring
		err error
	}
	ch := make(chan result, 1)
	go func() {
		kr, err := keyring.Open(cfg)
		ch <- result{kr, err}
	}()
	ctx, cancel := context.WithTimeout(context.Background(), keyringTimeout())
	defer cancel()
	select {
	case res := <-ch:
		return res.kr, res.err
	case <-ctx.Done():
		return nil, ErrKeyringTimeout
	}
}

func (s *Store) Set(key, value string) error {
	if s == nil || s.kr == nil {
		return errors.New("secret store not initialized")
	}
	return s.withTimeout(func() error {
		return s.kr.Set(keyring.Item{
			Key:   key,
			Data:  []byte(value),
			Label: fmt.Sprintf("magecli %s", key),
		})
	})
}

func (s *Store) Get(key string) (string, error) {
	if s == nil || s.kr == nil {
		return "", errors.New("secret store not initialized")
	}
	var item keyring.Item
	err := s.withTimeout(func() error {
		var getErr error
		item, getErr = s.kr.Get(key)
		return getErr
	})
	if err != nil {
		if errors.Is(err, keyring.ErrKeyNotFound) {
			return "", os.ErrNotExist
		}
		return "", err
	}
	return string(item.Data), nil
}

func (s *Store) Delete(key string) error {
	if s == nil || s.kr == nil {
		return errors.New("secret store not initialized")
	}
	err := s.withTimeout(func() error { return s.kr.Remove(key) })
	if errors.Is(err, keyring.ErrKeyNotFound) {
		return nil
	}
	return err
}

func (s *Store) withTimeout(fn func() error) error {
	ch := make(chan error, 1)
	go func() { ch <- fn() }()
	ctx, cancel := context.WithTimeout(context.Background(), keyringTimeout())
	defer cancel()
	select {
	case err := <-ch:
		return err
	case <-ctx.Done():
		return fmt.Errorf("%w; %s", ErrKeyringTimeout, timeoutHint())
	}
}

func TokenKey(hostKey string) string {
	return fmt.Sprintf("host/%s/token", hostKey)
}

func IsNoKeyringError(err error) bool {
	return errors.Is(err, keyring.ErrNoAvailImpl)
}

func resolveAllowedBackends(opts openOptions) []keyring.BackendType {
	if len(opts.allowedBackends) > 0 {
		return opts.allowedBackends
	}
	if backendEnv := strings.TrimSpace(os.Getenv(envBackend)); backendEnv != "" {
		return parseBackendList(backendEnv, opts.allowFile)
	}
	backends := defaultBackends()
	if opts.allowFile {
		backends = append(backends, keyring.FileBackend)
	}
	return backends
}

func defaultBackends() []keyring.BackendType {
	switch runtime.GOOS {
	case "darwin":
		return []keyring.BackendType{keyring.KeychainBackend}
	case "windows":
		return []keyring.BackendType{keyring.WinCredBackend}
	default:
		if IsHeadless() {
			return []keyring.BackendType{keyring.KeyCtlBackend, keyring.PassBackend}
		}
		return []keyring.BackendType{
			keyring.SecretServiceBackend,
			keyring.KWalletBackend,
			keyring.KeyCtlBackend,
			keyring.PassBackend,
		}
	}
}

func parseBackendList(raw string, allowFile bool) []keyring.BackendType {
	parts := strings.Split(raw, ",")
	var backends []keyring.BackendType
	for _, part := range parts {
		switch strings.TrimSpace(strings.ToLower(part)) {
		case "keychain":
			backends = append(backends, keyring.KeychainBackend)
		case "wincred":
			backends = append(backends, keyring.WinCredBackend)
		case "secret-service", "secretservice":
			backends = append(backends, keyring.SecretServiceBackend)
		case "kwallet":
			backends = append(backends, keyring.KWalletBackend)
		case "keyctl":
			backends = append(backends, keyring.KeyCtlBackend)
		case "pass":
			backends = append(backends, keyring.PassBackend)
		case "file":
			backends = append(backends, keyring.FileBackend)
		}
	}
	if !allowFile {
		filtered := backends[:0]
		for _, b := range backends {
			if b != keyring.FileBackend {
				filtered = append(filtered, b)
			}
		}
		backends = filtered
	}
	return backends
}

func configureFileBackend(cfg *keyring.Config, opts openOptions) error {
	passphrase := opts.passphrase
	if passphrase == "" {
		if pwd := os.Getenv("KEYRING_FILE_PASSWORD"); pwd != "" {
			passphrase = pwd
		} else if pwd := os.Getenv("KEYRING_PASSWORD"); pwd != "" {
			passphrase = pwd
		}
	}
	switch {
	case passphrase != "":
		cfg.FilePasswordFunc = keyring.FixedStringPrompt(passphrase)
	case IsHeadless():
		return fmt.Errorf(
			"file backend requires a passphrase in headless environments; "+
				"set %s (or KEYRING_FILE_PASSWORD) or use %s to bypass the keyring entirely",
			envPassphrase, EnvToken,
		)
	default:
		cfg.FilePasswordFunc = keyring.TerminalPrompt
	}
	dir := opts.fileDir
	if dir == "" {
		if userDir, err := os.UserConfigDir(); err == nil {
			dir = filepath.Join(userDir, serviceName, "secrets")
		}
	}
	if dir != "" {
		cfg.FileDir = dir
	}
	return nil
}

func usesFileBackend(backends []keyring.BackendType) bool {
	for _, b := range backends {
		if b == keyring.FileBackend {
			return true
		}
	}
	return false
}

func envEnabled(raw string) bool {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}
