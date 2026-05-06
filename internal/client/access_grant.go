package client

// AccessGrant represents a single access permission entry for a resource.
type AccessGrant struct {
	//ID            string `json:"id"`
	Permission    string `json:"permission"`
	PrincipalID   string `json:"principal_id"`
	PrincipalType string `json:"principal_type"`
	//ResourceType  string `json:"resource_type"`
	//ResourceID    string `json:"resource_id"`
	//CreatedAt     int64  `json:"created_at"`
}
