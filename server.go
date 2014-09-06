package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"text/template"

	"github.com/go-martini/martini"
	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
	"github.com/martini-contrib/render"
	"github.com/mbardea/podd.club/logger"
	"github.com/mbardea/podd.club/util"
)

var rssTemplate = `{{ $host := .Host }}<?xml version="1.0" encoding="utf-8"?>
<rss version="2.0" xmlns:itunes="http://www.itunes.com/DTDs/Podcast-1.0.dtd" xmlns:media="http://search.yahoo.com/mrss/">

<channel>
<title>{{.Category.Name}}</title>
<description>Custom Podcast feeds from YouTube</description>
<itunes:author>Manuel Bardea, Paul Bardea</itunes:author>
<link>http://{{.Host}}</link>
<itunes:image href="http://{{.Host}}/poddclub.png" />
<pubDate>Fri, 05 Sep 2014 21:00:00 EST </pubDate>
<language>en-us</language>
<copyright>Original Authors</copyright>

{{range $podcast := .Podcasts}}
	<item>
		<title>{{$podcast.Title}}</title>
		<description></description>
		<itunes:author></itunes:author>
		<pubDate>Fri, 05 Sep 2014 21:00:00 EST</pubDate>
		<enclosure url="http://{{$host}}/api/podcasts/{{$podcast.Id}}/download" length="{{$podcast.Duration}}" type="audio/mpeg" /> 
	</item>
{{end}}
</channel>
</rss>
`

type Category struct {
	Id     int64  `primaryKey:"yes" json:"id"`
	UserId int64  `json:"user_id"`
	Name   string `json:"name"`
}

func (c Category) TableName() string {
	return "categories"
}

type Podcast struct {
	Id               int64 `json: "id" primaryKey:"yes"`
	UserId           int64 `json:"user_id"`
	CategoryId       int64
	Title            string `json:"title"`
	Description      string
	Duration         int `json:"duration"`
	DownloadMetadata string
}

type SimplePodcast struct {
	Id         int64  `json:"id" primaryKey:"yes"`
	UserId     int64  `json:"user_id"`
	CategoryId int64  `json:"category_id"`
	Title      string `json:"title"`
	Duration   int    `json:"duration"`
}

func (SimplePodcast) TableName() string {
	return "podcasts"
}

type User struct {
	Id       int64 `primaryKey:"yes"`
	Name     string
	Email    string
	Password string
}

type SimpleUser struct {
	Id    int64 `primaryKey:"yes"`
	Name  string
	Email string
}

func (SimpleUser) TableName() string {
	return "users"
}

func testDb(db *gorm.DB) {
	db.LogMode(true)
	var cat Category
	db.First(&cat, 1)

	fmt.Printf("RSS here: User: %v \n", cat)
}

type MyBuffer struct {
}

type DownloadMeta struct {
	Title       string
	Description string
	Duration    int
}

type DownloadJob struct {
	UserId     int64
	CategoryId int64
	Url        string
	Db         *gorm.DB
	err        error
}

// type DownloadRequest struct {
// 	Url string
// }
//

