package application

// StackResponse contains the response that we receive from 
// stack overflow api
type StackResponse struct {
	Items []Item `json:"items"`
}

// Item stores the link and title that are necessary to return in slack
type Item struct {
	Link  string `json:"link"`
	Title string `json:"title"`
}
