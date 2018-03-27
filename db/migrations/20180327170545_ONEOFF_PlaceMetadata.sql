
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
UPDATE places
    SET fb_url = 'https://www.facebook.com/bikinidotcom',
        instagram_url = 'https://www.instagram.com/bikinidotcom/',
        shipping_policy = '{"url":"https://www.bikini.com/shipping-costs-methods","desc":"Ships out immediately, Free standard US delivery"}',
        return_policy = '{"url":"https://www.bikini.com/shipping-costs-methods","desc":"Returns accepted for refund on most products"}',
        ratings = '{"rating":4.60,"count":50}'
WHERE id = 678;

UPDATE places
    SET fb_url = 'https://www.facebook.com/shopthefinest/',
        instagram_url = 'https://www.instagram.com/shopthefinest/',
        shipping_policy = '{"url":"https://www.shopthefinest.com/pages/shipping/","desc":"Ships out immediately, Standard US delivery"}',
        return_policy = '{"url":"https://www.shopthefinest.com/pages/return-policy/","desc":"Returns accepted for store credit or exchange"}',
        ratings = '{"rating":4.20,"count":30}'
WHERE id = 51;

UPDATE places
    SET fb_url = 'https://www.facebook.com/ThePocketSquareIndustry',
        instagram_url = 'https://www.instagram.com/sebastiancruzcouture/',
        shipping_policy = '{"url":"Free Shipping Worldwide","desc":""}',
        return_policy = '{"url":"https://www.sebastiancruzcouture.com/pages/policies","desc":"Returns accepted for full refund or exchange"}',
        ratings = '{}'
WHERE id = 90;

UPDATE places
    SET fb_url = 'https://www.facebook.com/eksterwallets',
        instagram_url = 'https://www.instagram.com/eksterwallets/',
        shipping_policy = '{"url":"Free Shipping Worldwide","desc":""}',
        return_policy = '{"url":"https://ekster.com/pages/return","desc":"Returns accepted for full refund or exchange"}',
        ratings = '{}'
WHERE id = 12;

UPDATE places
    SET fb_url = 'https://www.facebook.com/indiraactive/',
        instagram_url = 'https://www.instagram.com/indira_active/',
        shipping_policy = '{"url":"Free for two items or more (US), $150+ (Canada), $250+ (International)","desc":""}',
        return_policy = '{"url":"https://returns.indiraactive.com/?_ga=2.160959477.936440416.1522171164-1067094825.1522171164","desc":"Returns accepted for full refund or exchange"}',
        ratings = '{}'
WHERE id = 1448;

UPDATE places
    SET fb_url = 'https://www.facebook.com/Suxalys',
        instagram_url = 'https://www.instagram.com/suxalys/',
        shipping_policy = '{"url":"https://www.suxalys.com/pages/shipping-info","desc":"Variable shipping cost and time"}',
        return_policy = '{"url":"https://www.suxalys.com/pages/return-policy","desc":"This store does not accept returns"}',
        ratings = '{}'
WHERE id = 1671;

UPDATE places
    SET fb_url = 'https://www.facebook.com/skinyummies',
        instagram_url = 'https://www.instagram.com/sallybskinyummies/#',
        shipping_policy = '{"url":"https://www.sallybskinyummies.com/pages/the-fine-print-we-hope-you-never-have-to-read","desc":"Ships out immediately, Free Shipping over $75 (US)"}',
        return_policy = '{"url":"https://www.sallybskinyummies.com/pages/the-fine-print-we-hope-you-never-have-to-read","desc":"Returns accepted for refund or store credit"}',
        ratings = '{"rating":4.20,"count":358}'
WHERE id = 228;

UPDATE places
    SET fb_url = 'https://www.facebook.com/ARubyLLC',
        instagram_url = 'https://www.instagram.com/arubystyle/',
        shipping_policy = '{"url":"https://aruby.net/pages/shipping","desc":"Ships out immediately, Free standard US delivery over $100"}',
        return_policy = '{"url":"https://aruby.net/pages/returns-and-exchanges","desc":"Returns accepted for refund or exchange"}',
        ratings = '{}'
