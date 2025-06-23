package sync

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/kirinyoku/vlxck/internal/config"
	"github.com/kirinyoku/vlxck/internal/utils"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

// GoogleDriveSync implements synchronization with Google Drive
type GoogleDriveSync struct {
	client *drive.Service
	config *config.Config
	token  *oauth2.Token
	ctx    context.Context
}

// validateClientIDFormat validates the format of the Client ID
func validateClientIDFormat(clientID string) error {
	pattern := `^[0-9]+-[a-z0-9]+\.apps\.googleusercontent\.com$`
	re, err := regexp.Compile(pattern)
	if err != nil {
		return fmt.Errorf("failed to compile regex: %v", err)
	}
	if !re.MatchString(clientID) {
		return fmt.Errorf("invalid Client ID format; it should look like '123456789012-abcdefghi.apps.googleusercontent.com'")
	}
	return nil
}

// validateCredentials validates the Client ID and Client Secret
func validateCredentials(clientID, clientSecret string) error {
	oauthConfig := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  "http://localhost:8080/oauth2callback",
		Scopes:       []string{drive.DriveFileScope},
		Endpoint:     google.Endpoint,
	}

	// Attempt to exchange an invalid code to check for invalid_client error
	_, err := oauthConfig.Exchange(context.Background(), "invalid-code")
	if err != nil && strings.Contains(err.Error(), "invalid_client") {
		return fmt.Errorf("invalid Client ID or Client Secret: %v", err)
	}

	// Note: A successful exchange won't happen, but we expect a different error (e.g., invalid_grant)
	return nil
}

// getClientCredentials retrieves or prompts for ClientID and ClientSecret
func getClientCredentials(cfg *config.Config, masterPassword string) (string, string, error) {
	if len(cfg.Sync.EncryptedClientId) > 0 && len(cfg.Sync.EncryptedClientSecret) > 0 {
		clientID, clientSecret, err := config.DecryptClientCredentials(cfg.Sync.EncryptedClientId, cfg.Sync.EncryptedClientSecret, masterPassword)
		if err != nil {
			return "", "", fmt.Errorf("failed to decrypt client credentials: %v", err)
		}
		// Validate stored credentials
		if err := validateClientIDFormat(clientID); err != nil {
			return "", "", err
		}
		if err := validateCredentials(clientID, clientSecret); err != nil {
			return "", "", err
		}
		return clientID, clientSecret, nil
	}

	fmt.Println("Google Drive API credentials are required for synchronization.")
	fmt.Println("Please follow the instructions in https://github.com/kirinyoku/vlxck/blob/main/README.md#configuring-google-drive-sync to obtain your Client ID and Client Secret.")

	for {
		clientID, err := utils.PromptForInput("Enter Client ID: ", "", func(input string) error {
			return validateClientIDFormat(input)
		})
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			fmt.Println("Please re-enter your credentials.")
			continue
		}

		clientSecret, err := utils.PromptForInput("Enter Client Secret: ", "", func(input string) error {
			if input == "" {
				return fmt.Errorf("client secret cannot be empty")
			}
			return nil
		})
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			fmt.Println("Please re-enter your credentials.")
			continue
		}

		return clientID, clientSecret, nil
	}
}

// NewGoogleDriveSync creates a new synchronizer instance
func NewGoogleDriveSync(ctx context.Context, masterPassword string) (*GoogleDriveSync, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %v", err)
	}

	if cfg.Sync.Provider != "google_drive" && cfg.Sync.Provider != "" {
		return nil, fmt.Errorf("google_drive provider not configured")
	}

	token, err := config.DecryptToken(cfg.Sync.EncryptedToken, masterPassword)
	if err != nil && len(cfg.Sync.EncryptedToken) > 0 {
		return nil, fmt.Errorf("failed to decrypt token: %v", err)
	}

	clientID, clientSecret, err := getClientCredentials(cfg, masterPassword)
	if err != nil {
		return nil, fmt.Errorf("failed to get client credentials: %v", err)
	}

	oauthConfig := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  "http://localhost:8080/oauth2callback",
		Scopes:       []string{drive.DriveFileScope},
		Endpoint:     google.Endpoint,
	}

	var client *http.Client
	if token != nil {
		client = oauthConfig.Client(ctx, token)
	} else {
		client = &http.Client{}
	}

	driveService, err := drive.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("failed to create drive service: %v", err)
	}

	return &GoogleDriveSync{
		client: driveService,
		config: cfg,
		token:  token,
		ctx:    ctx,
	}, nil
}

