package models

type Ad struct {
    ID        string `json:"id"`
    ImageURL  string `json:"image_url"`
    TargetURL string `json:"target_url"`
}
