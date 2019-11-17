package main

type Path struct {
	ProxyPath       string
	ServiceName     string
	ActualPath      string
	Method          string
	AuthWhitelisted bool
}

const (
	// list of services - sorted by name
	AUTH_SERVICE               = "auth-go"
	CARI_RUMAH_BACKEND_SERVICE = "cari-rumah-backend"
	EMAIL_SERVICE              = "email-service"

	// list of port numbers - sorted by port number
	CARI_RUMAH_BACKEND_SERVICE_PORT = 3000
	AUTH_SERVICE_PORT               = 3000
	EMAIL_SERVICE_PORT              = 3000
)

// sorted alphabetically by name
var ServiceToDnsResolver = map[string]int{
	AUTH_SERVICE:               AUTH_SERVICE_PORT,
	CARI_RUMAH_BACKEND_SERVICE: CARI_RUMAH_BACKEND_SERVICE_PORT,
	EMAIL_SERVICE:              EMAIL_SERVICE_PORT,
}

// sorted by proxy path
var Paths = []Path{
	Path{"/auth/api/v1/credentials", AUTH_SERVICE, "/credentials", "POST", false},
	Path{"/auth/api/v1/credentials/initiate_password_reset", AUTH_SERVICE, "/credentials/initiate_password_reset", "POST", true},
	Path{"/auth/api/v1/credentials/reset_password", AUTH_SERVICE, "/credentials/reset_password", "POST", true},
	Path{"/auth/api/v1/login", AUTH_SERVICE, "/login", "POST", true},
	Path{"/auth/api/v1/token", AUTH_SERVICE, "/token", "POST", true},
	Path{"/auth/api/v1/verify", AUTH_SERVICE, "/verify", "POST", true},
	Path{"/cari-rumah-backend/graphql", CARI_RUMAH_BACKEND_SERVICE, "/graphql", "POST", false},
	Path{"/cari-rumah-backend/google/autocomplete", CARI_RUMAH_BACKEND_SERVICE, "/google/autocomplete", "GET", false},
	Path{"/cari-rumah-backend/google/placeGeometry", CARI_RUMAH_BACKEND_SERVICE, "/google/placeGeometry", "GET", false},
	Path{"/email/api/v1/send", EMAIL_SERVICE, "/api/v1/send", "POST", false},
}
