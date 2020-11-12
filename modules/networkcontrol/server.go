package networkcontrol

import "net/http"

func (m *Manager) createMux() *http.ServeMux {
	mux := http.NewServeMux()
	return mux
}

func (m *Manager) StartServer(address string) {
	mux := m.createMux()
	srv := &http.Server{
		Addr:    address,
		Handler: mux,
	}
	m.server = srv
	go func() {
		packageLogger.WithField("address", address).Debug("server starting")
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			packageLogger.Errorf("server stopped unexpectedly: %+v", err)
		} else {
			packageLogger.Debug("server shut down")
		}
	}()
}

func (m *Manager) StopServer() error {
	return nil
}
