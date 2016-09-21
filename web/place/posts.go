package place

import (
	"net/http"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/presenter"
	"bitbucket.org/moodie-app/moodie-api/lib/ws"
	"upper.io/db.v2"
)

func CreatePost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := ctx.Value("session.user").(*data.User)
	place := ctx.Value("place").(*data.Place)

	var payload struct {
		data.Post

		// Ignore
		ID          interface{} `json:"id,omitempty"`
		PlaceID     interface{} `jsoN:"placeId,omitempty"`
		UserID      interface{} `json:"userId,omitempty"`
		Comments    interface{} `json:"comments,omitempty"`
		Likes       interface{} `json:"comments,omitempty"`
		PromoStatus interface{} `json:"promoStatus,omitempty"`
		CreatedAt   interface{} `json:"createdAt,omitempty"`
		UpdatedAt   interface{} `json:"updatedAt,omitempty"`
	}
	if err := ws.Bind(r.Body, &payload); err != nil {
		ws.Respond(w, http.StatusBadRequest, err)
		return
	}

	newPost := &payload.Post
	newPost.UserID = user.ID
	newPost.PlaceID = place.ID

	newPost.PlaceID = place.ID
	if err := data.DB.Post.Save(newPost); err != nil {
		ws.Respond(w, http.StatusInternalServerError, err)
		return
	}

	ws.Respond(w, http.StatusCreated, newPost)
}

func GetPosts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	place := ctx.Value("place").(*data.Place)

	cursor := ws.NewPage(r)
	q := data.DB.Post.
		Find(db.Cond{"place_id": place.ID}).
		OrderBy("-score")
	q = cursor.UpdateQueryUpper(q)
	var posts []*data.Post
	if err := q.All(&posts); err != nil {
		ws.Respond(w, http.StatusInternalServerError, err)
		return
	}

	presented := make([]*presenter.Post, len(posts))
	for i, p := range posts {
		presented[i] = presenter.NewPost(p).WithUser()
	}
	ws.Respond(w, http.StatusOK, presented, cursor.Update(presented))
}