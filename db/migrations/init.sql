--
-- PostgreSQL database dump
--

-- Dumped from database version 10.3
-- Dumped by pg_dump version 10.3

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: plpgsql; Type: EXTENSION; Schema: -; Owner: 
--

CREATE EXTENSION IF NOT EXISTS plpgsql WITH SCHEMA pg_catalog;


--
-- Name: EXTENSION plpgsql; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION plpgsql IS 'PL/pgSQL procedural language';


--
-- Name: postgis; Type: EXTENSION; Schema: -; Owner: 
--

CREATE EXTENSION IF NOT EXISTS postgis WITH SCHEMA public;


--
-- Name: EXTENSION postgis; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION postgis IS 'PostGIS geometry, geography, and raster spatial types and functions';


SET default_tablespace = '';

SET default_with_oids = false;

--
-- Name: cart_items; Type: TABLE; Schema: public; Owner: localyyz
--

CREATE TABLE public.cart_items (
    id integer NOT NULL,
    cart_id bigint NOT NULL,
    product_id bigint NOT NULL,
    variant_id bigint NOT NULL,
    quantity smallint DEFAULT 0 NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    updated_at timestamp without time zone,
    deleted_at timestamp without time zone,
    place_id bigint
);


ALTER TABLE public.cart_items OWNER TO localyyz;

--
-- Name: cart_items_id_seq; Type: SEQUENCE; Schema: public; Owner: localyyz
--

CREATE SEQUENCE public.cart_items_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.cart_items_id_seq OWNER TO localyyz;

--
-- Name: cart_items_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: localyyz
--

ALTER SEQUENCE public.cart_items_id_seq OWNED BY public.cart_items.id;


--
-- Name: carts; Type: TABLE; Schema: public; Owner: localyyz
--

CREATE TABLE public.carts (
    id integer NOT NULL,
    user_id bigint NOT NULL,
    status smallint DEFAULT 0 NOT NULL,
    etc jsonb DEFAULT '{}'::jsonb NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    updated_at timestamp without time zone,
    deleted_at timestamp without time zone
);


ALTER TABLE public.carts OWNER TO localyyz;

--
-- Name: carts_id_seq; Type: SEQUENCE; Schema: public; Owner: localyyz
--

CREATE SEQUENCE public.carts_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.carts_id_seq OWNER TO localyyz;

--
-- Name: carts_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: localyyz
--

ALTER SEQUENCE public.carts_id_seq OWNED BY public.carts.id;


--
-- Name: cells; Type: TABLE; Schema: public; Owner: localyyz
--

CREATE TABLE public.cells (
    id integer NOT NULL,
    cell_id bigint NOT NULL,
    locale_id bigint
);


ALTER TABLE public.cells OWNER TO localyyz;

--
-- Name: cells_id_seq; Type: SEQUENCE; Schema: public; Owner: localyyz
--

CREATE SEQUENCE public.cells_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.cells_id_seq OWNER TO localyyz;

--
-- Name: cells_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: localyyz
--

ALTER SEQUENCE public.cells_id_seq OWNED BY public.cells.id;


--
-- Name: followings; Type: TABLE; Schema: public; Owner: localyyz
--

