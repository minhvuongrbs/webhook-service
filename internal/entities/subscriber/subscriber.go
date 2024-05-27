package subscriber

import "time"

type Subscriber struct {
	Id           string    `json:"id"`
	Status       string    `json:"status"`
	Email        string    `json:"email"`
	Source       string    `json:"source"`
	FirstName    string    `json:"first_name"`
	LastName     string    `json:"last_name"`
	Segments     []Segment `json:"segments"`
	CustomFields struct {
		Property1 string `json:"property1"`
		Property2 string `json:"property2"`
	} `json:"custom_fields"`
	OptinIp        string    `json:"optin_ip"`
	OptinTimestamp time.Time `json:"optin_timestamp"`
	CreatedAt      time.Time `json:"created_at"`
}

type Segment struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}
