package service

import (
	"context"
	"github.com/mkrill/gosea/src/seabackend/domain/entity"
)

type (
	ISeaBackendService interface{
		LoadPosts(ctx context.Context) ([]entity.RemotePost, error)
		LoadUsers(ctx context.Context) ([]entity.RemoteUser, error)
		LoadUser(ctx context.Context, id string) (entity.RemoteUser, error)
	}
)