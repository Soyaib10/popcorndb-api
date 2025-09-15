package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/Soyaib10/popcorndb-api/internal/validator"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Movie struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"-"`
	Title     string    `json:"title"`
	Year      int32     `json:"year,omitempty"`
	Runtime   int32     `json:"runtime,omitempty"`
	Genres    []string  `json:"genres,omitempty"`
	Version   int32     `json:"version"`
}

func ValidateMovie(v *validator.Validator, movie *Movie) {
	v.Check(movie.Title != "", "title", "must be provided")
	v.Check(len(movie.Title) <= 500, "title", "must not be more than 500 bytes long")
	v.Check(movie.Year != 0, "year", "must be provided")
	v.Check(movie.Year >= 1888, "year", "must be greater than 1888")
	v.Check(movie.Year <= int32(time.Now().Year()), "year", "must not be in the future")
	v.Check(movie.Runtime != 0, "runtime", "must be provided")
	v.Check(movie.Runtime > 0, "runtime", "must be a positive integer")
	v.Check(movie.Genres != nil, "genres", "must be provided")
	v.Check(len(movie.Genres) >= 1, "genres", "must contain at least 1 genre")
	v.Check(len(movie.Genres) <= 5, "genres", "must not contain more than 5 genres")
	v.Check(validator.Unique(movie.Genres), "genres", "must not contain duplicate values")
}

type MovieModel struct {
	DB *pgxpool.Pool
}

func (m MovieModel) Get(id int64) (*Movie, error) {
	/*
		- check for param err
		- create a context
		- write query
		- get data from db
		- bind data with destinaton struct
		- handle possible errors
	*/

	if id < 1 {
		return nil, ErrRecordNotFound
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
	SELECT id, created_at, title, year, runtime, genres, version
	FROM movies
	WHERE id = $1;
	`
	row := m.DB.QueryRow(ctx, query, id)

	var movie Movie
	err := row.Scan(
		&movie.ID,
		&movie.CreatedAt,
		&movie.Title,
		&movie.Year,
		&movie.Runtime,
		&movie.Genres,
		&movie.Version,
	)

	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &movie, nil
}

func (m MovieModel) GetAll(title string, genres []string, filters Filters) ([]*Movie, Metadata, error) {
	/*
		- create a context
		- write query
		- get data from db and close rows
		- bind data with destinaton struct
		- handle possible errors
	*/

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := fmt.Sprintf(`
		SELECT count(*) OVER(), id, created_at, title, year, runtime, genres, version
		FROM movies
		WHERE (to_tsvector('simple', title) @@ plainto_tsquery('simple', $1) OR $1 = '')
		AND (genres @> $2 OR $2 = '{}')
		ORDER BY %s %s, id ASC
		LIMIT $3 OFFSET $4`, filters.sortColumn(), filters.sortDirection())

	args := []interface{}{title, genres, filters.limit(), filters.offset()}

	rows, err := m.DB.Query(ctx, query, args...)
	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()


	movies := []*Movie{}
	totalRecords := 0
	for rows.Next() {
		var movie Movie
		err := rows.Scan(
			&totalRecords,
			&movie.ID,
			&movie.CreatedAt,
			&movie.Title,
			&movie.Year,
			&movie.Runtime,
			&movie.Genres,
			&movie.Version,
		)
		if err != nil {
			return nil, Metadata{}, err
		}
		movies = append(movies, &movie)
	}

	if rows.Err() != nil {
		return nil, Metadata{}, rows.Err()
	}

	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)

	return  movies, metadata, nil
}

func (m MovieModel) Insert(movie *Movie) error {
	/*
		- create a context
		- prepare arguments from struct
		- execute query and return inserted row
		- bind returned values into destination struct
		- handle possible errors
	*/

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		INSERT INTO movies (title, year, runtime, genres)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, version
		`
	args := []interface{}{movie.Title, movie.Year, movie.Runtime, movie.Genres}
	row := m.DB.QueryRow(ctx, query, args...)

	err := row.Scan(
		&movie.ID,
		&movie.CreatedAt,
		&movie.Version,
	)
	if err != nil {
		return err
	}

	return nil
}

func (m MovieModel) Update(movie *Movie) error {
	/*
		- create a context
		- write query
		- prepare arguments from struct
		- execute the query (with RETURN maybe)
		- bind returned values into destination struct
		- handle possible errors
	*/

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		UPDATE movies
		SET title = $1, year = $2, runtime = $3, genres = $4, version = version + 1
		WHERE id = $5 AND version = $6
		RETURNING version
	`
	args := []interface{}{
		movie.Title,
		movie.Year,
		movie.Runtime,
		movie.Genres,
		movie.ID,
		movie.Version,
	}
	row := m.DB.QueryRow(ctx, query, args...)

	err := row.Scan(&movie.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}
	return nil
}

func (m MovieModel) Delete(id int64) error {
	/*
		- check param
		- create a context
		- write query
		- execute the query using Exec()
		- check how many rows were affected
		- if no rows affected, return ErrRecordNotFound
		- otherwise, return nil
	*/

	if id < 1 {
		return ErrRecordNotFound
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		DELETE FROM movies
		WHERE id = $1
		`
	result, err := m.DB.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrRecordNotFound
	}
	return nil
}
