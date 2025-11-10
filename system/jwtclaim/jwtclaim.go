package jwtclaim

const (
	JwtClaimDataKey = "JwtClaimDataKey"
)

type JwtClaimData struct {
	UserId       uint64
	Username     string
	IsSuperAdmin bool
}
