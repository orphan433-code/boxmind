package pagemeta

import (
	"context"

	"pet-link/internal/domain"
)

type MetaFallback struct {
	extractor Extractor
}

func NewMetaFallback(extractor Extractor) *MetaFallback {
	return &MetaFallback{extractor: extractor}
}

func (f *MetaFallback) FallbackEnrich(ctx context.Context, rawURL string) (domain.BookmarkEnrichment, bool) {
	return FallbackEnrichment(ctx, f.extractor, rawURL)
}