// InitGoogleDrive initializes synchronization
func InitGoogleDrive(ctx context.Context, masterPassword string) (*config.Config, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %v", err)
	}

	clientID, clientSecret, err := getClientCredentials(cfg, masterPassword)
	if err != nil {
		return nil, fmt.Errorf("failed to get client credentials: %v", err)
	}

	oauthConfig := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  "http://localhost:8080/oauth2callback",
		Scopes:       []string{drive.DriveFileScope},
		Endpoint:     google.Endpoint,
	}

	url := oauthConfig.AuthCodeURL("state", oauth2.AccessTypeOffline)
	fmt.Printf("Please visit the following URL to authorize:\n%s\n", url)

	tokenChan := make(chan *oauth2.Token)
	errChan := make(chan error)
	http.HandleFunc("/oauth2callback", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code == "" {
			err := fmt.Errorf("authorization code missing")
			errChan <- err
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		token, err := oauthConfig.Exchange(ctx, code)
		if err != nil {
			if strings.Contains(err.Error(), "invalid_client") {
				err = fmt.Errorf("invalid Client ID or Client Secret: %v", err)
				// Clear credentials to force re-entry
				cfg.Sync.EncryptedClientId = nil
				cfg.Sync.EncryptedClientSecret = nil
				if saveErr := config.SaveConfig(cfg); saveErr != nil {
					fmt.Fprintf(os.Stderr, "Failed to clear invalid credentials: %v\n", saveErr)
				}
			}
			errChan <- err
			http.Error(w, fmt.Sprintf("Failed to exchange token: %v", err), http.StatusInternalServerError)
			return
		}
		tokenChan <- token
		fmt.Fprintf(w, "Authorization successful! You can close this window.")
	})

	server := &http.Server{Addr: ":8080"}
	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "OAuth server error: %v\n", err)
		}
	}()

	select {
	case token := <-tokenChan:
		// Save credentials only after successful authorization
		encryptedClientID, encryptedClientSecret, err := config.EncryptClientCredentials(clientID, clientSecret, masterPassword)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt client credentials: %v", err)
		}
		cfg.Sync.EncryptedClientId = encryptedClientID
		cfg.Sync.EncryptedClientSecret = encryptedClientSecret

		encryptedToken, err := config.EncryptToken(token, masterPassword)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt token: %v", err)
		}

		cfg.Sync.Provider = "google_drive"
		cfg.Sync.EncryptedToken = encryptedToken

		client := oauthConfig.Client(ctx, token)
		driveService, err := drive.NewService(ctx, option.WithHTTPClient(client))
		if err != nil {
			return nil, fmt.Errorf("failed to create drive service: %v", err)
		}

		file := &drive.File{
			Name:    "store.dat",
			Parents: []string{"root"},
		}
		createdFile, err := driveService.Files.Create(file).Media(bytes.NewReader([]byte{})).Do()
		if err != nil {
			return nil, fmt.Errorf("failed to create file: %v", err)
		}
		cfg.Sync.FileID = createdFile.Id

		fmt.Printf("Saving config with FileID: %s, Provider: %s\n", cfg.Sync.FileID, cfg.Sync.Provider)
		if err := config.SaveConfig(cfg); err != nil {
			return nil, fmt.Errorf("failed to save config: %v", err)
		}

		if err := server.Shutdown(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to shutdown OAuth server: %v\n", err)
		}
		return cfg, nil
	case err := <-errChan:
		if err := server.Shutdown(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to shutdown OAuth server: %v\n", err)
		}
		return nil, err
	}
}

// Push uploads the local store.dat to Google Drive
func (g *GoogleDriveSync) Push(storePath string) error {
	data, err := os.Open(storePath)
	if err != nil {
		return fmt.Errorf("failed to read store: %v", err)
	}
	defer data.Close()

	_, err = g.client.Files.Update(g.config.Sync.FileID, nil).Media(data).Do()
	if err != nil {
		return fmt.Errorf("failed to upload file: %v", err)
	}

	file, err := g.client.Files.Get(g.config.Sync.FileID).Fields("id, modifiedTime, md5Checksum").Do()
	if err != nil {
		return fmt.Errorf("failed to get file metadata: %v", err)
	}
	g.config.Sync.Etag = file.Md5Checksum
	if err := config.SaveConfig(g.config); err != nil {
		return fmt.Errorf("failed to save config: %v", err)
	}

	return nil
}

// Pull downloads store.dat from Google Drive
func (g *GoogleDriveSync) Pull(storePath string) error {
	resp, err := g.client.Files.Get(g.config.Sync.FileID).Download()
	if err != nil {
		return fmt.Errorf("failed to download file: %v", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read file: %v", err)
	}

	if err := os.MkdirAll(filepath.Dir(storePath), 0755); err != nil {
		return fmt.Errorf("failed to create store directory: %v", err)
	}

	if err := os.WriteFile(storePath, data, 0600); err != nil {
		return fmt.Errorf("failed to write store: %v", err)
	}

	file, err := g.client.Files.Get(g.config.Sync.FileID).Fields("id, modifiedTime, md5Checksum").Do()
	if err != nil {
		return fmt.Errorf("failed to get file metadata: %v", err)
	}
	g.config.Sync.Etag = file.Md5Checksum
	if err := config.SaveConfig(g.config); err != nil {
		return fmt.Errorf("failed to save config: %v", err)
	}

	return nil
}

// GetMetadata returns file metadata
func (g *GoogleDriveSync) GetMetadata() (string, time.Time, error) {
	file, err := g.client.Files.Get(g.config.Sync.FileID).Fields("id, modifiedTime, md5Checksum").Do()
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to get file metadata: %v", err)
	}
	modifiedTime, err := time.Parse(time.RFC3339, file.ModifiedTime)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to parse modified time: %v", err)
	}
	return file.Md5Checksum, modifiedTime, nil
}
