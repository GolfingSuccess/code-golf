package routes

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func ideas(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	type idea struct {
		ID, ThumbsDown, ThumbsUp int
		Title                    string
	}

	rows, err := db.Query("SELECT * FROM ideas ORDER BY thumbs_up - thumbs_down DESC, title")
	if err != nil {
		panic(err)
	}

	defer rows.Close()

	var ideas []idea

	for rows.Next() {
		var i idea

		if err := rows.Scan(&i.ID, &i.ThumbsDown, &i.ThumbsUp, &i.Title); err != nil {
			panic(err)
		}

		ideas = append(ideas, i)
	}

	Render(w, r, http.StatusOK, "ideas", "Ideas", ideas)
}
