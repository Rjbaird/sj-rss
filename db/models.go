package db

type Series struct {
	Name       string `json:"name" redis:"name"`
	Handle     string `json:"handle" redis:"handle"`
	URL        string `json:"url" redis:"url"`
	LastUpdate int64  `json:"last_update" redis:"last_update"`
	Image      string `json:"image" redis:"image"`
}
