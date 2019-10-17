package routes

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"
)

func user(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	userID := 0
	data := struct {
		Holes          []Hole
		Langs          []Lang
		Login          string
		Points         int
		Ranks          map[string]map[string]int
		Trophies       []Trophy
		TrophiesEarned map[string]*time.Time
	}{
		Holes:          holes,
		Langs:          langs,
		Login:          ps[0].Value,
		Ranks:          map[string]map[string]int{},
		Trophies:       trophies,
		TrophiesEarned: map[string]*time.Time{},
	}

	if err := db.QueryRow(
		"SELECT id FROM users WHERE login = $1",
		data.Login,
	).Scan(&userID); err == sql.ErrNoRows {
		Render(w, r, http.StatusNotFound, "404", "", nil)
		return
	} else if err != nil {
		panic(err)
	}

	if err := db.QueryRow(
		"SELECT points FROM points WHERE user_id = $1",
		userID,
	).Scan(&data.Points); err != nil && err != sql.ErrNoRows {
		panic(err)
	}

	rows, err := db.Query(
		"SELECT earned, trophy FROM trophies WHERE user_id = $1",
		userID,
	)
	if err != nil {
		panic(err)
	}

	defer rows.Close()

	for rows.Next() {
		var earned time.Time
		var trophy string

		if err := rows.Scan(&earned, &trophy); err != nil {
			panic(err)
		}

		data.TrophiesEarned[trophy] = &earned
	}

	if err := rows.Err(); err != nil {
		panic(err)
	}

	rows, err = db.Query(
		`WITH matrix AS (
		  SELECT user_id,
		         hole,
		         lang,
		         RANK() OVER (PARTITION BY hole, lang ORDER BY LENGTH(code))
		    FROM solutions
		   WHERE NOT failing
		) SELECT hole, lang, rank
		    FROM matrix
		   WHERE user_id = $1`,
		userID,
	)
	if err != nil {
		panic(err)
	}

	defer rows.Close()

	for rows.Next() {
		var hole, lang string
		var rank int

		if err := rows.Scan(&hole, &lang, &rank); err != nil {
			panic(err)
		}

		if data.Ranks[hole] == nil {
			data.Ranks[hole] = map[string]int{}
		}

		data.Ranks[hole][lang] = rank
	}

	if err := rows.Err(); err != nil {
		panic(err)
	}

	Render(w, r, http.StatusOK, "user", data.Login, data)
}
