package middleware

import (
	"net/http"

	"github.com/netbirdio/netbird/management/server/http/util"
	"github.com/netbirdio/netbird/management/server/status"

	"github.com/netbirdio/netbird/management/server/jwtclaims"
)

type IsUserAdminFunc func(claims jwtclaims.AuthorizationClaims) (bool, error)

// AccessControl middleware to restrict to make POST/PUT/DELETE requests by admin only
type AccessControl struct {
	isUserAdmin   IsUserAdminFunc
	claimsExtract jwtclaims.ClaimsExtractor
}

// NewAccessControl instance constructor
func NewAccessControl(audience, userIDClaim string, isUserAdmin IsUserAdminFunc) *AccessControl {
	return &AccessControl{
		isUserAdmin: isUserAdmin,
		claimsExtract: *jwtclaims.NewClaimsExtractor(
			jwtclaims.WithAudience(audience),
			jwtclaims.WithUserIDClaim(userIDClaim),
		),
	}
}

// Handler method of the middleware which forbids all modify requests for non admin users
// It also adds
func (a *AccessControl) Handler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims := a.claimsExtract.FromRequestContext(r)

		ok, err := a.isUserAdmin(claims)
		if err != nil {
			util.WriteError(status.Errorf(status.Unauthorized, "invalid JWT"), w)
			return
		}

		if !ok {
			switch r.Method {
			case http.MethodDelete, http.MethodPost, http.MethodPatch, http.MethodPut:
				util.WriteError(status.Errorf(status.PermissionDenied, "only admin can perform this operation"), w)
				return
			}
		}

		h.ServeHTTP(w, r)
	})
}
