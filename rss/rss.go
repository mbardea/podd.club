package rss

import (
	"html/template"
	"io"

	"github.com/mbardea/podd.club/model"
)

const rssTemplate = `{{ $host := .Host }}{{ .Header }} 
<rss version="2.0" xmlns:itunes="http://www.itunes.com/dtds/podcast-1.0.dtd" xmlns:media="http://search.yahoo.com/mrss/">

<channel>
<title>{{ .Category.Name }}</title>
<description>Custom Podcast feeds from YouTube</description>
<itunes:author>Manuel Bardea, Paul Bardea</itunes:author>
<link>http://{{ .Host }}</link>
<itunes:image href="http://{{ .Host }}/poddclub.png" />
<pubDate>Fri, 05 Sep 2014 21:00:00 EST </pubDate>
<language>en-us</language>
<copyright>Original Authors</copyright>

{{range $podcast := .Podcasts}}
	<item>
		<title>{{ $podcast.Title }}</title>
		<description>{{ .Description }}</description>
		<itunes:author></itunes:author>
		<pubDate>Fri, 05 Sep 2014 21:00:00 EST</pubDate>
		<guid>http://podd.club-{{$podcast.Id}}</guid>
		<enclosure url="http://{{ $host }}/api/podcasts/{{ $podcast.Id }}/download" length="{{ $podcast.Duration }}" type="audio/mpeg" /> 
		<itunes:image href="{{ .Thumbnail }}"></itunes:image>
	</item>
{{end}}
</channel>
</rss>
`

type Rss struct {
	Host     string
	Header   template.HTML
	Category model.Category
	Podcasts []model.Podcast
}

func Execute(w io.Writer, rss *Rss) error {
	templ := template.Must(template.New("rss").Parse(rssTemplate))
	err := templ.Execute(w, rss)
	return err
}
