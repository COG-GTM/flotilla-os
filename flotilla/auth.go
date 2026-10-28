package flotilla

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"

	"github.com/stitchfix/flotilla-os/config"
	"github.com/stitchfix/flotilla-os/exceptions"
	"github.com/stitchfix/flotilla-os/state"
)

// Role is a permission granted to an API key.
type Role string

const (
	// RoleUser can read state, manage definitions and templates, and launch or stop runs.
	RoleUser Role = "user"
	// RoleOperator can manage clusters and workers.
	RoleOperator Role = "operator"
	// RoleWorker is the only role that can report run status (PUT /{run_id}/status).
	RoleWorker Role = "worker"
)

var knownRoles = map[Role]bool{RoleUser: true, RoleOperator: true, RoleWorker: true}

// Principal is the verified identity behind a request.
type Principal struct {
	Identity string
	Roles    map[Role]bool
	// Unverified is set only when authentication is explicitly disabled.
	Unverified bool
}

// HasAnyRole reports whether the principal holds at least one of the roles.
func (p Principal) HasAnyRole(roles ...Role) bool {
	for _, r := range roles {
		if p.Roles[r] {
			return true
		}
	}
	return false
}

type principalKey struct{}

// PrincipalFromContext returns the principal attached by the authenticator.
func PrincipalFromContext(ctx context.Context) (Principal, bool) {
	p, ok := ctx.Value(principalKey{}).(Principal)
	return p, ok
}

// WithPrincipal attaches a principal to a context.
func WithPrincipal(ctx context.Context, p Principal) context.Context {
	return context.WithValue(ctx, principalKey{}, p)
}

// Authenticator validates bearer API keys against SHA-256 digests from config.
type Authenticator struct {
	disabled bool
	keys     map[string]Principal
}

// NewAuthenticator builds an Authenticator from `http_server_api_keys`.
//
// Each entry has the form `<identity>:<role>[+<role>...]:<hex sha256 of the key>`.
// Entries may be given as a YAML list or as a comma separated string
// (e.g. the HTTP_SERVER_API_KEYS environment variable). Authentication can only
// be turned off explicitly with `http_server_auth_disabled: true`.
func NewAuthenticator(conf config.Config) (*Authenticator, error) {
	if conf.GetBool("http_server_auth_disabled") {
		return &Authenticator{disabled: true}, nil
	}
	keys := make(map[string]Principal)
	for _, entry := range configList(conf, "http_server_api_keys") {
		parts := strings.Split(entry, ":")
		if len(parts) != 3 {
			return nil, fmt.Errorf("invalid http_server_api_keys entry for [%s]: expected <identity>:<roles>:<sha256>", parts[0])
		}
		identity := strings.TrimSpace(parts[0])
		digest := strings.ToLower(strings.TrimSpace(parts[2]))
		if identity == "" {
			return nil, fmt.Errorf("invalid http_server_api_keys entry: empty identity")
		}
		if b, err := hex.DecodeString(digest); err != nil || len(b) != sha256.Size {
			return nil, fmt.Errorf("invalid http_server_api_keys entry for [%s]: key digest must be 64 hex characters", identity)
		}
		roles := make(map[Role]bool)
		for _, r := range strings.Split(parts[1], "+") {
			role := Role(strings.TrimSpace(r))
			if !knownRoles[role] {
				return nil, fmt.Errorf("invalid http_server_api_keys entry for [%s]: unknown role [%s]", identity, role)
			}
			roles[role] = true
		}
		if _, dup := keys[digest]; dup {
			return nil, fmt.Errorf("invalid http_server_api_keys entry for [%s]: duplicate key digest", identity)
		}
		keys[digest] = Principal{Identity: identity, Roles: roles}
	}
	if len(keys) == 0 {
		return nil, fmt.Errorf("no API keys configured: set http_server_api_keys (or http_server_auth_disabled: true for local development only)")
	}
	return &Authenticator{keys: keys}, nil
}

// Disabled reports whether authentication was explicitly turned off.
func (a *Authenticator) Disabled() bool {
	return a != nil && a.disabled
}

