// Package http provides HTTP-related functionality for the application.
package http

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/muhadif/sprt/domain/usecase"
)

// CallbackServer represents an HTTP server that handles Spotify OAuth callbacks.
type CallbackServer struct {
	server      *http.Server
	authUseCase usecase.AuthUseCase
}

// NewCallbackServer creates a new instance of CallbackServer.
func NewCallbackServer(authUseCase usecase.AuthUseCase) *CallbackServer {
	return &CallbackServer{
		authUseCase: authUseCase,
	}
}

// Start starts the callback server on the specified port.
func (s *CallbackServer) Start(port int) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/callback", s.handleCallback)

	s.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	fmt.Printf("Callback server started on http://localhost:%d\n", port)
	fmt.Println("Waiting for Spotify authorization...")

	return s.server.ListenAndServe()
}

// Stop stops the callback server.
func (s *CallbackServer) Stop(ctx context.Context) error {
	if s.server == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return s.server.Shutdown(ctx)
}

// handleCallback handles the callback from Spotify with the authorization code.
func (s *CallbackServer) handleCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Authorization code not found", http.StatusBadRequest)
		return
	}

	err := s.authUseCase.HandleCallback(r.Context(), code)
	if err != nil {
		log.Printf("Error handling callback: %v", err)
		http.Error(w, "Error handling callback", http.StatusInternalServerError)
		return
	}

	// Exchange the code for a token
	err = s.authUseCase.ExchangeCodeForToken(r.Context())
	if err != nil {
		log.Printf("Error exchanging code for token: %v", err)
		http.Error(w, "Error exchanging code for token", http.StatusInternalServerError)
		return
	}

	// Return a success message to the user
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`
		<html>
			<body>
				<h1>Authentication Successful</h1>
				<p>You have successfully authenticated with Spotify. You can now close this window and return to the CLI.</p>
			</body>
		</html>
	`))

	// Log success message
	log.Println("Authentication successful. Token received and stored.")
}