func downloadWorker(job *DownloadJob) {
	url := job.Url
	db := job.Db

	logger.Infof("Downloading from URL: %s", job.Url)

	var err error
	tmpDir, err := ioutil.TempDir("", "ydownload")
	if err != nil {
		log.Panicf("Could not create temporary directory")
	}
	name := path.Join(tmpDir, "audio")
	audioFileName := name + ".m4a"
	metaFileName := path.Join(tmpDir, "audio.info.json")

	cmd := exec.Command("youtube-dl",
		"-x", "--audio-format=m4a",
		"-o", audioFileName,
		"--write-info-json", url)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	stdout.Grow(1000)
	stderr.Grow(1000)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		log.Printf("Download Failed: %s: %s, %s", url, err, stderr.String())
		return
	}

	var metaBuffer []byte
	metaFile := "out.info.json"
	metaBuffer, err = ioutil.ReadFile(metaFileName)
	if err != nil {
		log.Printf("Could not read file %s", metaFile)
		return
	}
	var downloadMeta DownloadMeta
	err = json.Unmarshal(metaBuffer, &downloadMeta)
	if err != nil {
		log.Printf("Could not parse download meta Json")
		return
	}

	podcast := Podcast{
		UserId:           job.UserId,
		CategoryId:       job.CategoryId,
		Title:            downloadMeta.Title,
		Description:      downloadMeta.Description,
		Duration:         downloadMeta.Duration,
		DownloadMetadata: string(metaBuffer)}

	db.Save(&podcast)

	// Move file in the media directory
	filePath := path.Join("media", fmt.Sprintf("%d", podcast.UserId))
	err = os.MkdirAll(filePath, os.ModePerm)
	if err != nil {
		log.Printf("Could not create directory %s", filePath)
	}
	fileName := fmt.Sprintf("%d.m4a", podcast.Id)
	newAudioFileName := path.Join(filePath, fileName)
	err = os.Rename(audioFileName, newAudioFileName)
	if err != nil {
		log.Printf("Could not move media file into %s", newAudioFileName)
		return
	}
	newMetaFileName := path.Join(fmt.Sprintf("%d.json", podcast.Id))
	newMetaFullPath := path.Join(filePath, newMetaFileName)
	err = os.Rename(metaFileName, newMetaFullPath)
	if err != nil {
		log.Printf("Could not move meta file name to %s", newMetaFileName)
		return
	}
}

