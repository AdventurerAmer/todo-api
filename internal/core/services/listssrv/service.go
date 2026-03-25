package listssrv

import (
	"context"
	"fmt"
	"unicode/utf8"

	"github.com/AdventurerAmer/todo-api/failures"
	"github.com/AdventurerAmer/todo-api/internal/core/domain"
	"github.com/AdventurerAmer/todo-api/internal/core/ports"
)

type Config struct {
	TitleMaxChars       int
	DescriptionMaxChars int
}

func DefaultConfig() Config {
	return Config{
		TitleMaxChars:       1024,
		DescriptionMaxChars: 2048,
	}
}

type service struct {
	Config
	listsRepo ports.ListsRepository
}

func New(listsRepo ports.ListsRepository, config Config) ports.ListsService {
	return &service{
		Config:    config,
		listsRepo: listsRepo,
	}
}

func (srv *service) Create(ctx context.Context, user domain.User, req ports.CreateListRequest) (ports.CreateListResponse, error) {
	v := failures.NewValidator()

	v.CheckUTF8("title", req.Title)
	v.CheckNotEmpty("title", req.Title)
	v.CheckAtMostInc("title", utf8.RuneCountInString(req.Title), srv.TitleMaxChars, "characters long")

	v.CheckUTF8("description", req.Description)
	v.CheckNotEmpty("description", req.Description)
	v.CheckAtMostInc("description", utf8.RuneCountInString(req.Description), srv.DescriptionMaxChars, "characters long")

	if err := v.Err(); err != nil {
		return ports.CreateListResponse{}, fmt.Errorf("validation failed: %w", err)
	}

	list := &domain.List{
		UserID:      user.ID,
		Title:       req.Title,
		Description: req.Description,
	}
	if err := srv.listsRepo.Create(ctx, list); err != nil {
		return ports.CreateListResponse{}, fmt.Errorf("'listsRepo.Create' failed: %w", err)
	}

	resp := ports.CreateListResponse{
		List: list,
	}
	return resp, nil
}

func (srv *service) Get(ctx context.Context, req ports.GetListRequest) (ports.GetListResponse, error) {
	v := failures.NewValidator()
	v.CheckUTF8("id", req.ID)
	v.CheckNotEmpty("id", req.ID)

	if err := v.Err(); err != nil {
		return ports.GetListResponse{}, fmt.Errorf("validation failed: %w", err)
	}

	list, err := srv.listsRepo.Get(ctx, req.ID)
	if err != nil {
		return ports.GetListResponse{}, fmt.Errorf("'listsRepo.Get' failed: %w", err)
	}

	resp := ports.GetListResponse{
		List: &list,
	}
	return resp, nil
}

func (srv *service) Update(ctx context.Context, req ports.UpdateListRequest) (ports.UpdateListResponse, error) {
	v := failures.NewValidator()

	v.CheckUTF8("id", req.ID)
	v.CheckNotEmpty("id", req.ID)

	if req.Title != nil {
		v.CheckUTF8("title", *req.Title)
		v.CheckNotEmpty("title", *req.Title)
		v.CheckAtMostInc("title", utf8.RuneCountInString(*req.Title), srv.TitleMaxChars, "characters long")
	}

	if req.Description != nil {
		v.CheckUTF8("description", *req.Description)
		v.CheckNotEmpty("description", *req.Description)
		v.CheckAtMostInc("description", utf8.RuneCountInString(*req.Description), srv.DescriptionMaxChars, "characters long")
	}

	if err := v.Err(); err != nil {
		return ports.UpdateListResponse{}, fmt.Errorf("validation failed: %w", err)
	}

	list, err := srv.listsRepo.Get(ctx, req.ID)
	if err != nil {
		return ports.UpdateListResponse{}, fmt.Errorf("'listsRepo.Get' failed: %w", err)
	}

	if req.Title != nil {
		list.Title = *req.Title
	}
	if req.Description != nil {
		list.Description = *req.Description
	}

	if err := srv.listsRepo.Update(ctx, &list); err != nil {
		return ports.UpdateListResponse{}, fmt.Errorf("'listsRepo.Update' failed: %w", err)
	}

	resp := ports.UpdateListResponse{
		List: &list,
	}
	return resp, nil
}

func (srv *service) Delete(ctx context.Context, req ports.DeleteListRequest) (ports.DeleteListResponse, error) {
	v := failures.NewValidator()
	v.CheckUTF8("id", req.ID)
	v.CheckNotEmpty("id", req.ID)

	if err := v.Err(); err != nil {
		return ports.DeleteListResponse{}, fmt.Errorf("validation failed: %w", err)
	}

	if err := srv.listsRepo.Delete(ctx, req.ID); err != nil {
		return ports.DeleteListResponse{}, fmt.Errorf("'listsRepo.Delete' failed: %w", err)
	}

	resp := ports.DeleteListResponse{
		Message: "list was deleted successfully",
	}

	return resp, nil
}
