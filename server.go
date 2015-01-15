package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime/debug"
	"strconv"
	"strings"

	"github.com/go-martini/martini"
	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
	"github.com/martini-contrib/render"
	"github.com/martini-contrib/sessions"
	"github.com/mbardea/podd.club/auth"
	"github.com/mbardea/podd.club/downloader"
	"github.com/mbardea/podd.club/logger"
	"github.com/mbardea/podd.club/model"
	"github.com/mbardea/podd.club/rss"
	"github.com/mbardea/podd.club/util"
)

func testDb(db *gorm.DB) {
	db.LogMode(true)
	var cat model.Category
	db.First(&cat, 1)

	fmt.Printf("RSS here: User: %v \n", cat)
}

func httpError(r render.Render, statusCode int, msg string) {
	logger.Errorf(msg, "")
	debug.PrintStack()
	r.Data(statusCode, []byte(msg))
}

func main() {
	auth.InitRandom()
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

	store := sessions.NewCookieStore([]byte("234./2kdf@#.HL"))
	m.Use(sessions.Sessions("session", store))

	m.Get("/rss/:category_id", func(w http.ResponseWriter, p martini.Params, r render.Render, db *gorm.DB) {
		var category model.Category
		id := string(p["category_id"])
		query := db.First(&category, id)
		if query.Error != nil {
			logger.Errorf("Error: %v", query.Error)
			r.Status(http.StatusNotFound)
			return
		}

		var podcasts []model.Podcast = []model.Podcast{}
		query = db.Where("user_id = ? and category_id = ? order by id", category.UserId, category.Id).Find(&podcasts)
		if query.Error != nil && !query.RecordNotFound() {
			logger.Errorf("Error: %v", query.Error)
			r.Status(http.StatusInternalServerError)
			return
		}

		templateArgs := &rss.Rss{
			Host:     "podd.club",
			Category: category,
			Podcasts: podcasts,
		}
		r.Header().Add("Content-Type", "application/rss+xml")
		r.Status(http.StatusOK)
		rss.Execute(w, templateArgs)
		if err != nil {
			logger.Errorf("Cannot execute RSS template: %s", err)
			r.Status(http.StatusInternalServerError)
			return
		}
	})

	m.Get("/api/users", func(p martini.Params, r render.Render, db *gorm.DB) {
		var users []model.SimpleUser
		db.Find(&users)
		r.JSON(200, users)
	})

	m.Post("/api/users", func(p martini.Params, req *http.Request, r render.Render, db *gorm.DB) {
		name := req.PostFormValue("name")
		email := req.PostFormValue("email")
		password := req.PostFormValue("password")

		var users []model.User
		q := db.Where("email = ?", email).Find(&users)
		if !q.RecordNotFound() {
			httpError(r, http.StatusBadRequest, "Email already exists.")
			return
		}
		user := &model.User{
			Name:     name,
			Email:    email,
			Password: auth.MakePassword(password),
		}
		db.Save(user)
	})

	m.Put("/api/users/login", func(p martini.Params, req *http.Request, r render.Render, db *gorm.DB) {
		email := req.PostFormValue("email")
		password := req.PostFormValue("password")

		user := &model.User{}
		q := db.Where("email = ?", email).Find(user)
		if q.RecordNotFound() {
			r.Error(http.StatusNotFound)
			return
		}

		verified, err := auth.CheckPassword(password, user.Password)
		if err != nil {
			logger.Errorf("Error checking password: %s", err)
			r.Error(http.StatusInternalServerError)
			return
		}
		if verified {
			r.Status(http.StatusOK)
		} else {
			r.Status(http.StatusUnauthorized)
		}
	})

	m.Get("/api/users/:user_id/categories", func(p martini.Params, r render.Render, db *gorm.DB) {
		var categories []model.Category
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

		var category = &model.Category{UserId: userId, Name: name}
		db.Save(category)
		r.JSON(200, "")
	})

	m.Get("/api/users/:user_id/categories/:category_id/podcasts", func(p martini.Params, r render.Render, db *gorm.DB) {
		userId := string(p["user_id"])
		categoryId := string(p["category_id"])
		var podcasts []model.SimplePodcast = []model.SimplePodcast{}
		query := db.Where("user_id = ? and category_id = ? order by id", userId, categoryId).Find(&podcasts)
		if query.Error != nil && !query.RecordNotFound() {
			logger.Errorf("Error: %v", query.Error)
			r.Status(http.StatusInternalServerError)
			return
		}
		r.JSON(200, podcasts)
	})

	m.Get("/api/podcasts/:podcast_id", func(p martini.Params, r render.Render, db *gorm.DB) {
		podcastId := string(p["podcast_id"])
		var podcast model.SimplePodcast
		query := db.Where("id = ? ", podcastId).Find(&podcast)
		if query.Error != nil {
			logger.Errorf("Error: %v", query.Error)
			r.Status(http.StatusInternalServerError)
			return
		}
		r.JSON(200, podcast)
	})

	m.Delete("/api/podcasts/:podcast_id", func(p martini.Params, r render.Render, db *gorm.DB) {
		podcastId := string(p["podcast_id"])

		podcast := &model.Podcast{}
		query := db.Where("id = ? ", podcastId).Find(podcast)
		if query.Error != nil {
			r.Error(http.StatusNotFound)
			return
		}

		audioFile := downloader.MediaAudioFileName(podcast.UserId, podcast.Id)
		metaFile := downloader.MediaMetaFileName(podcast.UserId, podcast.Id)

		query = db.Where("id = ? ", podcastId).Delete(podcast)
		if query.Error != nil {
			r.Error(http.StatusNotFound)
			return
		}
		err := os.Remove(audioFile)
		if err != nil {
			logger.Errorf("Could not remove file %s", audioFile)
		}
		err = os.Remove(metaFile)
		if err != nil {
			logger.Errorf("Could not remove file %s", metaFile)
		}
		r.Error(http.StatusOK)
	})

	m.Get("/api/podcasts/:podcast_id/download", func(p martini.Params, req *http.Request, w http.ResponseWriter, r render.Render, db *gorm.DB) {
		podcastId := string(p["podcast_id"])

		var podcast model.Podcast
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
		logger.Infof("Headers: %s", headers)

		fileName := fmt.Sprintf("media/1/%s.mp3", podcastId)
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
		hasRangeHeader, startPos, endPos := util.ParseRangeHeader(&req.Header, startPos, endPos)

		logger.Infof("Serving range: %v", startPos, endPos)

		_, err = file.Seek(startPos, os.SEEK_SET)
		if err != nil {
			logger.Errorf("Failed to seek. File: %s. %s", fileName, err)
			r.Status(http.StatusInternalServerError)
			return
		}

		w.Header().Add("Content-Length", fmt.Sprintf("%d", endPos-startPos))
		// w.Header().Add("Cache-Control", "private")
		// w.Header().Add("Pragma", "private")
		w.Header().Add("Content-Type", "audio/mpeg")
		w.Header().Add("Last-Modified", "Wed, 03 Sep 2014 19:44:10 GMT")

		w.Header().Add("Accept-Ranges", "bytes")
		// w.Header().Add("X-Content-Duration", strconv.Itoa(podcast.Duration))
		if hasRangeHeader {
			w.Header().Add("Content-Range", fmt.Sprintf(" bytes %d-%d/%d", startPos, endPos, stat.Size()))
			r.Status(http.StatusPartialContent)
		} else {
			// w.Header().Add("X-Content-Duration", strconv.Itoa(podcast.Duration))
		}

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

		job := &model.DownloadJob{
			UserId:     userId,
			CategoryId: categoryId,
			Url:        url,
			Db:         db}

		go downloader.DownloadWorker(job)

		// out := stdout.String()
		// out := fmt.Sprintf("%v", downloadMeta)
		out := "Scheduled"
		return http.StatusOK, out
	})

	m.Run()
}
