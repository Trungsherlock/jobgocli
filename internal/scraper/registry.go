package scraper

import "fmt"

type Registry struct {
	scrapers map[string]Scraper
}

func NewRegistry() *Registry {
	r := &Registry{
		scrapers: make(map[string]Scraper),
	}

	r.Register(NewLeverScraper())
	r.Register(NewGreenhouseScraper())

	return r
}

func (r *Registry) Register(s Scraper) {
	r.scrapers[s.Name()] = s
}

func (r *Registry) Get(platform string) (Scraper, error) {
	s, ok := r.scrapers[platform]
	if !ok {
		return nil, fmt.Errorf("no scraper registered for platform: %s", platform)
	}
	return s, nil
}

func (r *Registry) Platforms() []string {
	platforms := make([]string, 0, len(r.scrapers))
	for k := range r.scrapers {
		platforms = append(platforms, k)
	}
	return platforms
}