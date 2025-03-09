package vkservice

type AttachmentPostFave struct {
	Type string `json:"type"`
	Doc  *struct {
		Url string `json:"url"`
	} `json:"doc"`
	Article *struct {
		Url   string `json:"url"`
		Title string `json:"title"`
	} `json:"article"`
	Link *struct {
		Url string `json:"url"`
	} `json:"link"`
	Photo *struct {
		OrigPhoto *struct {
			Url string `json:"url"`
		} `json:"orig_photo"`
	} `json:"photo"`
}

type PostFave struct {
	Id      int    `json:"id"`
	FromId  int    `json:"from_id"`
	OwnerId int    `json:"owner_id"`
	Text    string `json:"text"`
	Views   struct {
		Count int `json:"count"`
	} `json:"views"`
	Attachments []AttachmentPostFave `json:"attachments"`
}

type FaveResponseItem struct {
	AddedDate int      `json:"added_date"`
	Type      string   `json:"type"`
	Post      PostFave `json:"post"`
}

type FavesPostResponse struct {
	Response struct {
		Count int
		Items []faveResponseItem
	}
}

func (v *VkService) fetchPostFaves() error {
	return nil
}

func (v *VkService) fetchFaves() error {
	return nil
}
