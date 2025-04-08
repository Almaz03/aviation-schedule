package model

type Flight struct {
	ID            int    `json:"id"`
	Number        string `json:"number"`
	Origin        string `json:"origin"`
	Destination   string `json:"destination"`
	DepartureTime string `json:"departure_time"`
	ArrivalTime   string `json:"arrival_time"`
	Status        string `json:"status"`
}
