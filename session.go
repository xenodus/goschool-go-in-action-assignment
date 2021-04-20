package main

import "net/url"

type session struct {
	Id           string
	LastModified int64
	LastVisited  *url.URL
}