func (a *Authenticator) authenticate(r *http.Request) (Principal, error) {
	if a == nil {
		return Principal{}, exceptions.Unauthenticated{ErrorString: "authentication is not configured"}
	}
	if a.disabled {
		return Principal{Identity: "anonymous", Unverified: true, Roles: map[Role]bool{RoleUser: true, RoleOperator: true, RoleWorker: true}}, nil
	}
	header := r.Header.Get("Authorization")
	scheme, token, found := strings.Cut(header, " ")
	token = strings.TrimSpace(token)
	if !found || !strings.EqualFold(scheme, "Bearer") || token == "" {
		return Principal{}, exceptions.Unauthenticated{ErrorString: "missing bearer token"}
	}
	sum := sha256.Sum256([]byte(token))
	p, ok := a.keys[hex.EncodeToString(sum[:])]
	if !ok {
		return Principal{}, exceptions.Unauthenticated{ErrorString: "invalid bearer token"}
	}
	return p, nil
}

// Require wraps a handler so it only runs for callers holding one of the roles.
func (a *Authenticator) Require(next http.HandlerFunc, roles ...Role) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		p, err := a.authenticate(r)
		if err != nil {
			w.Header().Set("WWW-Authenticate", `Bearer realm="flotilla"`)
			writeAuthError(w, http.StatusUnauthorized, err)
			return
		}
		if !p.HasAnyRole(roles...) {
			writeAuthError(w, http.StatusForbidden, exceptions.Forbidden{
				ErrorString: fmt.Sprintf("caller [%s] lacks a required role %v", p.Identity, roles)})
			return
		}
		next(w, r.WithContext(WithPrincipal(r.Context(), p)))
	}
}

func writeAuthError(w http.ResponseWriter, status int, err error) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_, _ = fmt.Fprintf(w, "{\"error\":%q}\n", err.Error())
}

// LaunchPolicy holds operator-controlled limits on what callers may launch.
type LaunchPolicy struct {
	serviceAccounts map[string]bool
	imageRegistries []string
}

// NewLaunchPolicy builds the policy from config. Only the default EKS/EMR job
// service accounts plus `allowed_service_accounts` may be requested. When
// `allowed_image_registries` is set, images must come from one of them.
func NewLaunchPolicy(conf config.Config) LaunchPolicy {
	p := LaunchPolicy{serviceAccounts: make(map[string]bool)}
	for _, key := range []string{"eks_default_service_account", "emr_default_service_account"} {
		if sa := strings.TrimSpace(conf.GetString(key)); sa != "" {
			p.serviceAccounts[sa] = true
		}
	}
	for _, sa := range configList(conf, "allowed_service_accounts") {
		p.serviceAccounts[sa] = true
	}
	for _, reg := range configList(conf, "allowed_image_registries") {
		if reg = strings.TrimRight(reg, "/"); reg != "" {
			p.imageRegistries = append(p.imageRegistries, reg)
		}
	}
	return p
}

// CheckServiceAccount rejects service accounts that are not allowlisted. An
// empty value is normalised to nil so the engine's default account is used.
func (p LaunchPolicy) CheckServiceAccount(common *state.ExecutionRequestCommon) error {
	if common == nil || common.ServiceAccount == nil {
		return nil
	}
	sa := strings.TrimSpace(*common.ServiceAccount)
	if sa == "" {
		common.ServiceAccount = nil
		return nil
	}
	if !p.serviceAccounts[sa] {
		return exceptions.MalformedInput{ErrorString: fmt.Sprintf("service_account [%s] is not allowed", sa)}
	}
	common.ServiceAccount = &sa
	return nil
}

// CheckImage rejects images outside the configured registries.
func (p LaunchPolicy) CheckImage(image string) error {
	if len(p.imageRegistries) == 0 || image == "" {
		return nil
	}
	for _, reg := range p.imageRegistries {
		if strings.HasPrefix(image, reg+"/") {
			return nil
		}
	}
	return exceptions.MalformedInput{ErrorString: fmt.Sprintf("image [%s] is not from an allowed registry", image)}
}

func configList(conf config.Config, key string) []string {
	var out []string
	for _, item := range conf.GetStringSlice(key) {
		for _, v := range strings.Split(item, ",") {
			if v = strings.TrimSpace(v); v != "" {
				out = append(out, v)
			}
		}
	}
	return out
}
