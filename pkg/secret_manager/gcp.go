package secret_manager

import (
	"context"
	"fmt"
	"os"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"github.com/avareum/avareum-hubble-signer/pkg/secret_manager/types"
	secretmanagerpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1"
)

type GCPSecretManager struct {
	types.SecretManager
	client *secretmanager.Client
	cfg    GCPSecretManagerConfig
}

type GCPSecretManagerConfig struct {
	ProjectID string
}

func NewGCPSecretManager() (types.SecretManager, error) {
	sm := &GCPSecretManager{
		cfg: GCPSecretManagerConfig{
			ProjectID: os.Getenv("GCP_PROJECT"),
		},
	}
	err := sm.init()
	if err != nil {
		return nil, err
	}
	return sm, nil
}

func (s *GCPSecretManager) init() error {
	client, err := secretmanager.NewClient(context.TODO())
	if err != nil {
		return fmt.Errorf("GCPSecretManager: new client failed: %v", err)
	}
	s.client = client
	return nil
}

func (s *GCPSecretManager) Create(id string, payload []byte) (string, error) {
	createSecretReq := &secretmanagerpb.CreateSecretRequest{
		Parent:   fmt.Sprintf("projects/signer-%s", s.cfg.ProjectID),
		SecretId: id,
		Secret: &secretmanagerpb.Secret{
			Replication: &secretmanagerpb.Replication{
				Replication: &secretmanagerpb.Replication_Automatic_{
					Automatic: &secretmanagerpb.Replication_Automatic{},
				},
			},
		},
	}

	secret, err := s.client.CreateSecret(context.TODO(), createSecretReq)
	if err != nil {
		return "", fmt.Errorf("GCPSecretManager: failed to create secret: %v", err)
	}

	// Build the request.
	addSecretVersionReq := &secretmanagerpb.AddSecretVersionRequest{
		Parent: secret.Name,
		Payload: &secretmanagerpb.SecretPayload{
			Data: payload,
		},
	}

	// Call the API.
	version, err := s.client.AddSecretVersion(context.TODO(), addSecretVersionReq)
	if err != nil {
		return "", fmt.Errorf("GCPSecretManager: failed to add secret version: %v", err)
	}

	return version.Name, nil
}

func (s *GCPSecretManager) Get(id string) ([]byte, error) {
	// Build the request.
	accessRequest := &secretmanagerpb.AccessSecretVersionRequest{
		Name: fmt.Sprintf("projects/%s/secrets/signer-%s/versions/latest", s.cfg.ProjectID, id),
	}

	// Call the API.
	result, err := s.client.AccessSecretVersion(context.TODO(), accessRequest)
	if err != nil {
		return nil, fmt.Errorf("GCPSecretManager: failed to access secret version: %v", err)
	}

	return result.Payload.Data, nil
}