package main

type payload struct {
	Url    string `json:"url"`
	Size   int64  `json:"size"`
	SHA1   string `json:"sha1"`
	SHA256 string `json:"sha256"`
}
