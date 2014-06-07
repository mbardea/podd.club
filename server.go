package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/go-martini/martini"
	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
	"github.com/martini-contrib/render"
)

type Category struct {
	Id   int    `primaryKey:"yes" json:"id"`
	Name string `json:"name"`
}

type Podcast struct {
	Id               int64 `primaryKey:"yes"`
	CategoryId       int64
	Title            string
	Description      string
	DownloadMetadata string
}

type User struct {
	Id   int64 `primaryKey:"yes"`
	Name string
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
}

// type DownloadRequest struct {
// 	Url string
// }
//

func main() {
	m := martini.Classic()

	var db gorm.DB
	var err error
	db, err = gorm.Open("postgres", "user=poddpadd dbname=poddpadd password=poddpadd sslmode=disable")
	// db, err = gorm.Open("postgres", "host=localhost user=et dbname=et password=et sslmode=disable")
	if err != nil {
		panic(fmt.Sprintf("Cannot connect to the database: %s", err))
	}
	testDb(&db)
	m.Map(&db)

	m.Use(render.Renderer())

	m.Get("/", func() string {
		return "Hello world!"
	})

	m.Get("/rss/:category_id", func(p martini.Params, r render.Render, db *gorm.DB) {
		var cat Category
		id := string(p["category_id"])
		db.First(&cat, id)

		// return fmt.Sprintf("Category: %s", cat.name)
		r.JSON(200, cat)
	})

	m.Post("/api/download", func(req *http.Request, r render.Render, db *gorm.DB) string {
		url := req.PostFormValue("url")
		// return fmt.Sprintf("Url: %v", url2)

		if !strings.HasPrefix(url, "http") {
			log.Panicf("Bad download URL: %s", url)
		}

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
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		// r.JSON(200, map[string]interface{}{"status": "Scheduled"})
		err = cmd.Run()
		if err != nil {
			r.JSON(500, "Download Failed")
		}

		var metaBuffer []byte
		metaFile := "out.info.json"
		metaBuffer, err = ioutil.ReadFile(metaFileName)
		if err != nil {
			log.Fatalf("Could not read file %s", metaFile)
		}
		var downloadMeta DownloadMeta
		err = json.Unmarshal(metaBuffer, &downloadMeta)
		if err != nil {
			log.Fatalf("Could not parse download meta Json")
		}

		podcast := Podcast{
			CategoryId:       1,
			Title:            downloadMeta.Title,
			Description:      downloadMeta.Description,
			DownloadMetadata: string(metaBuffer)}

		db.Save(&podcast)

		// Move file in the media directory
		newAudioFileName := path.Join(fmt.Sprintf("media/%d.m4a", podcast.Id))
		err = os.Rename(audioFileName, newAudioFileName)
		if err != nil {
			log.Fatalf("Could not move media file into %s", newAudioFileName)
		}
		newMetaFileName := path.Join(fmt.Sprintf("media/%d.json", podcast.Id))
		err = os.Rename(metaFileName, newMetaFileName)
		if err != nil {
			log.Fatalf("Could not move meta file name to %s", newMetaFileName)
		}

		// out := stdout.String()
		// out := fmt.Sprintf("%v", downloadMeta)
		out := "Done"

		return out
	})

	m.Run()
}
