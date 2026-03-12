package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/atlanticbt/magecli/pkg/cmdutil"
)

type apiOptions struct {
	Method  string
	Input   string
	Fields  []string
	Headers []string
	Params  []string
}

func NewCmdAPI(f *cmdutil.Factory) *cobra.Command {
	opts := &apiOptions{}
	cmd := &cobra.Command{
		Use:   "api <path>",
		Short: "Make raw Magento API requests",
		Long: `Call Magento REST APIs directly for endpoints without first-class commands.

Examples:
  magecli api /V1/store/storeViews
  magecli api /V1/products -X GET --param "searchCriteria[pageSize]=5"
  magecli api /V1/cmsPage/1`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAPI(cmd, f, opts, args[0])
		},
	}

	cmd.Flags().StringVarP(&opts.Method, "method", "X", "", "HTTP method (default GET, or POST with body)")
	cmd.Flags().StringVarP(&opts.Input, "input", "d", "", "JSON string for the request body")
	cmd.Flags().StringArrayVarP(&opts.Fields, "field", "F", nil, "Add JSON body field (key=value)")
	cmd.Flags().StringArrayVarP(&opts.Headers, "header", "H", nil, "Add an HTTP request header (Key: Value)")
	cmd.Flags().StringArrayVarP(&opts.Params, "param", "P", nil, "Append query parameter (key=value)")

	return cmd
}

func runAPI(cmd *cobra.Command, f *cmdutil.Factory, opts *apiOptions, path string) error {
	method := strings.ToUpper(strings.TrimSpace(opts.Method))

	var body any
	if len(opts.Fields) > 0 && opts.Input != "" {
		return fmt.Errorf("--field and --input flags cannot be combined")
	}

	if len(opts.Fields) > 0 {
		payload := make(map[string]any, len(opts.Fields))
		for _, field := range opts.Fields {
			key, value, err := parseKeyValue(field)
			if err != nil {
				return fmt.Errorf("parse field %q: %w", field, err)
			}
			payload[key] = inferJSONValue(value)
		}
		body = payload
	} else if strings.TrimSpace(opts.Input) != "" {
		raw := json.RawMessage(opts.Input)
		body = raw
	}

	if method == "" {
		if body != nil {
			method = "POST"
		} else {
			method = "GET"
		}
	}

	override := cmdutil.FlagValue(cmd, "context")
	_, ctx, host, err := cmdutil.ResolveContext(f, cmd, override)
	if err != nil {
		return err
	}

	if method != "GET" && method != "HEAD" && !ctx.AllowWrites {
		return fmt.Errorf("write operations (%s) are not allowed on this context; recreate with --allow-writes or use a context that permits writes", method)
	}

	httpClient, err := cmdutil.NewHTTPClient(host, ctx.StoreCode)
	if err != nil {
		return err
	}

	req, err := httpClient.NewRequest(cmd.Context(), method, path, body)
	if err != nil {
		return err
	}

	for _, header := range opts.Headers {
		key, value, err := parseHeader(header)
		if err != nil {
			return err
		}
		req.Header.Set(key, value)
	}

	if len(opts.Params) > 0 {
		query := req.URL.Query()
		for _, param := range opts.Params {
			key, value, err := parseKeyValue(param)
			if err != nil {
				return fmt.Errorf("parse param %q: %w", param, err)
			}
			query.Add(key, value)
		}
		req.URL.RawQuery = query.Encode()
	}

	ios, err := f.Streams()
	if err != nil {
		return err
	}

	settings, err := cmdutil.ResolveOutputSettings(cmd)
	if err != nil {
		return err
	}

	needsStructuredOutput := settings.Format != "" || settings.JQ != "" || settings.Template != ""
	if !needsStructuredOutput {
		return httpClient.Do(req, ios.Out)
	}

	var buf bytes.Buffer
	if err := httpClient.Do(req, &buf); err != nil {
		return err
	}

	var data any
	if buf.Len() > 0 {
		decoder := json.NewDecoder(bytes.NewReader(buf.Bytes()))
		decoder.UseNumber()
		if err := decoder.Decode(&data); err != nil {
			return fmt.Errorf("response is not valid JSON: %w", err)
		}
	}

	return cmdutil.WriteOutput(cmd, ios.Out, data, func() error {
		if buf.Len() == 0 {
			return nil
		}
		_, err := ios.Out.Write(buf.Bytes())
		return err
	})
}

func parseKeyValue(input string) (string, string, error) {
	parts := strings.SplitN(input, "=", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("expected key=value format")
	}
	return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]), nil
}

func parseHeader(input string) (string, string, error) {
	parts := strings.SplitN(input, ":", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("header must be in \"Key: Value\" format")
	}
	return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]), nil
}

func inferJSONValue(raw string) any {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return ""
	}
	var v any
	if err := json.Unmarshal([]byte(trimmed), &v); err == nil {
		return v
	}
	return raw
}
