// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package cloudplan

import (
	"encoding/json"
	"errors"
	"io"
	"os"
)

var ErrInvalidRemotePlanFormat = errors.New("invalid remote plan format, must be 1")
var ErrInvalidRunID = errors.New("invalid run ID")
var ErrInvalidHostname = errors.New("invalid hostname")

type SavedPlanBookmark struct {
	RemotePlanFormat int    `json:"remote_plan_format"`
	RunID            string `json:"run_id"`
	Hostname         string `json:"hostname"`
}

func LoadSavedPlanBookmark(filepath string) (SavedPlanBookmark, error) {
	bookmark := SavedPlanBookmark{}

	file, err := os.Open(filepath)
	if err != nil {
		return bookmark, err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return bookmark, err
	}

	e := json.Unmarshal([]byte(data), &bookmark)
	if e != nil {
		return bookmark, e
	}

	if bookmark.RemotePlanFormat != 1 {
		return bookmark, ErrInvalidRemotePlanFormat
	} else if bookmark.Hostname == "" {
		return bookmark, ErrInvalidHostname
	} else if bookmark.RunID == "" {
		return bookmark, ErrInvalidRunID
	}

	return bookmark, err
}

func (s *SavedPlanBookmark) Save(filepath string) error {
	data, _ := json.Marshal(s)

	err := os.WriteFile(filepath, data, 0644)
	if err != nil {
		return err
	}

	return nil
}
