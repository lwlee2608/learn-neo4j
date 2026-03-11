package domain

type Movie struct {
	Title    string `json:"title"`
	Released int    `json:"released"`
	Tagline  string `json:"tagline"`
}

type Person struct {
	Name string `json:"name"`
	Born int    `json:"born"`
}

type ActedIn struct {
	PersonName string `json:"person_name"`
	MovieTitle string `json:"movie_title"`
	Role       string `json:"role"`
}

type MovieWithCast struct {
	Movie Movie        `json:"movie"`
	Cast  []CastMember `json:"cast"`
}

type CastMember struct {
	Name string `json:"name"`
	Role string `json:"role"`
}
