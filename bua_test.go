package bua

import (
	"testing"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr bool
	}{
		{
			name: "valid config",
			cfg: Config{
				APIKey: "test-key",
				Model:  "gemini-2.5-flash",
			},
			wantErr: false,
		},
		{
			name: "missing API key",
			cfg: Config{
				Model: "gemini-2.5-flash",
			},
			wantErr: true,
		},
		{
			name: "with profile name",
			cfg: Config{
				APIKey:      "test-key",
				ProfileName: "test-profile",
			},
			wantErr: false,
		},
		{
			name: "with custom viewport",
			cfg: Config{
				APIKey:   "test-key",
				Viewport: &Viewport{Width: 1280, Height: 720},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent, err := New(tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && agent == nil {
				t.Error("New() returned nil agent without error")
			}
		})
	}
}

func TestViewportDefaults(t *testing.T) {
	if DesktopViewport == nil {
		t.Error("DesktopViewport is nil")
	}
	if DesktopViewport.Width != 1920 || DesktopViewport.Height != 1080 {
		t.Errorf("DesktopViewport = %v, want 1920x1080", DesktopViewport)
	}

	if TabletViewport == nil {
		t.Error("TabletViewport is nil")
	}
	if TabletViewport.Width != 768 || TabletViewport.Height != 1024 {
		t.Errorf("TabletViewport = %v, want 768x1024", TabletViewport)
	}

	if MobileViewport == nil {
		t.Error("MobileViewport is nil")
	}
	if MobileViewport.Width != 375 || MobileViewport.Height != 812 {
		t.Errorf("MobileViewport = %v, want 375x812", MobileViewport)
	}
}

func TestConfigDefaults(t *testing.T) {
	cfg := Config{
		APIKey: "test-key",
	}

	agent, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	// Check that defaults were applied
	if agent.config.Model == "" {
		t.Error("Default model was not set")
	}
	if agent.config.MaxTokens == 0 {
		t.Error("Default MaxTokens was not set")
	}
	if agent.config.Viewport == nil {
		t.Error("Default Viewport was not set")
	}
}
