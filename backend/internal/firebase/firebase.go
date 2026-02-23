package firebase

import (
	"context"
	"fmt"

	firebase "firebase.google.com/go/v4"
	"github.com/erp-sppg/backend/internal/config"
	"google.golang.org/api/option"
)

func Initialize(cfg *config.Config) (*firebase.App, error) {
	ctx := context.Background()

	opt := option.WithCredentialsFile(cfg.FirebaseCredentialsPath)
	
	firebaseConfig := &firebase.Config{
		DatabaseURL:   cfg.FirebaseDatabaseURL,
		StorageBucket: cfg.StorageBucket,
	}

	app, err := firebase.NewApp(ctx, firebaseConfig, opt)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Firebase: %w", err)
	}

	return app, nil
}
