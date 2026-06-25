package pagemeta

import "context"

type ImageFetcher struct {
	extractor Extractor
}

func NewImageFetcher(extractor Extractor) *ImageFetcher {
	return &ImageFetcher{extractor: extractor}
}

func (f *ImageFetcher) FetchImageURL(ctx context.Context, rawURL string) (string, error) {
	if thumb := PlatformThumbnailURL(rawURL); thumb != "" {
		return thumb, nil
	}

	page, err := f.extractor.Extract(ctx, rawURL)
	if err != nil {
		return "", err
	}
	return page.ImageURL, nil
}
