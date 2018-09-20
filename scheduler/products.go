package scheduler

import (
	"fmt"
	"math/rand"
	"time"

	"bitbucket.org/moodie-app/moodie-api/data"
	"github.com/pressly/lg"
	db "upper.io/db.v3"
)

const (
	FavouriteProductLimit    = 5
	FavouriteProductInterval = 1 * time.Hour
)

type favNotification struct {
	Product *data.Product
	UserID  int64
	Heading string
	Content string
}

func (h *Handler) FavouriteProductHandler() {
	h.wg.Add(1)
	defer h.wg.Done()

	s := time.Now()
	lg.Info("fav product running...")
	defer func() {
		lg.Infof("fav product finished in %s", time.Since(s))
	}()

	favs, err := data.DB.FavouriteProduct.
		FindAll(db.Cond{
			// send a notification to people who favourited an item 4h ago.
			"created_at": db.Between(
				db.Raw("NOW() - interval '8 hour'"),
				db.Raw("NOW() - interval '4 hour'"),
			),
		})
	if err != nil {
		lg.Alertf("failed to schedule fav product push: %v", err)
		return
	}

	if len(favs) == 0 {
		lg.Infof("nothing to do.")
		return
	}

	userFavs := map[int64][]*data.FavouriteProduct{}
	for _, f := range favs {
		userFavs[f.UserID] = append(userFavs[f.UserID], f)
	}

	var toSend []data.Notification
	for userID, uf := range userFavs {
		// TODO: send only for products with low inventory count
		// check, if we've sent a notification in the last.. 48h
		alreadySent, err := data.DB.Notification.Find(db.Cond{
			"user_id":      userID,
			"scheduled_at": db.Gt(db.Raw("now() - interval '4 hour'")), // have not touched in last 4h.
		}).Exists()
		if err != nil {
			lg.Infof("skipping fav push for user(%d) exist errored %v. being safe this iteration", userID, err)
			continue
		}
		if alreadySent {
			lg.Infof("skipping fav push for user(%d) already touched within 4h", userID)
			continue
		}

		selected := uf[0]

		// get the product title
		product, err := data.DB.Product.FindByID(selected.ProductID)
		if err != nil {
			lg.Warnf("failed to fetch product(%d): %v", selected.ProductID, err)
			continue
		}

		ntf := data.Notification{
			UserID:    userID,
			ProductID: selected.ProductID,
			Heading:   "Nice find!",
			Content:   fmt.Sprintf("%d people also added '%s' to their favorite list.", rand.Intn(100), product.Title),
		}
		toSend = append(toSend, ntf)
	}

	// for each, send notf (playerID -> content)
	for _, notf := range toSend {
		//user, err := data.DB.User.FindByID(notf.UserID)
		//if err != nil {
		//lg.Warnf("failed to fetch user(%d): %v", notf.UserID, err)
		//continue
		//}
		//req := onesignal.NotificationRequest{
		//Headings:         map[string]string{"en": notf.Heading},
		//Contents:         map[string]string{"en": notf.Content},
		//IncludePlayerIDs: []string{user.Etc.OSPlayerID},
		//}
		//resp, _, err := connect.ON.Notifications.Create(&req)
		//if err != nil {
		//lg.Warnf("failed to schedule notification: %v", err)
		//continue
		//}
		//notf.ExternalID = resp.ID
		if err := data.DB.Notification.Save(&notf); err != nil {
			lg.Warnf("failed to save notification to db: %v", err)
		}
		lg.Infof("scheduled push notifications with: %s to user(%d)", notf.Content, notf.UserID)
	}
	lg.Infof("scheduled %d push notifications", len(toSend))
}
