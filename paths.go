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
	AUTH_SERVICE      = "auth-go"
	EMAIL_SERVICE     = "email-service"
	PROJECT_K_BACKEND = "project-k-backend"

	// list of port numbers - sorted by port number
	AUTH_SERVICE_PORT      = 3000
	EMAIL_SERVICE_PORT     = 3000
	PROJECT_K_BACKEND_PORT = 3000
)

// sorted alphabetically by name
var ServiceToDnsResolver = map[string]int{
	AUTH_SERVICE:      AUTH_SERVICE_PORT,
	EMAIL_SERVICE:     EMAIL_SERVICE_PORT,
	PROJECT_K_BACKEND: PROJECT_K_BACKEND_PORT,
}

// sorted by proxy path
var Paths = []Path{
	Path{"/auth/api/v1/credentials", AUTH_SERVICE, "/credentials", "POST", true},
	Path{"/auth/api/v1/credentials/initiate_password_reset", AUTH_SERVICE, "/credentials/initiate_password_reset", "POST", true},
	Path{"/auth/api/v1/credentials/reset_password", AUTH_SERVICE, "/credentials/reset_password", "POST", true},
	Path{"/auth/api/v1/login", AUTH_SERVICE, "/login", "POST", true},
	Path{"/auth/api/v1/token", AUTH_SERVICE, "/token", "POST", true},
	Path{"/auth/api/v1/verify", AUTH_SERVICE, "/verify", "POST", true},
	Path{"/auth/api/v1/verify_session_token", AUTH_SERVICE, "/verify_session_token", "POST", true},
	Path{"/email/api/v1/send", EMAIL_SERVICE, "/api/v1/send", "POST", false},

	// Project K Backend is a special case where we want to just pass through API calls
	// wildcards need to be named but because we don't need it just use 'x' as varName
	Path{"/backend/*x", PROJECT_K_BACKEND, "*", "GET", false},
	Path{"/backend/*x", PROJECT_K_BACKEND, "*", "DELETE", false},
	Path{"/backend/*x", PROJECT_K_BACKEND, "*", "PUT", false},
	Path{"/backend/*x", PROJECT_K_BACKEND, "*", "POST", false},
	Path{"/backend/*x", PROJECT_K_BACKEND, "*", "OPTIONS", false},
}
