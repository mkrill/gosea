package Service

import (
	"context"
	"github.com/mkrill/gosea/src/seaBackend/domain/Entity"
)

type (
	ISeaBackendService interface{
		LoadPosts(ctx context.Context) ([]Entity.RemotePost, error)
		LoadUsers(ctx context.Context) ([]Entity.RemoteUser, error)
		LoadUser(ctx context.Context, id string) (Entity.RemoteUser, error)
	}
)