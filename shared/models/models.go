package models

import "time"

type User struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	CreatedAt time.Time `json:"created_at"`
}

type Movie struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Year        int      `json:"year"`
	Rating      float64  `json:"rating"`
	PosterURL   string   `json:"poster_url"`
	Cast        []string `json:"cast"`
	MagnetLinks []string `json:"magnet_links"`
}

type Comment struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	MovieID   string    `json:"movie_id"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