WHERE id = 259;

UPDATE places
    SET fb_url = 'https://www.facebook.com/peternappistudio',
        instagram_url = 'https://www.instagram.com/peternappi/',
        shipping_policy = '{"url":"https://peternappi.com/pages/faq","desc":"Ships out immediately, Free standard US delivery over $300"}',
        return_policy = '{"url":"https://peternappi.com/pages/faq","desc":"Returns accepted for full refund or exchange"}',
        ratings = '{"rating":4.90,"count":70}'
WHERE id = 133;

UPDATE places
    SET fb_url = 'https://www.facebook.com/CladinStores/',
        instagram_url = 'https://www.instagram.com/cladinpvd/',
        shipping_policy = '{"url":"https://cladin.com/pages/shipping","desc":"Ships out immediately, Standard US delivery"}',
        return_policy = '{"url":"https://cladin.com/pages/returns","desc":"Returns accepted for refund or exchange on most products (US ONLY)"}',
        ratings = '{"rating":5.00,"count":1}'
WHERE id = 222;

UPDATE places
    SET fb_url = 'https://www.facebook.com/bocnewyork',
        instagram_url = 'https://www.instagram.com/bocnyc/',
        shipping_policy = '{"url":"https://bocnyc.com/pages/shipping","desc":"Ships out immediately, Standard US delivery"}',
        return_policy = '{"url":"https://bocnyc.com/pages/returns","desc":"Returns accepted for refund or exchange on most products"}',
        ratings = '{}'
WHERE id = 156;

UPDATE places
    SET fb_url = 'https://www.facebook.com/birdiesslippers',
        instagram_url = 'https://www.instagram.com/birdies/',
        shipping_policy = '{"url":"","desc":"Free standard US delivery"}',
        return_policy = '{"url":"https://birdiesslippers.returnly.com/","desc":"Returns accepted for full refund or exchange"}',
        ratings = '{}'
WHERE id = 163;

UPDATE places
    SET fb_url = 'https://www.facebook.com/befashionest/?fref=ts',
        instagram_url = 'https://www.instagram.com/befashionest/',
        shipping_policy = '{"url":"https://www.fashionest.com/pages/shipping-handling","desc":"Standard US delivery"}',
        return_policy = '{"url":"https://www.fashionest.com/pages/shipping-handling","desc":"Returns accepted for full refund or exchange"}',
        ratings = '{"rating":4.70,"count":14}'
WHERE id = 44;

UPDATE places
    SET fb_url = 'https://www.facebook.com/MadisonStyleBH/',
        instagram_url = 'https://www.instagram.com/madisonstyle/',
        shipping_policy = '{"url":"https://www.madisonstyle.com/pages/shipping-los-angeles-beverly-hills-brentwood","desc":"Ships out immediately, Free standard US delivery on $200+"}',
        return_policy = '{"url":"https://www.madisonstyle.com/pages/shipping-los-angeles-beverly-hills-brentwood","desc":"Returns accepted for refund or exchange"}',
        ratings = '{"rating":4.30,"count":15}'
WHERE id = 50;

UPDATE places
    SET fb_url = 'https://www.facebook.com/StudioGearCosmetics/',
        instagram_url = 'https://www.instagram.com/studiogearcosmetics/',
        shipping_policy = '{"url":"","desc":"Variable shipping cost and time"}',
        return_policy = '{"url":"https://studiogearcosmetics.com/pages/product-returns","desc":"Returns accepted for refund or exchange"}',
        ratings = '{"rating":4.20,"count":187}'
WHERE id = 571;

UPDATE places
    SET fb_url = 'https://www.facebook.com/LITBOUTIQUE/',
        instagram_url = 'https://www.instagram.com/LITBoutique/',
        shipping_policy = '{"url":"","desc":"Variable shipping cost and time"}',
        return_policy = '{"url":"https://litboutique.com/pages/terms-conditions","desc":"Returns accepted for refund or exchange"}',
        ratings = '{}'