func main() {
	m := martini.Classic()

	var db gorm.DB
	var err error
	db, err = gorm.Open("postgres", "user=podd dbname=podd password=podd sslmode=disable")
	// db, err = gorm.Open("postgres", "host=localhost user=et dbname=et password=et sslmode=disable")
	if err != nil {
		panic(fmt.Sprintf("Cannot connect to the database: %s", err))
	}
	testDb(&db)
	m.Map(&db)

	m.Use(render.Renderer())
	m.Use(martini.Static("ui"))

	m.Get("/", func() string {
		return "Hello world!"
	})

	m.Get("/rss/:category_id", func(w http.ResponseWriter, p martini.Params, r render.Render, db *gorm.DB) {
		var category Category
		id := string(p["category_id"])
		query := db.First(&category, id)
		if query.Error != nil {
			logger.Errorf("Error: %v", query.Error)
			r.Status(http.StatusNotFound)
			return
		}

		var podcasts []Podcast = []Podcast{}
		query = db.Where("user_id = ? and category_id = ?", category.UserId, category.Id).Find(&podcasts)
		if query.Error != nil && !query.RecordNotFound() {
			logger.Errorf("Error: %v", query.Error)
			r.Status(http.StatusInternalServerError)
			return
		}

		type Rss struct {
			Host     string
			Category Category
			Podcasts []Podcast
		}
		rss := &Rss{
			Host:     "192.168.0.10:3000",
			Category: category,
			Podcasts: podcasts,
		}
		templ := template.Must(template.New("rss").Parse(rssTemplate))
		r.Header().Add("Content-Type", "application/rss+xml")
		r.Status(http.StatusOK)
		err := templ.Execute(w, rss)
		if err != nil {
			logger.Errorf("Cannot execute RSS template: %s", err)
			r.Status(http.StatusInternalServerError)
			return
		}
	})

	m.Get("/api/users", func(p martini.Params, r render.Render, db *gorm.DB) {
		var users []SimpleUser
		db.Find(&users)
		r.JSON(200, users)
	})

	m.Get("/api/users/:user_id/categories", func(p martini.Params, r render.Render, db *gorm.DB) {
		var categories []Category
		userId := string(p["user_id"])
		query := db.Where("user_id = ?", userId).Find(&categories)
		if query.Error != nil {
			logger.Errorf("Could not query Categories")
			r.Status(http.StatusInternalServerError)
			return
		}
		r.JSON(200, categories)
	})

	m.Post("/api/users/:user_id/categories", func(p martini.Params, req *http.Request, r render.Render, db *gorm.DB) {
		userId, _ := strconv.ParseInt(p["user_id"], 10, 64)
		name := req.PostFormValue("name")

		var category = &Category{UserId: userId, Name: name}
		db.Save(category)
		r.JSON(200, "")
	})

	m.Get("/api/users/:user_id/categories/:category_id/podcasts", func(p martini.Params, r render.Render, db *gorm.DB) {
		userId := string(p["user_id"])
		categoryId := string(p["category_id"])
		var podcasts []SimplePodcast = []SimplePodcast{}
		query := db.Where("user_id = ? and category_id = ?", userId, categoryId).Find(&podcasts)
		if query.Error != nil && !query.RecordNotFound() {
			logger.Errorf("Error: %v", query.Error)
			r.Status(http.StatusInternalServerError)
			return
		}
		r.JSON(200, podcasts)
	})

	m.Get("/api/podcasts/:podcast_id", func(p martini.Params, r render.Render, db *gorm.DB) {
		podcastId := string(p["podcast_id"])
		var podcast SimplePodcast
		query := db.Where("id = ? ", podcastId).Find(&podcast)
		if query.Error != nil {
			logger.Errorf("Error: %v", query.Error)
			r.Status(http.StatusInternalServerError)
			return
		}
		r.JSON(200, podcast)
	})

	m.Get("/api/podcasts/:podcast_id/download", func(p martini.Params, req *http.Request, w http.ResponseWriter, r render.Render, db *gorm.DB) {
		podcastId := string(p["podcast_id"])

		var podcast Podcast
		query := db.Where("id = ? ", podcastId).Find(&podcast)
		if query.Error != nil {
			logger.Errorf("Podcast not found: %s", err)
			r.Status(http.StatusNotFound)
			return
		}

		headers := ""
		for k, v := range req.Header {
			headers = headers + fmt.Sprintf("%s: %s\n", k, v)
		}
		logger.Errorf("Headers: %s", headers)

		fileName := fmt.Sprintf("media/1/%s.m4a", podcastId)
		file, err := os.Open(fileName)
		if err != nil {
			logger.Errorf("Cannot open file: %s. %s", fileName, err)
			r.Status(http.StatusNotFound)
			return
		}
		defer file.Close()

		stat, err := file.Stat()
		if err != nil {
			logger.Errorf("Cannot open file stats: %s. %s", fileName, err)
			r.Status(http.StatusInternalServerError)
			return
		}

		var startPos int64 = 0
		var endPos int64 = stat.Size()
		startPos, endPos = util.ParseRangeHeader(&req.Header, startPos, endPos)

		logger.Infof("Range request received: %v", startPos, endPos)

		_, err = file.Seek(startPos, os.SEEK_SET)
		if err != nil {
			logger.Errorf("Failed to seek. File: %s. %s", fileName, err)
			r.Status(http.StatusInternalServerError)
			return
		}

		w.Header().Add("Content-Length", fmt.Sprintf("%d", endPos-startPos))
		w.Header().Add("Cache-Control", "private")
		w.Header().Add("Pragma", "private")
		w.Header().Add("X-Content-Duration", strconv.Itoa(podcast.Duration))
		w.Header().Add("Content-Type", "audio/mp4")

		io.CopyN(w, file, endPos-startPos)
	})

	m.Post("/api/users/:user_id/categories/:category_id/schedule-download", func(p martini.Params, req *http.Request, r render.Render, db *gorm.DB) (int, string) {
		userId, _ := strconv.ParseInt(p["user_id"], 10, 64)
		categoryId, _ := strconv.ParseInt(p["category_id"], 10, 64)
		url := req.PostFormValue("url")
		// return fmt.Sprintf("Url: %v", url2)

		if !strings.HasPrefix(url, "http") {
			logger.Errorf("Bad download URL: %s", url)
			return http.StatusBadRequest, "Invalid URL"
		}

		job := &DownloadJob{
			UserId:     userId,
			CategoryId: categoryId,
			Url:        url,
			Db:         db}

		go downloadWorker(job)

		// out := stdout.String()
		// out := fmt.Sprintf("%v", downloadMeta)
		out := "Scheduled"
		return http.StatusOK, out
	})

	m.Run()
}
