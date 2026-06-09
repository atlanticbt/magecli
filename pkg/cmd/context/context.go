package context

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/atlanticbt/magecli/internal/config"
	"github.com/atlanticbt/magecli/internal/secret"
	"github.com/atlanticbt/magecli/pkg/cmdutil"
)

func NewCmdContext(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "context",
		Short: "Manage Magento CLI contexts",
	}
	cmd.AddCommand(newCreateCmd(f))
	cmd.AddCommand(newUseCmd(f))
	cmd.AddCommand(newListCmd(f))
	cmd.AddCommand(newDeleteCmd(f))
	return cmd
}

type createOptions struct {
	Host        string
	StoreCode   string
	SetActive   bool
	AllowWrites bool
}

func newCreateCmd(f *cmdutil.Factory) *cobra.Command {
	opts := &createOptions{}
	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a new CLI context",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCreate(cmd, f, args[0], opts)
		},
	}
	cmd.Flags().StringVar(&opts.Host, "host", "", "Host key or base URL (required)")
	cmd.Flags().StringVar(&opts.StoreCode, "store-code", "default", "Magento store code")
	cmd.Flags().BoolVar(&opts.SetActive, "set-active", false, "Set the new context as active")
	cmd.Flags().BoolVar(&opts.AllowWrites, "allow-writes", false, "Allow non-GET API requests (POST, PUT, DELETE)")
	return cmd
}

func runCreate(cmd *cobra.Command, f *cmdutil.Factory, name string, opts *createOptions) error {
	ios, err := f.Streams()
	if err != nil {
		return err
	}

	cfg, err := f.ResolveConfig()
	if err != nil {
		return err
	}

	hostKey := strings.TrimSpace(opts.Host)
	if hostKey == "" {
		return fmt.Errorf("--host is required")
	}

	if _, ok := cfg.Hosts[hostKey]; !ok {
		baseURL, err := cmdutil.NormalizeBaseURL(hostKey)
		if err != nil {
			return fmt.Errorf("host %q not found and is not a valid URL; run `%s auth login` first", opts.Host, f.ExecutableName)
		}
		key, err := cmdutil.HostKeyFromURL(baseURL)
		if err != nil {
			return err
		}
		hostKey = key
		if _, ok := cfg.Hosts[key]; !ok {
			// Register the host from the provided URL only when the credential
			// comes from the MAGECLI_TOKEN env var, so headless bootstrap works
			// without a prior `auth login`. Otherwise reject the unknown host
			// to keep typo protection.
			if secret.TokenFromEnv() == "" {
				return fmt.Errorf("host %q not found; run `%s auth login` first (or set %s to register it directly)", opts.Host, f.ExecutableName, secret.EnvToken)
			}
			cfg.SetHost(key, &config.Host{BaseURL: baseURL})
			_, _ = fmt.Fprintf(ios.Out, "Registered host %s\n", baseURL)
		}
	}

	ctx := &config.Context{
		Host:        hostKey,
		StoreCode:   opts.StoreCode,
		AllowWrites: opts.AllowWrites,
	}

	cfg.SetContext(name, ctx)

	if opts.SetActive || cfg.ActiveContext == "" {
		if err := cfg.SetActiveContext(name); err != nil {
			return err
		}
	}

	if err := cfg.Save(); err != nil {
		return err
	}

	_, _ = fmt.Fprintf(ios.Out, "Created context %q (host: %s, store: %s)\n", name, hostKey, opts.StoreCode)
	if cfg.ActiveContext == name {
		_, _ = fmt.Fprintf(ios.Out, "Context %q is now active\n", name)
	}
	return nil
}

func newUseCmd(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "use <name>",
		Short: "Activate an existing context",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUse(cmd, f, args[0])
		},
	}
}

func runUse(cmd *cobra.Command, f *cmdutil.Factory, name string) error {
	ios, err := f.Streams()
	if err != nil {
		return err
	}
	cfg, err := f.ResolveConfig()
	if err != nil {
		return err
	}
	if err := cfg.SetActiveContext(name); err != nil {
		return err
	}
	if err := cfg.Save(); err != nil {
		return err
	}
	_, _ = fmt.Fprintf(ios.Out, "Activated context %q\n", name)
	return nil
}

func newListCmd(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List available contexts",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(cmd, f)
		},
	}
}

func runList(cmd *cobra.Command, f *cmdutil.Factory) error {
	ios, err := f.Streams()
	if err != nil {
		return err
	}
	cfg, err := f.ResolveConfig()
	if err != nil {
		return err
	}

	type summary struct {
		Name      string `json:"name"`
		Host      string `json:"host"`
		StoreCode string `json:"store_code,omitempty"`
		Active    bool   `json:"active"`
	}

	var names []string
	for name := range cfg.Contexts {
		names = append(names, name)
	}
	sort.Strings(names)

	var contexts []summary
	for _, name := range names {
		ctx := cfg.Contexts[name]
		contexts = append(contexts, summary{
			Name:      name,
			Host:      ctx.Host,
			StoreCode: ctx.StoreCode,
			Active:    cfg.ActiveContext == name,
		})
	}

	payload := struct {
		Active   string    `json:"active_context,omitempty"`
		Contexts []summary `json:"contexts"`
	}{
		Active:   cfg.ActiveContext,
		Contexts: contexts,
	}

	return cmdutil.WriteOutput(cmd, ios.Out, payload, func() error {
		if len(contexts) == 0 {
			_, _ = fmt.Fprintf(ios.Out, "No contexts configured. Use `%s context create` to add one.\n", f.ExecutableName)
			return nil
		}
		for _, ctx := range contexts {
			marker := " "
			if ctx.Active {
				marker = "*"
			}
			_, _ = fmt.Fprintf(ios.Out, "%s %s (host: %s, store: %s)\n", marker, ctx.Name, ctx.Host, ctx.StoreCode)
		}
		return nil
	})
}

func newDeleteCmd(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:     "delete <name>",
		Aliases: []string{"rm"},
		Short:   "Delete a context",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDelete(cmd, f, args[0])
		},
	}
}

func runDelete(cmd *cobra.Command, f *cmdutil.Factory, name string) error {
	ios, err := f.Streams()
	if err != nil {
		return err
	}
	cfg, err := f.ResolveConfig()
	if err != nil {
		return err
	}
	if _, err := cfg.Context(name); err != nil {
		return err
	}
	cfg.DeleteContext(name)
	if err := cfg.Save(); err != nil {
		return err
	}
	_, _ = fmt.Fprintf(ios.Out, "Deleted context %q\n", name)
	return nil
}
