package downloader

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"

	"github.com/jinzhu/gorm"
	"github.com/mbardea/podd.club/logger"
	"github.com/mbardea/podd.club/model"
)

const (
	MEDIA_BASE_DIR = "media"
)

func MediaDirName(userId int64) string {
	return path.Join(MEDIA_BASE_DIR, fmt.Sprintf("%d", userId))
}

func MediaAudioFileName(userId int64, podcastId int64) string {
	return path.Join(MediaDirName(userId), fmt.Sprintf("%d.mp3", podcastId))
}

func MediaMetaFileName(userId int64, podcastId int64) string {
	return path.Join(MediaDirName(userId), fmt.Sprintf("%d.json", podcastId))
}

func testDb(db *gorm.DB) {
	db.LogMode(true)
	var cat model.Category
	db.First(&cat, 1)

	fmt.Printf("RSS here: User: %v \n", cat)
}

func runCommand(cmd *exec.Cmd) (bytes.Buffer, bytes.Buffer, error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	stdout.Grow(1000)
	stderr.Grow(1000)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		logger.Errorf("Command failed: %v", cmd)
	}
	return stdout, stderr, err
}

func DownloadWorker(job *model.DownloadJob) {
	url := job.Url
	db := job.Db

	logger.Infof("Downloading from URL: %s", job.Url)

	var err error
	tmpDir, err := ioutil.TempDir("", "ydownload")
	if err != nil {
		log.Panicf("Could not create temporary directory")
	}
	name := path.Join(tmpDir, "audio")
	origAudioFile := name + ".m4a"
	convertedAudioFile := name + ".mp3"
	metaFileName := path.Join(tmpDir, "audio.info.json")

	cmd := exec.Command("youtube-dl",
		"-x", "--audio-format=m4a",
		"-o", origAudioFile,
		"--write-info-json", url)

	_, stderr, err := runCommand(cmd)
	if err != nil {
		log.Printf("Download Failed: %s: %s, %s", url, err, stderr.String())
		return
	}

	cmd = exec.Command("avconv",
		"-i", origAudioFile,
		"-b", "64k",
		convertedAudioFile)

	_, stderr, err = runCommand(cmd)
	if err != nil {
		log.Printf("MP3 Conversion failed: %s - %s: %s", convertedAudioFile, err, stderr.String())
		return
	}

	var metaBuffer []byte
	metaFile := "out.info.json"
	metaBuffer, err = ioutil.ReadFile(metaFileName)
	if err != nil {
		log.Printf("Could not read file %s", metaFile)
		return
	}
	var downloadMeta model.DownloadMeta
	err = json.Unmarshal(metaBuffer, &downloadMeta)
	if err != nil {
		log.Printf("Could not parse download meta Json")
		return
	}

	podcast := model.Podcast{
		UserId:           job.UserId,
		CategoryId:       job.CategoryId,
		SourceUrl:        job.Url,
		Title:            downloadMeta.Title,
		Description:      downloadMeta.Description,
		Duration:         downloadMeta.Duration,
		Thumbnail:        downloadMeta.Thumbnail,
		DownloadMetadata: string(metaBuffer)}

	db.Save(&podcast)

	// Move file in the media directory
	baseDir := MediaDirName(podcast.UserId)
	err = os.MkdirAll(baseDir, os.ModePerm)
	if err != nil {
		log.Printf("Could not create directory %s", baseDir)
	}
	newAudioFileName := MediaAudioFileName(podcast.UserId, podcast.Id)
	err = os.Rename(convertedAudioFile, newAudioFileName)
	if err != nil {
		log.Printf("Could not move media file into %s", newAudioFileName)
		return
	}
	newMetaFileName := MediaMetaFileName(podcast.UserId, podcast.Id)
	err = os.Rename(metaFileName, newMetaFileName)
	if err != nil {
		log.Printf("Could not move meta file name to %s", newMetaFileName)
		return
	}

	// Update the audio file size in the DB
	stat, err := os.Stat(newAudioFileName)
	if err != nil {
		log.Printf("Could not read stas for audio file %s", newAudioFileName)
		return
	}
	podcast.Size = stat.Size()
	db.Save(&podcast)
}
