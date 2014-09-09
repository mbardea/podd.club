package rss

import (
	"html/template"
	"io"

	"github.com/mbardea/podd.club/model"
)

const rssTemplate = `{{ $host := .Host }}{{ XmlHeader }}
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
		<description>{{ .Description | Cdata }}</description>
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
	Category model.Category
	Podcasts []model.Podcast
}

type RssString string

func Cdata(str string) template.HTML {
	return template.HTML("<![CDATA[" + str + "]]>")
}

func Trusted(str string) template.HTML {
	return template.HTML(str)
}

func XmlHeader() template.HTML {
	return Trusted(`<?xml version="1.0" encoding="utf-8"?>`)
}

func Execute(w io.Writer, rss *Rss) error {
	funcMap := template.FuncMap{
		"XmlHeader": XmlHeader,
		"Cdata":     Cdata,
	}

	templ := template.
		Must(template.New("rss").
		Funcs(funcMap).
		Parse(rssTemplate))
	err := templ.Execute(w, rss)
	return err
}