CREATE TABLE public.followings (
    id integer NOT NULL,
    user_id bigint NOT NULL,
    place_id bigint NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.followings OWNER TO localyyz;

--
-- Name: followings_id_seq; Type: SEQUENCE; Schema: public; Owner: localyyz
--

CREATE SEQUENCE public.followings_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.followings_id_seq OWNER TO localyyz;

--
-- Name: followings_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: localyyz
--

ALTER SEQUENCE public.followings_id_seq OWNED BY public.followings.id;


--
-- Name: goose_db_version; Type: TABLE; Schema: public; Owner: localyyz
--

CREATE TABLE public.goose_db_version (
    id integer NOT NULL,
    version_id bigint NOT NULL,
    is_applied boolean NOT NULL,
    tstamp timestamp without time zone DEFAULT now()
);


ALTER TABLE public.goose_db_version OWNER TO localyyz;

--
-- Name: goose_db_version_id_seq; Type: SEQUENCE; Schema: public; Owner: localyyz
--

CREATE SEQUENCE public.goose_db_version_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.goose_db_version_id_seq OWNER TO localyyz;

--
-- Name: goose_db_version_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: localyyz
--

ALTER SEQUENCE public.goose_db_version_id_seq OWNED BY public.goose_db_version.id;


--
-- Name: locales; Type: TABLE; Schema: public; Owner: localyyz
--

CREATE TABLE public.locales (
    id integer NOT NULL,
    name text DEFAULT ''::text NOT NULL,
    description text DEFAULT ''::text NOT NULL,
    shorthand text DEFAULT ''::text NOT NULL,
    image_url text DEFAULT ''::text NOT NULL,
    type smallint DEFAULT 0 NOT NULL
);


ALTER TABLE public.locales OWNER TO localyyz;

--
-- Name: locales_id_seq; Type: SEQUENCE; Schema: public; Owner: localyyz
--

CREATE SEQUENCE public.locales_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.locales_id_seq OWNER TO localyyz;

--
-- Name: locales_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: localyyz
--

ALTER SEQUENCE public.locales_id_seq OWNED BY public.locales.id;


--
-- Name: place_locales; Type: TABLE; Schema: public; Owner: localyyz
--

CREATE TABLE public.place_locales (
    place_id bigint NOT NULL,
    locale_id bigint NOT NULL
);


ALTER TABLE public.place_locales OWNER TO localyyz;

--
-- Name: places; Type: TABLE; Schema: public; Owner: localyyz
--

CREATE TABLE public.places (
    id integer NOT NULL,
    locale_id bigint,
    name text DEFAULT ''::text NOT NULL,
    address text DEFAULT ''::text NOT NULL,
    phone text DEFAULT ''::text NOT NULL,
    website text DEFAULT ''::text NOT NULL,
    description text DEFAULT ''::text NOT NULL,
    geo public.geography(Point,4326) DEFAULT public.st_geographyfromtext('SRID=4326;POINT(0 0)'::text),
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    image_url text,
    shopify_id text DEFAULT ''::text NOT NULL,
    tos_agreed_at timestamp without time zone,
    approved_at timestamp without time zone,
    status smallint DEFAULT 0 NOT NULL,
    tos_ip character varying(64) DEFAULT ''::character varying NOT NULL,
    billing jsonb DEFAULT '{}'::jsonb NOT NULL
);


ALTER TABLE public.places OWNER TO localyyz;

--
-- Name: places_id_seq; Type: SEQUENCE; Schema: public; Owner: localyyz
--

CREATE SEQUENCE public.places_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.places_id_seq OWNER TO localyyz;

--
-- Name: places_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: localyyz
--

ALTER SEQUENCE public.places_id_seq OWNED BY public.places.id;


--
-- Name: product_tags; Type: TABLE; Schema: public; Owner: localyyz
--

CREATE TABLE public.product_tags (
    id integer NOT NULL,
    product_id bigint NOT NULL,
    value character varying(128) DEFAULT ''::character varying NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    type smallint DEFAULT 0 NOT NULL,
    place_id bigint NOT NULL
);


ALTER TABLE public.product_tags OWNER TO localyyz;

--
-- Name: product_tags_id_seq; Type: SEQUENCE; Schema: public; Owner: localyyz
--

CREATE SEQUENCE public.product_tags_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.product_tags_id_seq OWNER TO localyyz;

--
-- Name: product_tags_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: localyyz
--

ALTER SEQUENCE public.product_tags_id_seq OWNED BY public.product_tags.id;


--
-- Name: product_variants; Type: TABLE; Schema: public; Owner: localyyz
--

CREATE TABLE public.product_variants (
    id integer NOT NULL,
    place_id bigint NOT NULL,
    limits bigint DEFAULT 0 NOT NULL,
    description text DEFAULT ''::text NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    etc jsonb DEFAULT '{}'::jsonb NOT NULL,
    deleted_at timestamp without time zone,
    updated_at timestamp without time zone,
    offer_id bigint DEFAULT 0 NOT NULL,
    product_id bigint NOT NULL
);


ALTER TABLE public.product_variants OWNER TO localyyz;

--
-- Name: products; Type: TABLE; Schema: public; Owner: localyyz
--

CREATE TABLE public.products (
    id integer NOT NULL,
    place_id bigint NOT NULL,
    external_handle text DEFAULT ''::text NOT NULL,
    title text DEFAULT ''::text NOT NULL,
    description text DEFAULT ''::text NOT NULL,
    image_url text DEFAULT ''::text NOT NULL,
    etc jsonb DEFAULT '{}'::jsonb NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    updated_at timestamp without time zone,
    deleted_at timestamp without time zone,
    external_id bigint
);


ALTER TABLE public.products OWNER TO localyyz;

--
-- Name: products_id_seq; Type: SEQUENCE; Schema: public; Owner: localyyz
--

CREATE SEQUENCE public.products_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.products_id_seq OWNER TO localyyz;

--
-- Name: products_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: localyyz
--

ALTER SEQUENCE public.products_id_seq OWNED BY public.products.id;


--
-- Name: promos_id_seq; Type: SEQUENCE; Schema: public; Owner: localyyz
--

CREATE SEQUENCE public.promos_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.promos_id_seq OWNER TO localyyz;

--
-- Name: promos_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: localyyz
--

ALTER SEQUENCE public.promos_id_seq OWNED BY public.product_variants.id;


--
-- Name: shares; Type: TABLE; Schema: public; Owner: localyyz
--

CREATE TABLE public.shares (
    id integer NOT NULL,
    user_id bigint NOT NULL,
    place_id bigint NOT NULL,
    network character varying(64) DEFAULT ''::character varying NOT NULL,
    network_share_id text DEFAULT ''::text NOT NULL,
    reach integer DEFAULT 0 NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    hash character varying(36) DEFAULT ''::character varying NOT NULL
);


ALTER TABLE public.shares OWNER TO localyyz;

--
-- Name: shares_id_seq; Type: SEQUENCE; Schema: public; Owner: localyyz
--

CREATE SEQUENCE public.shares_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.shares_id_seq OWNER TO localyyz;

--
-- Name: shares_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: localyyz
--

ALTER SEQUENCE public.shares_id_seq OWNED BY public.shares.id;


--
-- Name: shopify_creds; Type: TABLE; Schema: public; Owner: localyyz
--

CREATE TABLE public.shopify_creds (
    id integer NOT NULL,
    place_id bigint NOT NULL,
    auth_access_token text NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    updated_at timestamp without time zone,
    api_url text DEFAULT ''::text NOT NULL,
    status smallint DEFAULT 0 NOT NULL
);


ALTER TABLE public.shopify_creds OWNER TO localyyz;

--
-- Name: shopify_creds_id_seq; Type: SEQUENCE; Schema: public; Owner: localyyz
--

CREATE SEQUENCE public.shopify_creds_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.shopify_creds_id_seq OWNER TO localyyz;

--
-- Name: shopify_creds_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: localyyz
--

ALTER SEQUENCE public.shopify_creds_id_seq OWNED BY public.shopify_creds.id;


--
-- Name: user_addresses; Type: TABLE; Schema: public; Owner: localyyz
--

CREATE TABLE public.user_addresses (
    id integer NOT NULL,
    user_id bigint NOT NULL,
    is_shipping boolean DEFAULT false NOT NULL,
    is_billing boolean DEFAULT false NOT NULL,
    first_name character varying(128) DEFAULT ''::character varying NOT NULL,
    last_name character varying(128) DEFAULT ''::character varying NOT NULL,
    address character varying(256) DEFAULT ''::character varying NOT NULL,
    address_opt character varying(128) DEFAULT ''::character varying NOT NULL,
    city character varying(128) DEFAULT ''::character varying NOT NULL,
    country character varying(128) DEFAULT ''::character varying NOT NULL,
    province character varying(123) DEFAULT ''::character varying NOT NULL,
    zip character varying(64) DEFAULT ''::character varying NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    updated_at timestamp without time zone,
    deleted_at timestamp without time zone
);


ALTER TABLE public.user_addresses OWNER TO localyyz;

--
-- Name: user_addresses_id_seq; Type: SEQUENCE; Schema: public; Owner: localyyz
--

CREATE SEQUENCE public.user_addresses_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.user_addresses_id_seq OWNER TO localyyz;

--
-- Name: user_addresses_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: localyyz
--

ALTER SEQUENCE public.user_addresses_id_seq OWNED BY public.user_addresses.id;


--
-- Name: users; Type: TABLE; Schema: public; Owner: localyyz
--

CREATE TABLE public.users (
    id integer NOT NULL,
    username character varying(128) NOT NULL,
    email character varying(256) NOT NULL,
    name character varying(128) NOT NULL,
    avatar_url character varying(256) DEFAULT ''::character varying NOT NULL,
    network character varying(64) NOT NULL,
    access_token character varying(512) NOT NULL,
    geo public.geography(Point,4326) DEFAULT public.st_geographyfromtext('SRID=4326;POINT(0 0)'::text),
    etc jsonb DEFAULT '{}'::jsonb NOT NULL,
    logged_in boolean DEFAULT false NOT NULL,
    last_login_at timestamp without time zone,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    updated_at timestamp without time zone,
    deleted_at timestamp without time zone,
    device_token character varying(64),
    password_hash character varying(60) DEFAULT ''::character varying NOT NULL,
    invite_code character varying(8) DEFAULT ''::character varying NOT NULL,
    email_status smallint DEFAULT 0 NOT NULL
);


ALTER TABLE public.users OWNER TO localyyz;

--
-- Name: users_id_seq; Type: SEQUENCE; Schema: public; Owner: localyyz
--

CREATE SEQUENCE public.users_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.users_id_seq OWNER TO localyyz;

--
-- Name: users_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: localyyz
--

ALTER SEQUENCE public.users_id_seq OWNED BY public.users.id;


--
-- Name: webhooks; Type: TABLE; Schema: public; Owner: localyyz
--

CREATE TABLE public.webhooks (
    id integer NOT NULL,
    place_id bigint NOT NULL,
    topic text NOT NULL,
    external_id bigint NOT NULL,
    last_synced_at timestamp without time zone,
    created_at timestamp without time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.webhooks OWNER TO localyyz;

--
-- Name: webhooks_id_seq; Type: SEQUENCE; Schema: public; Owner: localyyz
--

CREATE SEQUENCE public.webhooks_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.webhooks_id_seq OWNER TO localyyz;

--
-- Name: webhooks_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: localyyz
--

ALTER SEQUENCE public.webhooks_id_seq OWNED BY public.webhooks.id;


--
-- Name: cart_items id; Type: DEFAULT; Schema: public; Owner: localyyz
--

ALTER TABLE ONLY public.cart_items ALTER COLUMN id SET DEFAULT nextval('public.cart_items_id_seq'::regclass);


--
-- Name: carts id; Type: DEFAULT; Schema: public; Owner: localyyz
--

ALTER TABLE ONLY public.carts ALTER COLUMN id SET DEFAULT nextval('public.carts_id_seq'::regclass);


--
-- Name: cells id; Type: DEFAULT; Schema: public; Owner: localyyz
--

ALTER TABLE ONLY public.cells ALTER COLUMN id SET DEFAULT nextval('public.cells_id_seq'::regclass);


--
-- Name: followings id; Type: DEFAULT; Schema: public; Owner: localyyz
--

ALTER TABLE ONLY public.followings ALTER COLUMN id SET DEFAULT nextval('public.followings_id_seq'::regclass);


--
-- Name: goose_db_version id; Type: DEFAULT; Schema: public; Owner: localyyz
--

ALTER TABLE ONLY public.goose_db_version ALTER COLUMN id SET DEFAULT nextval('public.goose_db_version_id_seq'::regclass);


--
-- Name: locales id; Type: DEFAULT; Schema: public; Owner: localyyz
--

ALTER TABLE ONLY public.locales ALTER COLUMN id SET DEFAULT nextval('public.locales_id_seq'::regclass);


--
-- Name: places id; Type: DEFAULT; Schema: public; Owner: localyyz
--

ALTER TABLE ONLY public.places ALTER COLUMN id SET DEFAULT nextval('public.places_id_seq'::regclass);


--
-- Name: product_tags id; Type: DEFAULT; Schema: public; Owner: localyyz
--

ALTER TABLE ONLY public.product_tags ALTER COLUMN id SET DEFAULT nextval('public.product_tags_id_seq'::regclass);


--
-- Name: product_variants id; Type: DEFAULT; Schema: public; Owner: localyyz
--

ALTER TABLE ONLY public.product_variants ALTER COLUMN id SET DEFAULT nextval('public.promos_id_seq'::regclass);


--
-- Name: products id; Type: DEFAULT; Schema: public; Owner: localyyz
--

ALTER TABLE ONLY public.products ALTER COLUMN id SET DEFAULT nextval('public.products_id_seq'::regclass);


--
-- Name: shares id; Type: DEFAULT; Schema: public; Owner: localyyz
--

ALTER TABLE ONLY public.shares ALTER COLUMN id SET DEFAULT nextval('public.shares_id_seq'::regclass);


--
-- Name: shopify_creds id; Type: DEFAULT; Schema: public; Owner: localyyz
--

ALTER TABLE ONLY public.shopify_creds ALTER COLUMN id SET DEFAULT nextval('public.shopify_creds_id_seq'::regclass);


--
-- Name: user_addresses id; Type: DEFAULT; Schema: public; Owner: localyyz
--

ALTER TABLE ONLY public.user_addresses ALTER COLUMN id SET DEFAULT nextval('public.user_addresses_id_seq'::regclass);


--
-- Name: users id; Type: DEFAULT; Schema: public; Owner: localyyz
--

ALTER TABLE ONLY public.users ALTER COLUMN id SET DEFAULT nextval('public.users_id_seq'::regclass);


--
-- Name: webhooks id; Type: DEFAULT; Schema: public; Owner: localyyz
--

ALTER TABLE ONLY public.webhooks ALTER COLUMN id SET DEFAULT nextval('public.webhooks_id_seq'::regclass);


--
-- Name: cart_items cart_items_pkey; Type: CONSTRAINT; Schema: public; Owner: localyyz
--

ALTER TABLE ONLY public.cart_items
    ADD CONSTRAINT cart_items_pkey PRIMARY KEY (id);


--
-- Name: carts carts_pkey; Type: CONSTRAINT; Schema: public; Owner: localyyz
--

ALTER TABLE ONLY public.carts
    ADD CONSTRAINT carts_pkey PRIMARY KEY (id);


--
-- Name: cells cells_pkey; Type: CONSTRAINT; Schema: public; Owner: localyyz
--

ALTER TABLE ONLY public.cells
    ADD CONSTRAINT cells_pkey PRIMARY KEY (id);


--
-- Name: followings followings_pkey; Type: CONSTRAINT; Schema: public; Owner: localyyz
--

ALTER TABLE ONLY public.followings
    ADD CONSTRAINT followings_pkey PRIMARY KEY (id);


--
-- Name: goose_db_version goose_db_version_pkey; Type: CONSTRAINT; Schema: public; Owner: localyyz
--

ALTER TABLE ONLY public.goose_db_version
    ADD CONSTRAINT goose_db_version_pkey PRIMARY KEY (id);


--
-- Name: locales locales_pkey; Type: CONSTRAINT; Schema: public; Owner: localyyz
--

ALTER TABLE ONLY public.locales
    ADD CONSTRAINT locales_pkey PRIMARY KEY (id);


--
-- Name: places places_pkey; Type: CONSTRAINT; Schema: public; Owner: localyyz
--

ALTER TABLE ONLY public.places
    ADD CONSTRAINT places_pkey PRIMARY KEY (id);


--
-- Name: product_tags product_tags_pkey; Type: CONSTRAINT; Schema: public; Owner: localyyz
--

ALTER TABLE ONLY public.product_tags
    ADD CONSTRAINT product_tags_pkey PRIMARY KEY (id);


--
-- Name: products products_pkey; Type: CONSTRAINT; Schema: public; Owner: localyyz
--

ALTER TABLE ONLY public.products
    ADD CONSTRAINT products_pkey PRIMARY KEY (id);


--
-- Name: product_variants promos_pkey; Type: CONSTRAINT; Schema: public; Owner: localyyz
--

ALTER TABLE ONLY public.product_variants
    ADD CONSTRAINT promos_pkey PRIMARY KEY (id);


--
-- Name: shares shares_pkey; Type: CONSTRAINT; Schema: public; Owner: localyyz
--

ALTER TABLE ONLY public.shares
    ADD CONSTRAINT shares_pkey PRIMARY KEY (id);


--
-- Name: shopify_creds shopify_creds_pkey; Type: CONSTRAINT; Schema: public; Owner: localyyz
--

ALTER TABLE ONLY public.shopify_creds
    ADD CONSTRAINT shopify_creds_pkey PRIMARY KEY (id);


--
-- Name: cart_items unique_cart_product_variant; Type: CONSTRAINT; Schema: public; Owner: localyyz
--

ALTER TABLE ONLY public.cart_items
    ADD CONSTRAINT unique_cart_product_variant UNIQUE (cart_id, product_id, variant_id);


--
-- Name: cells unique_cell_locale; Type: CONSTRAINT; Schema: public; Owner: localyyz
--

ALTER TABLE ONLY public.cells
    ADD CONSTRAINT unique_cell_locale UNIQUE (cell_id, locale_id);


--
-- Name: shares unique_hash; Type: CONSTRAINT; Schema: public; Owner: localyyz
--

ALTER TABLE ONLY public.shares
    ADD CONSTRAINT unique_hash UNIQUE (hash);


--
-- Name: place_locales unique_place_locale; Type: CONSTRAINT; Schema: public; Owner: localyyz
--

ALTER TABLE ONLY public.place_locales
    ADD CONSTRAINT unique_place_locale UNIQUE (place_id, locale_id);


--
-- Name: products unique_product_place_external_id; Type: CONSTRAINT; Schema: public; Owner: localyyz
--

ALTER TABLE ONLY public.products
    ADD CONSTRAINT unique_product_place_external_id UNIQUE (place_id, external_id);


--
-- Name: product_tags unique_product_value_type; Type: CONSTRAINT; Schema: public; Owner: localyyz
--

ALTER TABLE ONLY public.product_tags
    ADD CONSTRAINT unique_product_value_type UNIQUE (product_id, value, type);


--
-- Name: followings unique_user_following; Type: CONSTRAINT; Schema: public; Owner: localyyz
--

ALTER TABLE ONLY public.followings
    ADD CONSTRAINT unique_user_following UNIQUE (user_id, place_id);


--
-- Name: users unique_username; Type: CONSTRAINT; Schema: public; Owner: localyyz
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT unique_username UNIQUE (username);


--
-- Name: user_addresses user_addresses_pkey; Type: CONSTRAINT; Schema: public; Owner: localyyz
--

ALTER TABLE ONLY public.user_addresses
    ADD CONSTRAINT user_addresses_pkey PRIMARY KEY (id);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: localyyz
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- Name: webhooks webhooks_pkey; Type: CONSTRAINT; Schema: public; Owner: localyyz
--

ALTER TABLE ONLY public.webhooks
    ADD CONSTRAINT webhooks_pkey PRIMARY KEY (id);


--
-- Name: product_tags_place_id_idx; Type: INDEX; Schema: public; Owner: localyyz
--

CREATE INDEX product_tags_place_id_idx ON public.product_tags USING btree (place_id);


--
-- Name: product_variants_offer_id_idx; Type: INDEX; Schema: public; Owner: localyyz
--

CREATE INDEX product_variants_offer_id_idx ON public.product_variants USING btree (offer_id);


--
-- Name: product_variants_place_id_idx; Type: INDEX; Schema: public; Owner: localyyz
--

CREATE INDEX product_variants_place_id_idx ON public.product_variants USING btree (place_id);


--
-- Name: products_created_at_idx; Type: INDEX; Schema: public; Owner: localyyz
--

CREATE INDEX products_created_at_idx ON public.products USING btree (created_at);


--
-- Name: products_place_id_idx; Type: INDEX; Schema: public; Owner: localyyz
--

CREATE INDEX products_place_id_idx ON public.products USING btree (place_id);


--
-- Name: cart_items cart_items_cart_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: localyyz
--

ALTER TABLE ONLY public.cart_items
    ADD CONSTRAINT cart_items_cart_id_fkey FOREIGN KEY (cart_id) REFERENCES public.carts(id) ON DELETE CASCADE;


--
-- Name: cart_items cart_items_place_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: localyyz
--

ALTER TABLE ONLY public.cart_items
    ADD CONSTRAINT cart_items_place_id_fkey FOREIGN KEY (place_id) REFERENCES public.places(id) ON DELETE CASCADE;


--
-- Name: cart_items cart_items_product_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: localyyz
--

ALTER TABLE ONLY public.cart_items
    ADD CONSTRAINT cart_items_product_id_fkey FOREIGN KEY (product_id) REFERENCES public.products(id) ON DELETE CASCADE;


--
-- Name: cart_items cart_items_variant_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: localyyz
--

ALTER TABLE ONLY public.cart_items
    ADD CONSTRAINT cart_items_variant_id_fkey FOREIGN KEY (variant_id) REFERENCES public.product_variants(id) ON DELETE CASCADE;


--
-- Name: carts carts_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: localyyz
--

ALTER TABLE ONLY public.carts
    ADD CONSTRAINT carts_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: cells cells_locale_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: localyyz
--

ALTER TABLE ONLY public.cells
    ADD CONSTRAINT cells_locale_id_fkey FOREIGN KEY (locale_id) REFERENCES public.locales(id) ON DELETE CASCADE;


--
-- Name: followings followings_place_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: localyyz
--

ALTER TABLE ONLY public.followings
    ADD CONSTRAINT followings_place_id_fkey FOREIGN KEY (place_id) REFERENCES public.places(id) ON DELETE CASCADE;


--
-- Name: followings followings_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: localyyz
--

ALTER TABLE ONLY public.followings
    ADD CONSTRAINT followings_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: place_locales place_locales_locale_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: localyyz
--

ALTER TABLE ONLY public.place_locales
    ADD CONSTRAINT place_locales_locale_id_fkey FOREIGN KEY (locale_id) REFERENCES public.locales(id) ON DELETE CASCADE;


--
-- Name: place_locales place_locales_place_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: localyyz
--

ALTER TABLE ONLY public.place_locales
    ADD CONSTRAINT place_locales_place_id_fkey FOREIGN KEY (place_id) REFERENCES public.places(id) ON DELETE CASCADE;


--
-- Name: places places_locale_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: localyyz
--

ALTER TABLE ONLY public.places
    ADD CONSTRAINT places_locale_id_fkey FOREIGN KEY (locale_id) REFERENCES public.locales(id);


--
-- Name: product_tags product_tags_place_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: localyyz
--

ALTER TABLE ONLY public.product_tags
    ADD CONSTRAINT product_tags_place_id_fkey FOREIGN KEY (place_id) REFERENCES public.places(id) ON DELETE CASCADE;


--
-- Name: product_tags product_tags_product_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: localyyz
--

ALTER TABLE ONLY public.product_tags
    ADD CONSTRAINT product_tags_product_id_fkey FOREIGN KEY (product_id) REFERENCES public.products(id) ON DELETE CASCADE;


--
-- Name: products products_place_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: localyyz
--

ALTER TABLE ONLY public.products
    ADD CONSTRAINT products_place_id_fkey FOREIGN KEY (place_id) REFERENCES public.places(id) ON DELETE CASCADE;


--
-- Name: product_variants promos_place_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: localyyz
--

ALTER TABLE ONLY public.product_variants
    ADD CONSTRAINT promos_place_id_fkey FOREIGN KEY (place_id) REFERENCES public.places(id) ON DELETE CASCADE;


--
-- Name: product_variants promos_product_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: localyyz
--

ALTER TABLE ONLY public.product_variants
    ADD CONSTRAINT promos_product_id_fkey FOREIGN KEY (product_id) REFERENCES public.products(id) ON DELETE CASCADE;


--
-- Name: shares shares_place_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: localyyz
--

ALTER TABLE ONLY public.shares
    ADD CONSTRAINT shares_place_id_fkey FOREIGN KEY (place_id) REFERENCES public.places(id) ON DELETE CASCADE;


--
-- Name: shares shares_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: localyyz
--

ALTER TABLE ONLY public.shares
    ADD CONSTRAINT shares_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: shopify_creds shopify_creds_place_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: localyyz
--

ALTER TABLE ONLY public.shopify_creds
    ADD CONSTRAINT shopify_creds_place_id_fkey FOREIGN KEY (place_id) REFERENCES public.places(id) ON DELETE CASCADE;


--
-- Name: user_addresses user_addresses_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: localyyz
--

ALTER TABLE ONLY public.user_addresses
    ADD CONSTRAINT user_addresses_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: webhooks webhooks_place_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: localyyz
--

ALTER TABLE ONLY public.webhooks
    ADD CONSTRAINT webhooks_place_id_fkey FOREIGN KEY (place_id) REFERENCES public.places(id) ON DELETE CASCADE;


--
-- Name: DEFAULT PRIVILEGES FOR SEQUENCES; Type: DEFAULT ACL; Schema: public; Owner: postgres
--

ALTER DEFAULT PRIVILEGES FOR ROLE postgres IN SCHEMA public REVOKE ALL ON SEQUENCES  FROM postgres;
ALTER DEFAULT PRIVILEGES FOR ROLE postgres IN SCHEMA public GRANT ALL ON SEQUENCES  TO localyyz;


--
-- Name: DEFAULT PRIVILEGES FOR TABLES; Type: DEFAULT ACL; Schema: public; Owner: postgres
--

ALTER DEFAULT PRIVILEGES FOR ROLE postgres IN SCHEMA public REVOKE ALL ON TABLES  FROM postgres;
ALTER DEFAULT PRIVILEGES FOR ROLE postgres IN SCHEMA public GRANT ALL ON TABLES  TO localyyz;


--
-- PostgreSQL database dump complete
--

