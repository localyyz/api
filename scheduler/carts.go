package scheduler

import (
	"fmt"
	"time"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/connect"
	"bitbucket.org/moodie-app/moodie-api/lib/onesignal"
	"github.com/pressly/lg"
	db "upper.io/db.v3"
)

const (
	AbandonVariantLimit   = 5
	AbandonPushTemplateID = "6e76a7eb-3090-4579-bf05-bf53fbc9675c"
	AbandonPushOOSContent = "Hey {{ name | default: 'there'}}! %s is selling out fast, only %d left! Get yours before it goes out of stock.⏰"
	AbandonPushContent    = "Hey {{ name | default: 'there'}}! %s is selling out fast, Get yours before it goes out of stock.⏰"

	TouchIntervalMax        = "48 hour"
	AbandonTouchIntervalMin = 4 * time.Hour
)

func (h *Handler) AbandonCartHandler() {
	h.wg.Add(1)
	defer h.wg.Done()

	s := time.Now()
	lg.Info("abandon cart running...")
	defer func() {
		lg.Infof("abandon cart finished in %s", time.Since(s))
	}()

	// abandon cart handler pushes abandoned cart users.
	var carts []*data.Cart
	err := data.DB.Select(db.Raw("distinct on (c.id) c.*")).
		From("carts c").
		LeftJoin("cart_items ci").On("ci.cart_id = c.id").
		Where(db.Cond{
			"c.status": []data.CartStatus{
				data.CartStatusInProgress,
				data.CartStatusCheckout,
			}, // cart has not completed yet
			"ci.created_at": db.Between(db.Raw("now() - interval '4 hour'"), db.Raw("now()")), // cart item created within min interval.
			"ci.id":         db.IsNotNull(),                                                   // have at least one cart item
			"c.is_express":  false,
		}).
		All(&carts)
	if err != nil {
		lg.Alertf("failed to schedule abandon cart push: %v", err)
		return
	}

	if len(carts) == 0 {
		lg.Infof("nothing to do.")
		return
	}

	var toSend []data.Notification
	for _, cart := range carts {
		// for each cart, check if need to send.
		alreadySent, err := data.DB.Notification.Find(db.Cond{
			"cart_id":      cart.ID,
			"scheduled_at": db.Gt(db.Raw("now() - interval '48 hour'")), // have not touched in last 48h.
		}).Exists()
		if err != nil {
			lg.Infof("skipping cart(%d) exist errored %v. being safe this iteration", cart.ID, err)
			continue
		}
		if alreadySent {
			lg.Infof("skipping cart(%d) already touched within 48h", cart.ID)
			continue
		}

		var selected *data.ProductVariant
		cartItems, err := data.DB.CartItem.FindByCartID(cart.ID)
		if err != nil {
			lg.Warnf("failed to fetch cartItem on cart(%d): %v", cart.ID, err)
			continue
		}

		for i, ci := range cartItems {
			variant, err := data.DB.ProductVariant.FindByID(ci.VariantID)
			if err != nil {
				lg.Warnf("failed to fetch variant for cartItem(%d): %v", ci.VariantID, err)
				continue
			}

			// either chose the very first
			if i == 0 {
				selected = variant
			}

			// or find a variant that's almost out of stock
			if variant.Limits <= AbandonVariantLimit {
				selected = variant
				break
			}
		}

		// get the product title
		product, err := data.DB.Product.FindByID(selected.ProductID)
		if err != nil {
			lg.Warnf("failed to fetch product(%d): %v", selected.ProductID, err)
			continue
		}

		ntf := data.Notification{
			CartID:    &(cart.ID),
			UserID:    cart.UserID,
			ProductID: selected.ProductID,
			VariantID: &(selected.ID),
		}
		if selected.Limits <= AbandonVariantLimit {
			ntf.Heading = "Almost 🚀 gone!"
			ntf.Content = fmt.Sprintf(AbandonPushOOSContent, product.Title, selected.Limits)
		} else {
			ntf.Heading = "Hurry before its 🚀 gone!"
			ntf.Content = fmt.Sprintf(AbandonPushContent, product.Title)
		}
		toSend = append(toSend, ntf)
	}

	// for each cart, send notf (playerID -> content)
	for _, notf := range toSend {
		user, err := data.DB.User.FindByID(notf.UserID)
		if err != nil {
			lg.Warnf("failed to fetch user(%d): %v", notf.UserID, err)
			continue
		}
		req := onesignal.NotificationRequest{
			Headings:         map[string]string{"en": notf.Heading},
			Contents:         map[string]string{"en": notf.Content},
			IncludePlayerIDs: []string{user.Etc.OSPlayerID},
		}
		resp, _, err := connect.ON.Notifications.Create(&req)
		if err != nil {
			lg.Warnf("failed to schedule notification: %v", err)
			continue
		}

		notf.ExternalID = resp.ID
		if err := data.DB.Notification.Save(&notf); err != nil {
			lg.Warnf("failed to save notification to db: %v", err)
		}
	}
	lg.Infof("scheduled %d push notifications", len(toSend))
}
