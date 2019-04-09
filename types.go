package goodreads

// Response wraps a GoodReads GetUser response
type Response struct {
	User    User     `xml:"user"`
	Book    Book     `xml:"book"`
	Reviews []Review `xml:"reviews>review"`
	Updates []Update `xml:"updates>update"`
}

// AuthorResponse wraps the GoodReads GetBook response
type AuthorResponse struct {
	Author Author `xml:"author"`
}

// User represents a user element in the GoodReads response
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

// ReadingShelf returns the currently-reading shelf for a user
func (u User) ReadingShelf() Shelf {
	for _, shelf := range u.Shelves {
		if shelf.Name == "currently-reading" {
			return shelf
		}
	}

	return Shelf{}
}

// ReadShelf returns the read shelf for a user
func (u User) ReadShelf() Shelf {
	for _, shelf := range u.Shelves {
		if shelf.Name == "read" {
			return shelf
		}
	}

	return Shelf{}
}

// ToReadShelf returns the to-read shelf for a user
func (u User) ToReadShelf() Shelf {
	for _, shelf := range u.Shelves {
		if shelf.Name == "to-read" {
			return shelf
		}
	}

	return Shelf{}
}

// Actor represents an actor element in the GoodReads response
type Actor struct {
	ID       string `xml:"id"`
	Name     string `xml:"name"`
	ImageURL string `xml:"image_url"`
	Link     string `xml:"link"`
}

// Update represents an update element in the GoodReads response
type Update struct {
	Type       string `xml:"type,attr"`
	ActionText string `xml:"action_text"`
	Actor      Actor  `xml:"actor"`
	Object     Object `xml:"object"`
	Updated    string `xml:"updated_at"`
}

// Object represents an object element in the GoodReads response
type Object struct {
	ReadStatus ReadStatus `xml:"read_status"`
}

// ReadStatus represents a read_status element in the GoodReads response
type ReadStatus struct {
	ID       string `xml:"id"`
	ReviewID string `xml:"review_id"`
	UserID   string `xml:"user_id"`
	Status   string `xml:"status"`
	Updated  string `xml:"updated_at"`
	Review   Review `xml:"review"`
}

// Shelf represents a shelf element in the GoodReads response
type Shelf struct {
	ID        string `xml:"id"`
	BookCount string `xml:"book_count"`
	Name      string `xml:"name"`
}

// UserStatus represents a user status element in the GoodReads response
type UserStatus struct {
	Page    int    `xml:"page"`
	Percent int    `xml:"percent"`
	Updated string `xml:"updated_at"`
	Book    Book   `xml:"book"`
}

// UpdatedRelative returns the relative date for a user status
func (u UserStatus) UpdatedRelative() string {
	return relativeDate(u.Updated)
}

// Book represents a book element in the GoodReads response
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

// Author returns the first author in the Book's author list
func (b Book) Author() Author {
	return b.Authors[0]
}

// Author represents an author element in the GoodReads response
type Author struct {
	ID                   string `xml:"id"`
	Name                 string `xml:"name"`
	Link                 string `xml:"link"`
	FansCount            int    `xml:"fans_count"`
	AuthorFollowersCount int    `xml:"author_followers_count"`
	LargeImageURL        string `xml:"large_image_url"`
	ImageURL             string `xml:"image_url"`
	SmallImageURL        string `xml:"small_image_url"`
	WorksCount           int    `xml:"works_count"`
	Gender               string `xml:"gender"`
	Hometown             string `xml:"hometown"`
}

// Review represents an review element in the GoodReads response
type Review struct {
	Book   Book   `xml:"book"`
	Rating int    `xml:"rating"`
	ReadAt string `xml:"read_at"`
	Link   string `xml:"link"`
}

// FullStars returns a list representing the number of rating stars
func (r Review) FullStars() []bool {
	return make([]bool, r.Rating)
}

// EmptyStars returns a list representing the number of "empty" rating stars
func (r Review) EmptyStars() []bool {
	return make([]bool, 5-r.Rating)
}

// ReadAtShort returns a short string representing the read_at date
func (r Review) ReadAtShort() string {
	date, err := parseDate(r.ReadAt)
	if err != nil {
		return ""
	}

	return (string)(date.Format("2 Jan 2006"))
}

// ReadAtRelative returns a short string representing the relative read_at date
func (r Review) ReadAtRelative() string {
	return relativeDate(r.ReadAt)
}
