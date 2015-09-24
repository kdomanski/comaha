package main

import (
  "io/ioutil"
  "encoding/json"
)

type payload struct {
  Url string `json:"url"`
  Size int64 `json:"size"`
  SHA1 string `json:"sha1"`
  SHA256 string `json:"sha256"`
}

type payloadBackend interface {
  GetPayload() *payload
}

type singleFileBackend struct {
  payloads []payload
}

func NewSingleFileBackend(filename string) (*singleFileBackend, error) {
  data, err := ioutil.ReadFile(filename)
  if err != nil {
    return nil, err
  }

  var payloads []payload
  err = json.Unmarshal(data, &payloads)
  if err != nil {
    return nil, err
  }

  return &singleFileBackend{payloads: payloads}, nil
}


func (b *singleFileBackend) GetPayload() *payload {
  return &b.payloads[0]
}
