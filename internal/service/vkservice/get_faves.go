package vkservice

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/Vikot10/viarticles/internal/dto"
)

type faveResponseItem struct {
}

type favesResponse struct {
	Response struct {
		Count int
		Items []faveResponseItem
	}
}

func (v *VkService) GetFaves() ([]*dto.Fave, error) {
	uri := "https://api.vk.com/method/fave.get"
	form := url.Values{
		"access_token": {v.accessToken},
		"v":            {"5.199"},
	}
	resp, errResp := http.PostForm(uri, form)
	if errResp != nil {
		return nil, fmt.Errorf("error get faves: %w", errResp)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error get faves, unexpected status: %s", resp.Status)
	}

	body, errReadAllBody := io.ReadAll(resp.Body)
	if errReadAllBody != nil {
		return nil, fmt.Errorf("error read all body: %w", errReadAllBody)
	}

	var faves favesResponse
	errDecode := json.Unmarshal(body, &faves)
	if errDecode != nil {
		return nil, fmt.Errorf("error decode get faves: %w", errDecode)
	}

	return nil, nil
}
