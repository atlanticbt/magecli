package auth

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/atlanticbt/magecli/internal/config"
	"github.com/atlanticbt/magecli/internal/secret"
	"github.com/atlanticbt/magecli/pkg/cmdutil"
	"github.com/atlanticbt/magecli/pkg/iostreams"
)

func NewCmdAuth(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Manage Magento store authentication",
	}
	cmd.AddCommand(newLoginCmd(f))
	cmd.AddCommand(newStatusCmd(f))
	cmd.AddCommand(newLogoutCmd(f))
	return cmd
}

type loginOptions struct {
	Host               string
	Token              string
	AllowInsecureStore bool
}

func newLoginCmd(f *cmdutil.Factory) *cobra.Command {
	opts := &loginOptions{}
	cmd := &cobra.Command{
		Use:   "login <host>",
		Short: "Store a Magento Integration bearer token",
		Long: `Authenticate against a Magento 2 store using an Integration bearer token.

Create the token in Magento Admin > System > Integrations. The token is stored
in the OS keyring. Use MAGECLI_TOKEN env var to bypass the keyring.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				opts.Host = args[0]
			}
			return runLogin(cmd, f, opts)
		},
	}
	cmd.Flags().StringVar(&opts.Token, "token", "", "Integration bearer token")
	cmd.Flags().BoolVar(&opts.AllowInsecureStore, "allow-insecure-store", false, "Allow encrypted fallback when no OS keychain is available")
	return cmd
}

func runLogin(cmd *cobra.Command, f *cmdutil.Factory, opts *loginOptions) error {
	if secret.TokenFromEnv() != "" {
		return fmt.Errorf("%s environment variable is set; token is externally managed", secret.EnvToken)
	}

	ios, err := f.Streams()
	if err != nil {
		return err
	}

	reader := bufio.NewReader(ios.In)

	if opts.Host == "" {
		if !isTerminal(ios.In) {
			return fmt.Errorf("host is required when not running in a TTY")
		}
		opts.Host, err = promptString(reader, ios.Out, "Magento store URL (e.g. https://store.example.com)")
		if err != nil {
			return err
		}
	}

	baseURL, err := cmdutil.NormalizeBaseURL(opts.Host)
	if err != nil {
		return err
	}

	hostKey, err := cmdutil.HostKeyFromURL(baseURL)
	if err != nil {
		return err
	}

	if opts.Token == "" {
		if !isTerminal(ios.In) {
			return fmt.Errorf("--token is required when not running in a TTY")
		}
		opts.Token, err = promptSecret(ios, "Integration Bearer Token")
		if err != nil {
			return err
		}
	}

	// Store the token
	storeOpts := []secret.Option{}
	if opts.AllowInsecureStore {
		storeOpts = append(storeOpts, secret.WithAllowFileFallback(true))
	}
	store, err := secret.Open(storeOpts...)
	if err != nil {
		return fmt.Errorf("store token: %w", err)
	}
	if err := store.Set(secret.TokenKey(hostKey), opts.Token); err != nil {
		return fmt.Errorf("store token: %w", err)
	}

	cfg, err := f.ResolveConfig()
	if err != nil {
		return err
	}

	cfg.SetHost(hostKey, &config.Host{
		BaseURL:            baseURL,
		AllowInsecureStore: opts.AllowInsecureStore,
	})

	if err := cfg.Save(); err != nil {
		return err
	}

	_, _ = fmt.Fprintf(ios.Out, "Logged in to %s\n", baseURL)
	return nil
}

func newStatusCmd(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show authentication status for configured hosts",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatus(cmd, f)
		},
	}
}

func runStatus(cmd *cobra.Command, f *cmdutil.Factory) error {
	ios, err := f.Streams()
	if err != nil {
		return err
	}

	cfg, err := f.ResolveConfig()
	if err != nil {
		return err
	}

	type hostSummary struct {
		Key         string `json:"key"`
		BaseURL     string `json:"base_url"`
		TokenSource string `json:"token_source"`
	}

	var hostKeys []string
	for key := range cfg.Hosts {
		hostKeys = append(hostKeys, key)
	}
	sort.Strings(hostKeys)

	tokenSource := "keyring"
	if secret.TokenFromEnv() != "" {
		tokenSource = secret.EnvToken
	}

	var hosts []hostSummary
	for _, key := range hostKeys {
		h := cfg.Hosts[key]
		hosts = append(hosts, hostSummary{
			Key:         key,
			BaseURL:     h.BaseURL,
			TokenSource: tokenSource,
		})
	}

	payload := struct {
		ActiveContext string        `json:"active_context,omitempty"`
		Hosts        []hostSummary `json:"hosts"`
	}{
		ActiveContext: cfg.ActiveContext,
		Hosts:        hosts,
	}

	return cmdutil.WriteOutput(cmd, ios.Out, payload, func() error {
		if len(hosts) == 0 {
			_, _ = fmt.Fprintln(ios.Out, "No hosts configured. Run `magecli auth login` to add one.")
			return nil
		}
		_, _ = fmt.Fprintln(ios.Out, "Hosts:")
		for _, h := range hosts {
			_, _ = fmt.Fprintf(ios.Out, "  %s\n", h.BaseURL)
			_, _ = fmt.Fprintf(ios.Out, "    token source: %s\n", h.TokenSource)
		}
		return nil
	})
}

func newLogoutCmd(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "logout <host>",
		Short: "Remove stored credentials for a host",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLogout(cmd, f, args[0])
		},
	}
}

func runLogout(cmd *cobra.Command, f *cmdutil.Factory, hostArg string) error {
	if secret.TokenFromEnv() != "" {
		return fmt.Errorf("%s environment variable is set; unset it to use auth logout", secret.EnvToken)
	}

	ios, err := f.Streams()
	if err != nil {
		return err
	}

	cfg, err := f.ResolveConfig()
	if err != nil {
		return err
	}

	key := strings.TrimSpace(hostArg)
	if _, ok := cfg.Hosts[key]; !ok {
		baseURL, err := cmdutil.NormalizeBaseURL(key)
		if err != nil {
			return fmt.Errorf("unknown host %q", key)
		}
		key, err = cmdutil.HostKeyFromURL(baseURL)
		if err != nil {
			return err
		}
		if _, ok := cfg.Hosts[key]; !ok {
			return fmt.Errorf("host %q not found in configuration", hostArg)
		}
	}

	host := cfg.Hosts[key]
	storeOpts := []secret.Option{}
	if host.AllowInsecureStore {
		storeOpts = append(storeOpts, secret.WithAllowFileFallback(true))
	}
	store, err := secret.Open(storeOpts...)
	if err != nil {
		return fmt.Errorf("delete credentials: %w", err)
	}
	if err := store.Delete(secret.TokenKey(key)); err != nil {
		return fmt.Errorf("delete credentials: %w", err)
	}

	cfg.DeleteHost(key)
	for name, ctx := range cfg.Contexts {
		if ctx.Host == key {
			cfg.DeleteContext(name)
		}
	}

	if err := cfg.Save(); err != nil {
		return err
	}

	_, _ = fmt.Fprintf(ios.Out, "Removed credentials for %s\n", key)
	return nil
}

func promptString(reader *bufio.Reader, out io.Writer, label string) (string, error) {
	_, _ = fmt.Fprintf(out, "%s: ", label)
	value, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(value), nil
}

func promptSecret(ios *iostreams.IOStreams, label string) (string, error) {
	file, ok := ios.In.(*os.File)
	if ok && term.IsTerminal(int(file.Fd())) {
		_, _ = fmt.Fprintf(ios.Out, "%s: ", label)
		bytes, err := term.ReadPassword(int(file.Fd()))
		_, _ = fmt.Fprintln(ios.Out)
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(string(bytes)), nil
	}
	reader := bufio.NewReader(ios.In)
	return promptString(reader, ios.Out, label)
}

func isTerminal(in io.Reader) bool {
	file, ok := in.(*os.File)
	return ok && term.IsTerminal(int(file.Fd()))
}