WHERE id = 55;

UPDATE places
    SET fb_url = 'https://www.facebook.com/OndadeMarOdM/',
        instagram_url = 'https://www.instagram.com/ondademar/',
        shipping_policy = '{"url":"https://ondademar.com/pages/help#ship-delivery","desc":"Ships out immediately, Free standard continental US delivery on $180+"}',
        return_policy = '{"url":"https://ondademar.com/pages/help#return-p","desc":"Returns accepted for exchange or store credit"}',
        ratings = '{}'
WHERE id = 117;

UPDATE places
    SET fb_url = 'https://www.facebook.com/conditionculture',
        instagram_url = 'https://www.instagram.com/conditionculture/',
        shipping_policy = '{"url":"","desc":"Variable shipping cost and time"}',
        return_policy = '{"url":"https://conditionculture.com/pages/policies-terms-and-conditions","desc":"Returns accepted for refund or exchange"}',
        ratings = '{}'
WHERE id = 112;

UPDATE places
    SET fb_url = '-',
        instagram_url = 'https://www.instagram.com/vampsnyc/',
        shipping_policy = '{"url":"https://vampsnyc.com/pages/shipping-returns","desc":"Ships out immediately, Free standard continental US delivery over $50"}',
        return_policy = '{"url":"https://vampsnyc.com/pages/shipping-returns","desc":"Returns accepted for refund or exchange"}',
        ratings = '{}'
WHERE id = 14;

UPDATE places
    SET fb_url = 'https://www.facebook.com/designerrevival',
        instagram_url = 'https://www.instagram.com/designerrevival/',
        shipping_policy = '{"url":"","desc":"Ships out immediately, Free standard US delivery"}',
        return_policy = '{"url":"https://www.designerrevival.com/pages/return-policy","desc":"This store does not accept returns"}',
        ratings = '{"rating":4.30,"count":13}'
WHERE id = 46;

UPDATE places
    SET fb_url = 'https://www.facebook.com/shopmaude/',
        instagram_url = 'https://www.instagram.com/shopmaude/',
        shipping_policy = '{"url":"https://shopmaude.com/pages/shipping","desc":"Ships out immediately, Free standard US delivery"}',
        return_policy = '{"url":"https://shopmaude.com/pages/returns","desc":"Returns accepted for exchange or store credit"}',
        ratings = '{"rating":4.20,"count":334}'
WHERE id = 122;

UPDATE places
    SET fb_url = 'https://www.facebook.com/alchemydetroit',
        instagram_url = 'https://www.instagram.com/alchemy.detroit/',
        shipping_policy = '{"url":"https://alchemydetroit.com/pages/help-page","desc":"Ships out immediately, Free standard US delivery"}',
        return_policy = '{"url":"https://alchemydetroit.com/pages/help-page","desc":"Returns accepted for full refund or exchange"}',
        ratings = '{"rating":5.00,"count":7}'
WHERE id = 173;

UPDATE places
    SET fb_url = 'https://www.facebook.com/TuskNYC',
        instagram_url = 'https://www.instagram.com/tusknyc/',
        shipping_policy = '{"url":"https://tusk.com/pages/shipping-options","desc":"Ships out immediately, Free standard continental US delivery over $75"}',
        return_policy = '{"url":"https://tusk.com/pages/product-customer-care","desc":"Returns accepted for refund or exchange"}',
        ratings = '{"rating":4.30,"count":25}'
WHERE id = 47;

UPDATE places
    SET fb_url = 'https://www.facebook.com/SpongelleBeyondCleansing',
        instagram_url = 'https://www.instagram.com/spongellebeyondcleansing/',
        shipping_policy = '{"url":"https://spongelle.com/pages/shipping-terms","desc":"Ships out immediately, Free standard continental US delivery over $75"}',
        return_policy = '{"url":"https://spongelle.com/pages/return-policy","desc":"Returns accepted for refund or exchange"}',
        ratings = '{"rating":4.90,"count":87}'
WHERE id = 736;


-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

