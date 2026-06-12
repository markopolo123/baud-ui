package baud

// Crumb is one breadcrumb trail segment. Href is ignored on the last
// (current) crumb — it renders as bold text with aria-current="page",
// never a link.
type Crumb struct {
	Label string
	Href  string
}

// BreadcrumbProps configures the breadcrumb trail (see components/lists.css):
// `prod › core › ingest-gw` — non-current crumbs are --fg-faint links, the
// › separators are presentational (aria-hidden).
type BreadcrumbProps struct {
	Items []Crumb
}
