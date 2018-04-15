package goodreads

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
)

var (
	apiRoot = "https://www.goodreads.com"
)

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
		uri := fmt.Sprintf("%s/review/list/%s.xml?key=%s&v=2&page=%d&per_page=%d&shelf=%s", apiRoot, user.ID, c.apiKey, i, perPage, shelf)
		response := &Response{}
		err := c.getData(uri, response)
		if err != nil {
			return []Review{}, err
		}

		reviews = append(reviews, response.Reviews...)
	}

	return reviews, nil
}

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
