package main

import (
	"fmt"
	"log"
	"net/url"
	"strings"

	"bitbucket.org/moodie-app/moodie-api/data"
	db "upper.io/db.v3"
)

func parseCollections() {
	d := []string{
		"https://www.designerrevival.com/products/red-chanel-pleated-silk-blouse?variant=37261228033",
		"https://www.designerrevival.com/products/navy-blue-marc-by-marc-jacobs-faux-fur-trimmed-trench-coat?variant=38592628289",
		"https://www.designerrevival.com/collections/recently-added-tops/products/cream-anveglosa-leather-peplum-top?variant=38615086273",
		"https://www.designerrevival.com/collections/recently-added-shoes/products/black-manolo-blahnik-mary-jane-pumps?variant=4101564071977",
		"https://www.designerrevival.com/collections/recently-added-shoes/products/black-louis-vuitton-leather-ankle-boots?variant=121357107201",
		"https://www.designerrevival.com/products/teal-black-blue-duck-fur-vest?variant=136303476737",
		"https://www.designerrevival.com/collections/recently-added-tops/products/black-tom-ford-long-sleeve-blouse?variant=158334517249",
		"https://www.designerrevival.com/products/grey-pamela-dennis-silk-velvet-gown?variant=4113579540521",
		"https://www.designerrevival.com/collections/recently-added-dresses/products/emerald-green-alice-olivia-sequin-mini-dress?variant=4113591140393",
		"https://www.designerrevival.com/collections/recently-added-tops/products/black-chanel-pleated-lace-blouse?variant=4117005664297",
		"https://www.designerrevival.com/collections/recently-added-jackets/products/brown-miu-miu-check-wool-coat?variant=4117263089705",
		"https://www.designerrevival.com/collections/accessories/products/goldtone-vintage-chanel-drop-earrings?variant=4116973092905",
		"https://www.designerrevival.com/collections/accessories/products/goldtone-vintage-chanel-clip-on-earrings?variant=4116963786793",
		"https://www.designerrevival.com/collections/accessories/products/goldtone-vintage-chanel-chain-necklace?variant=4116962770985",
		"https://www.designerrevival.com/collections/accessories/products/goldtone-vintage-hermes-kyoto-cuff-bracelet?variant=4116980760617",
		"https://www.designerrevival.com/collections/recently-added-tops/products/pale-pink-vintage-chanel-faux-wrap-blouse?variant=4124699066409",
		"https://www.designerrevival.com/collections/recently-added-dresses/products/silver-michael-kors-embellished-dress?variant=36587301633",
		"https://www.fashionest.com/products/austin-chic-foldover-clutch#search",
		"https://www.fashionest.com/collections/earrings/products/adorlee-drops",
		"https://www.fashionest.com/products/dallas-bracelet#search",
		"https://www.fashionest.com/products/dallas-necklace#search",
		"https://www.fashionest.com/products/hymn-to-selene-door-knocker-earrings#search",
		"https://www.fashionest.com/collections/color-multicolor/products/ria-condiment-trinket-dishes",
		"https://jiacollection.com/collections/cocktail-dresses-skirts/products/gia",
		"https://jiacollection.com/products/marcella",
		"https://jiacollection.com/collections/coats/products/margaret-brown",
		"https://www.madisonstyle.com/collections/oxfords-loafers/products/black-white-block-heel-loafer",
		"https://www.madisonstyle.com/collections/oxfords-loafers/products/black-wingtip-sirio-lace-up-oxfords",
		"https://www.madisonstyle.com/collections/bottoms/products/navy-100-wool-blend-cropped-tailored-trousers",
		"https://www.madisonstyle.com/collections/bottoms/products/black-raina-spandex-polyester-trousers",
		"https://www.madisonstyle.com/products/gold-calf-leather-point-heel-pump",
		"https://www.madisonstyle.com/collections/boots/products/black-over-the-knee-stretch-boot",
		"https://www.madisonstyle.com/collections/boots/products/gold-bronze-leather-ankle-boot",
		"https://www.madisonstyle.com/collections/tops/products/white-100-cotton-ba0108-flori-short-shirt",
		"https://www.madisonstyle.com/collections/scarves-wraps/products/grey-90-modal-latika-stole-scarf",
		"https://www.madisonstyle.com/collections/scarves-wraps/products/light-grey-rabbit-fur-tierra-collar",
		"https://www.madisonstyle.com/collections/belts/products/black-3999-suede-belt",
		"https://www.madisonstyle.com/collections/belts/products/grey-3476-leather-belt-scamosciato",
		"https://www.madisonstyle.com/collections/hats/products/black-wool-classic-long-brim-hat-1",
		"https://www.madisonstyle.com/collections/bags/products/black-rabbit-fur-soft-mobile-bag",
		"https://www.madisonstyle.com/collections/heels/products/gold-metallic-leather-bari-ballerina-shoes",
		"https://www.pulsedesignerfashion.com/products/armani-collezioni-womens-coat-sml03t-sm600-999",
		"https://www.pulsedesignerfashion.com/products/armani-collezioni-womens-hat-697222-4a901-00020",
		"https://www.pulsedesignerfashion.com/products/balmain-paris-womens-ankle-boot-s6c-bh-350206m-west-176",
		"https://www.pulsedesignerfashion.com/products/dolce-gabbana-ladies-fingerless-jewel-glove-fig34k-f49cr-s8434",
		"https://www.pulsedesignerfashion.com/products/v-1969-italia-womens-jumpsuit-vanessa",
		"https://www.pulsedesignerfashion.com/products/sebastian-milano-ladies-ankle-boot-4037-black-los-angeles-nero",
		"https://www.pulsedesignerfashion.com/products/sebastian-milano-ladies-high-boot-s4588-vacchetta-nappa-nero",
		"https://www.pulsedesignerfashion.com/products/v-1969-italia-womens-handbag-v003-s-ruga-rosso",
		"https://www.pulsedesignerfashion.com/products/v-1969-italia-womens-handbag-v1969007-red",
		"https://www.pulsedesignerfashion.com/products/v-1969-italia-womens-handbag-10",
		"https://www.pulsedesignerfashion.com/products/v-1969-italia-womens-high-boot-g051x-camoscio-nero",
		"https://www.pulsedesignerfashion.com/products/v-1969-italia-womens-jacket-long-sleeves-black-giulia",
		"https://www.pulsedesignerfashion.com/products/v-1969-italia-womens-pump-3105124-velluto-391-bordeaux",
		"https://quinnshop.com/collections/womens-sweaters/products/lucille-cold-shoulder-mock-neck",
		"https://quinnshop.com/collections/womens-sweaters/products/thirtyfour-slash-neck-bell-sleeve",
		"https://quinnshop.com/collections/womens-sweaters/products/teresa-cashmere-turtleneck-tunic",
		"https://quinnshop.com/collections/womens-dresses/products/qw63612",
		"https://quinnshop.com/collections/womens-dresses/products/beacon-fitted-knit-jumper",
		"https://quinnshop.com/collections/womens-dresses/products/sherman-dolman-sleeve-dress",
		"https://quinnshop.com/collections/womens-dresses/products/winslow-half-sleeve-paneled-body-con-dress",
		"https://quinnshop.com/collections/womens-dresses/products/grant-paneled-spagetti-strap-dress",
		"https://quinnshop.com/collections/womens-dresses/products/chloe-paneled-mock-neck-dress",
		"https://tusk.com/products/cross-body-bag-with-chain?variant=34644665030",
		"https://tusk.com/products/nepal-top-zip-cross-body-bag-1?variant=33895684358",
		"https://tusk.com/products/nepal-accordion-clutch-wallet-blue?variant=299850137606",
		"https://tusk.com/products/nepal-cross-body-bag-with-chain-red?variant=34644551494",
		"https://untitledandco.com/collections/dresses/products/flo-mini-suede-in-chocolate",
		"https://vampsnyc.com/products/steve-madden-daisie-women-rose-gold?variant=54082421767",
		"https://vampsnyc.com/collections/pumps/products/jessica-simpson-pelanna-women-black?variant=55610613191",
	}

	c, _ := data.DB.Collection.FindOne(db.Cond{"id": 1})

	var notFound int
	for _, u := range d {
		uu, _ := url.Parse(u)

		pathParts := strings.Split(uu.Path, "/")
		trailPath := pathParts[len(pathParts)-1]
		fmt.Println(trailPath)

		product, err := data.DB.Product.FindOne(db.Cond{"external_handle": trailPath})
		if err != nil {
			log.Printf("%s not found", trailPath)
			notFound++
			continue
		}

		data.DB.CollectionProduct.Save(&data.CollectionProduct{CollectionID: c.ID, ProductID: product.ID})
	}
	log.Printf("total %d not found %d", len(d), notFound)

}
