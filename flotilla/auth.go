package flotilla

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/stitchfix/flotilla-os/config"
)

// Role is the permission level granted to an authenticated API caller.
type Role string

const (
	// RoleUser may manage definitions/templates and launch, view and stop runs.
	RoleUser Role = "user"
	// RoleOperator may additionally manage clusters and workers and update run status.
	RoleOperator Role = "operator"

	authDisabledConfKey = "http_server_auth_disabled"
	apiKeysConfKey      = "http_server_api_keys"
	healthCheckPath     = "/api/v6/health"
)

// Caller is the verified identity of the client making an API request.
type Caller struct {
	Identity string
	Role     Role
	// Anonymous is true only when authentication is explicitly disabled.
	Anonymous bool
}

type callerContextKey struct{}

func callerFromContext(ctx context.Context) (Caller, bool) {
	c, ok := ctx.Value(callerContextKey{}).(Caller)
	return c, ok
}

type apiKey struct {
	hash     [sha256.Size]byte
	identity string
	role     Role
}

// Authenticator validates bearer API keys on every request.
type Authenticator struct {
	disabled bool
	keys     []apiKey
}

// NewAuthenticator builds an Authenticator from configuration.
//
// `http_server_api_keys` (env HTTP_SERVER_API_KEYS) is a comma separated list of
// `<identity>:<role>:<token>` entries where role is `user` or `operator`.
// Authentication can only be turned off by setting `http_server_auth_disabled`
// to true; with no keys configured and auth enabled, startup fails.
func NewAuthenticator(conf config.Config) (*Authenticator, error) {
	if conf.GetBool(authDisabledConfKey) {
		return &Authenticator{disabled: true}, nil
	}
	var entries []string
	for _, v := range conf.GetStringSlice(apiKeysConfKey) {
		entries = append(entries, strings.Split(v, ",")...)
	}
	return newAuthenticatorFromEntries(entries)
}

func newAuthenticatorFromEntries(entries []string) (*Authenticator, error) {
	a := &Authenticator{}
	for _, entry := range entries {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		parts := strings.SplitN(entry, ":", 3)
		if len(parts) != 3 || parts[0] == "" || parts[2] == "" {
			return nil, fmt.Errorf("invalid [%s] entry: expected <identity>:<role>:<token>", apiKeysConfKey)
		}
		role := Role(parts[1])
		if role != RoleUser && role != RoleOperator {
			return nil, fmt.Errorf("invalid role [%s] for identity [%s] in [%s]", parts[1], parts[0], apiKeysConfKey)
		}
		a.keys = append(a.keys, apiKey{hash: sha256.Sum256([]byte(parts[2])), identity: parts[0], role: role})
	}
	if len(a.keys) == 0 {
		return nil, fmt.Errorf("no API keys configured: set [%s] or explicitly set [%s] to true", apiKeysConfKey, authDisabledConfKey)
	}
	return a, nil
}

func (a *Authenticator) authenticate(r *http.Request) (Caller, bool) {
	if a == nil {
		return Caller{}, false
	}
	if a.disabled {
		return Caller{Identity: "anonymous", Role: RoleOperator, Anonymous: true}, true
	}
	h := r.Header.Get("Authorization")
	const prefix = "bearer "
	if len(h) <= len(prefix) || !strings.EqualFold(h[:len(prefix)], prefix) {
		return Caller{}, false
	}
	presented := sha256.Sum256([]byte(strings.TrimSpace(h[len(prefix):])))
	var (
		match Caller
		found bool
	)
	for _, k := range a.keys {
		if subtle.ConstantTimeCompare(presented[:], k.hash[:]) == 1 && !found {
			match = Caller{Identity: k.identity, Role: k.role}
			found = true
		}
	}
	return match, found
}

// Middleware rejects unauthenticated requests and stores the caller in the request context.
func (a *Authenticator) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == healthCheckPath {
			next.ServeHTTP(w, r)
			return
		}
		caller, ok := a.authenticate(r)
		if !ok {
			w.Header().Set("WWW-Authenticate", `Bearer realm="flotilla"`)
			writeAuthError(w, http.StatusUnauthorized, "missing or invalid API key")
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), callerContextKey{}, caller)))
	})
}

// requireOperator restricts a handler to callers with the operator role.
func requireOperator(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		caller, ok := callerFromContext(r.Context())
		if !ok || caller.Role != RoleOperator {
			writeAuthError(w, http.StatusForbidden, "operator role required")
			return
		}
		next(w, r)
	}
}

func writeAuthError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": msg})
}

// resolveOwnerID returns the owner to record for a run. Authenticated users are
// always attributed to their verified identity; operators (e.g. schedulers) may
// launch on behalf of another owner.
func resolveOwnerID(r *http.Request, requested string) string {
	caller, ok := callerFromContext(r.Context())
	if !ok || caller.Anonymous {
		return requested
	}
	if caller.Role == RoleOperator && len(requested) > 0 {
		return requested
	}
	return caller.Identity
}
