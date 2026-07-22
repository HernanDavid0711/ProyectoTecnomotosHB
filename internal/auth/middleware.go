package auth

import (
	"context"
	"net/http"

	"tecnomotos/internal/shared"
)

type contextKey string

const claimsContextKey contextKey = "auth_claims"

func Authenticate(secret string, issuer string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, err := BearerToken(r.Header.Get("Authorization"))
			if err != nil {
				shared.WriteError(w, http.StatusUnauthorized, "no_autenticado", "token bearer requerido")
				return
			}

			claims, err := ParseToken(token, secret, issuer)
			if err != nil {
				shared.WriteError(w, http.StatusUnauthorized, "token_invalido", "token invalido o expirado")
				return
			}

			ctx := context.WithValue(r.Context(), claimsContextKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequireRole(roles ...string) func(http.Handler) http.Handler {
	allowed := make(map[string]bool, len(roles))
	for _, role := range roles {
		allowed[role] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := ClaimsFromContext(r.Context())
			if !ok {
				shared.WriteError(w, http.StatusUnauthorized, "no_autenticado", "token bearer requerido")
				return
			}

			if !allowed[claims.Rol] {
				shared.WriteError(w, http.StatusForbidden, "sin_permiso", "no tienes permisos para realizar esta accion")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func ClaimsFromContext(ctx context.Context) (*Claims, bool) {
	claims, ok := ctx.Value(claimsContextKey).(*Claims)
	return claims, ok
}

func Chain(handler http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}
	return handler
}
