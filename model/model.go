package model

import "github.com/jinzhu/gorm"

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
	Duration         int    `json:"duration"`
	Thumbnail        string `json:thumbnail`
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

type MyBuffer struct {
}

type DownloadMeta struct {
	Title       string
	Description string
	Duration    int
	Thumbnail   string
}

type DownloadJob struct {
	UserId     int64
	CategoryId int64
	Url        string
	Db         *gorm.DB
	err        error
}
