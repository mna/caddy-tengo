package caddytengo

import (
	"errors"
	"net/http"
	"os"
	"strconv"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"github.com/d5/tengo/v2"
	"github.com/d5/tengo/v2/stdlib"
	"go.uber.org/zap"
)

func init() {
	caddy.RegisterModule(Tengo{})
	httpcaddyfile.RegisterHandlerDirective("tengo", parseCaddyfile)
}

// Tengo implements an HTTP handler that runs a Tengo script to handle the
// request.
type Tengo struct {
	HandlerPath         string `json:"handler_path,omitempty"`
	MaxAllocs           int    `json:"max_allocs,omitempty"`
	MaxConstObjects     int    `json:"max_const_objects,omitempty"`
	ImportDir           string `json:"import_dir,omitempty"`
	CacheCompiledScript bool   `json:"cache_compiled_script,omitempty"`

	script *tengo.Compiled
	logger *zap.Logger
}

// CaddyModule returns the Caddy module information.
func (Tengo) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.handlers.tengo",
		New: func() caddy.Module { return new(Tengo) },
	}
}

// Provision implements caddy.Provisioner.
func (t *Tengo) Provision(ctx caddy.Context) error {
	if t.CacheCompiledScript {
		scr, err := t.loadHandlerScript()
		if err != nil {
			return err
		}

		cpl, err := scr.Compile()
		if err != nil {
			return err
		}
		t.script = cpl
	}
	t.logger = ctx.Logger(t)
	return nil
}

// Validate implements caddy.Validator.
func (t *Tengo) Validate() error {
	if t.HandlerPath == "" {
		return errors.New("the handler_path configuration option is required")
	}
	return nil
}

// ServeHTTP implements caddyhttp.MiddlewareHandler.
func (t Tengo) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
	if t.CacheCompiledScript {
		if err := t.script.RunContext(r.Context()); err != nil {
			return err
		}
	} else {
		scr, err := t.loadHandlerScript()
		if err != nil {
			return err
		}
		if _, err := scr.RunContext(r.Context()); err != nil {
			return err
		}
	}
	return next.ServeHTTP(w, r)
}

// UnmarshalCaddyfile implements caddyfile.Unmarshaler.
func (t *Tengo) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	asInt := func() (int, error) {
		var s string
		if !d.AllArgs(&s) {
			return 0, d.ArgErr()
		}
		i, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return 0, err
		}
		return int(i), nil
	}

	for d.Next() {
		for d.NextBlock(0) {
			switch field := d.Val(); field {
			case "max_allocs":
				i, err := asInt()
				if err != nil {
					return d.Errf("%s: %w", field, err)
				}
				t.MaxAllocs = i

			case "max_const_objects":
				i, err := asInt()
				if err != nil {
					return d.Errf("%s: %w", field, err)
				}
				t.MaxConstObjects = i

			case "cache_compiled_script":
				if d.CountRemainingArgs() > 0 {
					return d.Errf("%s: %w", field, d.ArgErr())
				}

			case "handler_path":
				if !d.Args(&t.HandlerPath) {
					return d.Errf("%s: %w", field, d.ArgErr())
				}

			case "import_dir":
				if !d.Args(&t.ImportDir) {
					return d.Errf("%s: %w", field, d.ArgErr())
				}

			default:
				return d.Errf("%s: unknown configuration option", field)
			}
		}
	}
	return nil
}

// parseCaddyfile unmarshals tokens from h into a new Lua.
func parseCaddyfile(h httpcaddyfile.Helper) (caddyhttp.MiddlewareHandler, error) {
	var t Tengo
	err := t.UnmarshalCaddyfile(h.Dispenser)
	return t, err
}

func (t *Tengo) loadHandlerScript() (*tengo.Script, error) {
	b, err := os.ReadFile(t.HandlerPath)
	if err != nil {
		return nil, err
	}

	scr := tengo.NewScript(b)
	scr.EnableFileImport(true)
	scr.SetImports(stdlib.GetModuleMap(stdlib.AllModuleNames()...))

	if t.ImportDir != "" {
		if err := scr.SetImportDir(t.ImportDir); err != nil {
			return nil, err
		}
	}
	if t.MaxAllocs > 0 {
		scr.SetMaxAllocs(int64(t.MaxAllocs))
	}
	if t.MaxConstObjects > 0 {
		scr.SetMaxConstObjects(t.MaxConstObjects)
	}

	return scr, nil
}

// interface guards
var (
	_ caddy.Provisioner           = (*Tengo)(nil)
	_ caddyfile.Unmarshaler       = (*Tengo)(nil)
	_ caddyhttp.MiddlewareHandler = (*Tengo)(nil)
	_ caddy.Validator             = (*Tengo)(nil)
)
