package goodreads

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

var (
	apiRoot = "https://www.goodreads.com"
)

type Response struct {
	User    User     `xml:"user"`
	Book    Book     `xml:"book"`
	Reviews []Review `xml:"reviews>review"`
}

type User struct {
	ID            string       `xml:"id"`
	Name          string       `xml:"name"`
	About         string       `xml:"about"`
	Link          string       `xml:"link"`
	ImageURL      string       `xml:"image_url"`
	SmallImageURL string       `xml:"small_image_url"`
	Location      string       `xml:"location"`
	LastActive    string       `xml:"last_active"`
	ReviewCount   int          `xml:"reviews_count"`
	Statuses      []UserStatus `xml:"user_statuses>user_status"`
	Shelves       []Shelf      `xml:"user_shelves>user_shelf"`
	LastRead      []Review
}

func (u User) ReadingShelf() Shelf {
	for _, shelf := range u.Shelves {
		if shelf.Name == "currently-reading" {
			return shelf
		}
	}

	return Shelf{}
}

func (u User) ReadShelf() Shelf {
	for _, shelf := range u.Shelves {
		if shelf.Name == "read" {
			return shelf
		}
	}

	return Shelf{}
}

func (u User) ToReadShelf() Shelf {
	for _, shelf := range u.Shelves {
		if shelf.Name == "to-read" {
			return shelf
		}
	}

	return Shelf{}
}

type Shelf struct {
	ID        string `xml:"id"`
	BookCount string `xml:"book_count"`
	Name      string `xml:"name"`
}

type UserStatus struct {
	Page    int    `xml:"page"`
	Percent int    `xml:"percent"`
	Updated string `xml:"updated_at"`
	Book    Book   `xml:"book"`
}

func (u UserStatus) UpdatedRelative() string {
	return relativeDate(u.Updated)
}

type Book struct {
	ID       string   `xml:"id"`
	Title    string   `xml:"title"`
	Link     string   `xml:"link"`
	ImageURL string   `xml:"image_url"`
	NumPages string   `xml:"num_pages"`
	Format   string   `xml:"format"`
	Authors  []Author `xml:"authors>author"`
	ISBN     string   `xml:"isbn"`
}

func (b Book) Author() Author {
	return b.Authors[0]
}

type Author struct {
	ID   string `xml:"id"`
	Name string `xml:"name"`
	Link string `xml:"link"`
}

type Review struct {
	Book   Book   `xml:"book"`
	Rating int    `xml:"rating"`
	ReadAt string `xml:"read_at"`
	Link   string `xml:"link"`
}

func (r Review) FullStars() []bool {
	return make([]bool, r.Rating)
}

func (r Review) EmptyStars() []bool {
	return make([]bool, 5-r.Rating)
}

func (r Review) ReadAtShort() string {
	date, err := parseDate(r.ReadAt)
	if err != nil {
		return ""
	}

	return (string)(date.Format("2 Jan 2006"))
}

func (r Review) ReadAtRelative() string {
	return relativeDate(r.ReadAt)
}

// PUBLIC

type Client struct {
	apiKey     string
	httpClient *http.Client
}

func NewClient(apiKey string) *Client {
	return &Client{apiKey: apiKey, httpClient: http.DefaultClient}
}

func NewClientWithHttpClient(apiKey string, httpClient *http.Client) *Client {
	return &Client{apiKey: apiKey, httpClient: httpClient}
}

func (c *Client) GetUser(id string, limit int) (*User, error) {
	uri := apiRoot + "/user/show/" + id + ".xml?key=" + c.apiKey
	response := &Response{}
	err := c.getData(uri, response)
	if err != nil {
		return nil, err
	}

	for i := range response.User.Statuses {
		status := &response.User.Statuses[i]
		bookid := status.Book.ID
		book, err := c.GetBook(bookid)
		if err != nil {
			return nil, err
		}
		status.Book = *book
	}

	if len(response.User.Statuses) >= limit {
		response.User.Statuses = response.User.Statuses[:limit]
	} else {
		remaining := limit - len(response.User.Statuses)
		lastRead, err := c.GetLastRead(id, remaining)
		if err != nil {
			return nil, err
		}
		response.User.LastRead = lastRead
	}

	return &response.User, nil
}

func (c *Client) GetBook(id string) (*Book, error) {
	uri := apiRoot + "/book/show/" + id + ".xml?key=" + c.apiKey
	response := &Response{}
	err := c.getData(uri, response)
	if err != nil {
		return nil, err
	}

	return &response.Book, nil
}

func (c *Client) GetLastRead(id string, limit int) ([]Review, error) {
	l := strconv.Itoa(limit)
	uri := apiRoot + "/review/list/" + id + ".xml?key=" + c.apiKey + "&v=2&shelf=read&sort=date_read&order=d&per_page=" + l

	response := &Response{}
	err := c.getData(uri, response)
	if err != nil {
		return []Review{}, err
	}

	return response.Reviews, nil
}

func (c *Client) ReviewsForShelf(user *User, shelf string) ([]Review, error) {
	reviews := make([]Review, 0)
	perPage := 200
	pages := (user.ReviewCount / perPage) + 1

	// Keep looping until we have all the reviews
	for i := 1; i <= pages; i++ {
		uri := fmt.Sprintf("%s/review/list/%s.xml?key=%s&v=2&page=%d&per_page=%d", apiRoot, user.ID, c.apiKey, i, perPage)
		response := &Response{}
		err := c.getData(uri, response)
		if err != nil {
			return []Review{}, err
		}

		reviews = append(reviews, response.Reviews...)
	}

	return reviews, nil
}

// PRIVATE

func (c *Client) getData(uri string, i interface{}) error {
	data, err := c.getRequest(uri)
	if err != nil {
		return err
	}
	return xmlUnmarshal(data, i)
}

func (c *Client) getRequest(uri string) ([]byte, error) {
	res, err := c.httpClient.Get(uri)
	if err != nil {
		return []byte{}, err
	}

	body, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return []byte{}, err
	}

	return body, nil
}

func xmlUnmarshal(b []byte, i interface{}) error {
	return xml.Unmarshal(b, i)
}

func parseDate(s string) (time.Time, error) {
	date, err := time.Parse(time.RFC3339, s)
	if err != nil {
		date, err = time.Parse(time.RubyDate, s)
		if err != nil {
			return time.Time{}, err
		}
	}

	return date, nil
}

func relativeDate(d string) string {
	date, err := parseDate(d)
	if err != nil {
		fmt.Println(err)
		return ""
	}

	s := time.Now().Sub(date)

	days := int(s / (24 * time.Hour))
	if days > 1 {
		return fmt.Sprintf("%v days ago", days)
	} else if days == 1 {
		return fmt.Sprintf("%v day ago", days)
	}

	hours := int(s / time.Hour)
	if hours > 1 {
		return fmt.Sprintf("%v hours ago", hours)
	}

	minutes := int(s / time.Minute)
	if minutes > 2 {
		return fmt.Sprintf("%v minutes ago", minutes)
	} else {
		return "Just now"
	}
}
