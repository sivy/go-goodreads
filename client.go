package goodreads

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
)

const (
	apiRoot        = "https://www.goodreads.com"
	reviewListPath = "/review/list.xml"
	authorShowPath = "/author/show.xml"
	bookShowPath   = "/book/show.xml"
	userShowPath   = "/user/show.xml"
)

// Client is the main GoodReads client
type Client struct {
	apiKey     string
	httpClient *http.Client
}

// NewClient returns a Client
func NewClient(apiKey string) *Client {
	return &Client{apiKey: apiKey, httpClient: http.DefaultClient}
}

// NewClientWithHTTPClient returns a Client from an existing http.Client
func NewClientWithHTTPClient(apiKey string, httpClient *http.Client) *Client {
	return &Client{apiKey: apiKey, httpClient: httpClient}
}

// GetUser sends a /user/show.xml request
func (c *Client) GetUser(id string, limit int) (*User, error) {
	params := toURLValues(map[string]string{
		"key": c.apiKey,
		"id":  id,
	})
	response := &Response{}
	err := c.getData(userShowPath, params, response)
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

// GetBook sends a /book/show.xml request
func (c *Client) GetBook(id string) (*Book, error) {
	params := toURLValues(map[string]string{
		"key": c.apiKey,
		"id":  id,
	})
	response := &Response{}
	err := c.getData(bookShowPath, params, response)
	if err != nil {
		return nil, err
	}

	return &response.Book, nil
}

// GetAuthor sends a /author/show.xml request
func (c *Client) GetAuthor(id string) (*Author, error) {
	params := toURLValues(map[string]string{
		"key": c.apiKey,
		"id":  id,
	})
	response := &AuthorResponse{}
	err := c.getData(authorShowPath, params, response)
	if err != nil {
		return nil, err
	}
	return &response.Author, nil
}

// GetLastRead sends a /review/list.xml request
func (c *Client) GetLastRead(id string, limit int) ([]Review, error) {
	l := strconv.Itoa(limit)
	params := toURLValues(map[string]string{
		"key":      c.apiKey,
		"v":        "2",
		"id":       id,
		"shelf":    "read",
		"sort":     "date_read",
		"order":    "d",
		"per_page": l,
	})

	response := &Response{}
	err := c.getData(reviewListPath, params, response)
	if err != nil {
		return []Review{}, err
	}

	return response.Reviews, nil
}

// ReviewsForShelf returns book reviews for one of the GoodReads
// shelves
func (c *Client) ReviewsForShelf(user *User, shelf string) ([]Review, error) {
	reviews := make([]Review, 0)
	perPage := 200
	pages := (user.ReviewCount / perPage) + 1

	// Keep looping until we have all the reviews
	for i := 1; i <= pages; i++ {
		params := toURLValues(map[string]string{
			"key":      c.apiKey,
			"id":       user.ID,
			"v":        "2",
			"page":     strconv.Itoa(i),
			"per_page": strconv.Itoa(perPage),
			"shelf":    shelf,
		})

		response := &Response{}
		err := c.getData(reviewListPath, params, response)
		if err != nil {
			return []Review{}, err
		}

		reviews = append(reviews, response.Reviews...)
	}

	return reviews, nil
}

func (c *Client) getData(path string, params url.Values, i interface{}) error {
	uri, err := url.Parse(fmt.Sprintf("%s%s", apiRoot, path))
	if err != nil {
		return err
	}
	uri.RawQuery = params.Encode()

	data, err := c.getRequest(uri.String())
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

func toURLValues(m map[string]string) url.Values {
	params := url.Values{}
	for key, value := range m {
		params.Add(key, value)
	}
	return params
}
