package collection

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/example/multitenant-search/internal/platform"
)

type collectionRepo struct{ c *platform.Collection }

func (r collectionRepo) CreateCollection(context.Context, *platform.Collection) error { return nil }
func (r collectionRepo) GetCollection(context.Context, string) (*platform.Collection, error) {
	return r.c, nil
}
func (r collectionRepo) UpdateCollection(context.Context, *platform.Collection) error { return nil }

func TestValidateAndExplainPreservesConflict(t *testing.T) {
	old := map[string]platform.FieldMapping{"title": {Type: platform.Text}}
	next := map[string]platform.FieldMapping{"title": {Type: platform.Keyword}}
	if err := ValidateAndExplain(old, next); !errors.Is(err, platform.ErrConflict) {
		t.Fatalf("conflict sentinel lost: %v", err)
	}
}

func TestPublishMappingCheckedClassifiesConflict(t *testing.T) {
	repo := collectionRepo{c: &platform.Collection{ID: "c1", Mappings: map[string]platform.FieldMapping{"title": {Type: platform.Text}}}}
	s := NewService(repo, platform.RealClock{})
	_, err := s.PublishMappingChecked(context.Background(), "c1", map[string]platform.FieldMapping{"title": {Type: platform.Keyword}})
	if !errors.Is(err, platform.ErrConflict) {
		t.Fatalf("publish conflict not classified: %v", err)
	}
}

func TestPublishMappingCheckedClassifiesInvalid(t *testing.T) {
	repo := collectionRepo{c: &platform.Collection{ID: "c1", Mappings: map[string]platform.FieldMapping{"title": {Type: platform.Text}}}}
	s := NewService(repo, platform.RealClock{})
	_, err := s.PublishMappingChecked(context.Background(), "c1", map[string]platform.FieldMapping{
		"title": {Type: platform.Text},
		"bad":   {Type: platform.FieldType("unsupported")},
	})
	if !errors.Is(err, platform.ErrInvalid) || !strings.Contains(err.Error(), "mapping rejected") {
		t.Fatalf("invalid mapping classification lost: %v", err)
	}
}
