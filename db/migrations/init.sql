--
-- PostgreSQL database dump
--

-- Dumped from database version 10.4 (Debian 10.4-2.pgdg90+1)
-- Dumped by pg_dump version 10.4 (Debian 10.4-2.pgdg90+1)

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
-- Name: pg_trgm; Type: EXTENSION; Schema: -; Owner: 
--

CREATE EXTENSION IF NOT EXISTS pg_trgm WITH SCHEMA public;


--
-- Name: EXTENSION pg_trgm; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION pg_trgm IS 'text similarity measurement and index searching based on trigrams';


--
-- Name: product_tsv_trigger(); Type: FUNCTION; Schema: public; Owner: localyyz
--

CREATE FUNCTION public.product_tsv_trigger() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
DECLARE
  name text;
begin
  select places.name into name from places where id = new.place_id;
  new.tsv :=
  setweight(to_tsvector(
    COALESCE(
      CASE WHEN new.gender = 1 THEN 'man'
          WHEN new.gender = 2 THEN 'woman'
      END, '')), 'A') ||
  setweight(to_tsvector(COALESCE(new.title,'')), 'A') ||
  setweight(to_tsvector(COALESCE(new.category->>'type','')), 'A') ||
  setweight(to_tsvector(COALESCE(new.category->>'value','')), 'A') ||
  setweight(to_tsvector(COALESCE(new.brand,'')), 'A') ||
  setweight(to_tsvector('simple', name), 'A');
  return new;
end
$$;


ALTER FUNCTION public.product_tsv_trigger() OWNER TO localyyz;

SET default_tablespace = '';

SET default_with_oids = false;

--
-- Name: billing_plans; Type: TABLE; Schema: public; Owner: localyyz
--

CREATE TABLE public.billing_plans (
    id integer NOT NULL,
    plan_type smallint DEFAULT 0 NOT NULL,
    billing_type smallint DEFAULT 0 NOT NULL,
    name character varying(512) DEFAULT ''::character varying NOT NULL,
    is_default boolean DEFAULT false NOT NULL,
    terms character varying(512) DEFAULT ''::character varying NOT NULL,
    recurring_price numeric(15,6) DEFAULT 0.0 NOT NULL,
    transaction_fee smallint DEFAULT 0 NOT NULL,
    commission_fee smallint DEFAULT 0 NOT NULL,
    other_fee smallint DEFAULT 0 NOT NULL
);


ALTER TABLE public.billing_plans OWNER TO localyyz;

--
-- Name: billing_plans_id_seq; Type: SEQUENCE; Schema: public; Owner: localyyz
--

CREATE SEQUENCE public.billing_plans_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.billing_plans_id_seq OWNER TO localyyz;

--
-- Name: billing_plans_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: localyyz
--

ALTER SEQUENCE public.billing_plans_id_seq OWNED BY public.billing_plans.id;


--
-- Name: blacklist; Type: TABLE; Schema: public; Owner: localyyz
--

CREATE TABLE public.blacklist (
    word character varying(256) NOT NULL
);


ALTER TABLE public.blacklist OWNER TO localyyz;

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
    deleted_at timestamp without time zone,
    is_express boolean DEFAULT false NOT NULL
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
-- Name: collection_products; Type: TABLE; Schema: public; Owner: localyyz
--

CREATE TABLE public.collection_products (
    collection_id bigint NOT NULL,
    product_id bigint NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.collection_products OWNER TO localyyz;

--
-- Name: collections; Type: TABLE; Schema: public; Owner: localyyz
--

CREATE TABLE public.collections (
    id integer NOT NULL,
    name character varying(512) DEFAULT ''::character varying NOT NULL,
    description text DEFAULT ''::text NOT NULL,
    image_url text DEFAULT ''::text NOT NULL,
    ordering smallint DEFAULT 1 NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    updated_at timestamp without time zone,
    gender smallint DEFAULT 0 NOT NULL,
    featured boolean DEFAULT false NOT NULL,
    place_ids bigint[],
    categories character varying(64)[]
);


ALTER TABLE public.collections OWNER TO localyyz;

--
-- Name: collections_id_seq; Type: SEQUENCE; Schema: public; Owner: localyyz
--

CREATE SEQUENCE public.collections_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.collections_id_seq OWNER TO localyyz;

--
-- Name: collections_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: localyyz
--

ALTER SEQUENCE public.collections_id_seq OWNED BY public.collections.id;


--
-- Name: feature_products; Type: TABLE; Schema: public; Owner: localyyz
--

CREATE TABLE public.feature_products (
    product_id bigint NOT NULL,
    ordering smallint DEFAULT 1 NOT NULL,
    image_url text DEFAULT ''::text NOT NULL,
    featured_at timestamp without time zone DEFAULT now() NOT NULL,
    end_featured_at timestamp without time zone
);


ALTER TABLE public.feature_products OWNER TO localyyz;

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
-- Name: merchant_approvals; Type: TABLE; Schema: public; Owner: localyyz
--

CREATE TABLE public.merchant_approvals (
    id integer NOT NULL,
    place_id bigint NOT NULL,
    collection smallint DEFAULT 0 NOT NULL,
    category smallint DEFAULT 0 NOT NULL,
    price_range smallint DEFAULT 0 NOT NULL,
    rejection_reason text,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    updated_at timestamp without time zone,
    approved_at timestamp without time zone,
    rejected_at timestamp without time zone
);


ALTER TABLE public.merchant_approvals OWNER TO localyyz;

--
-- Name: merchant_approvals_id_seq; Type: SEQUENCE; Schema: public; Owner: localyyz
--

CREATE SEQUENCE public.merchant_approvals_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.merchant_approvals_id_seq OWNER TO localyyz;

--
-- Name: merchant_approvals_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: localyyz
--

ALTER SEQUENCE public.merchant_approvals_id_seq OWNED BY public.merchant_approvals.id;


--
-- Name: place_billings; Type: TABLE; Schema: public; Owner: localyyz
--

CREATE TABLE public.place_billings (
    id integer NOT NULL,
    place_id bigint NOT NULL,
    plan_id bigint NOT NULL,
    status smallint DEFAULT 0 NOT NULL,
    external_id bigint NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    updated_at timestamp without time zone,
    accepted_at timestamp without time zone
);


ALTER TABLE public.place_billings OWNER TO localyyz;

--
-- Name: place_billings_id_seq; Type: SEQUENCE; Schema: public; Owner: localyyz
--

CREATE SEQUENCE public.place_billings_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.place_billings_id_seq OWNER TO localyyz;

--
-- Name: place_billings_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: localyyz
--

ALTER SEQUENCE public.place_billings_id_seq OWNED BY public.place_billings.id;


--
-- Name: place_charges; Type: TABLE; Schema: public; Owner: localyyz
--

CREATE TABLE public.place_charges (
    id integer NOT NULL,
    place_id bigint NOT NULL,
    external_id bigint NOT NULL,
    charge_type smallint DEFAULT 0 NOT NULL,
    amount numeric(15,6) DEFAULT 0.0 NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    expire_at timestamp without time zone NOT NULL
);


ALTER TABLE public.place_charges OWNER TO localyyz;

--
-- Name: place_charges_id_seq; Type: SEQUENCE; Schema: public; Owner: localyyz
--

CREATE SEQUENCE public.place_charges_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.place_charges_id_seq OWNER TO localyyz;

--
-- Name: place_charges_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: localyyz
--

ALTER SEQUENCE public.place_charges_id_seq OWNED BY public.place_charges.id;


--
-- Name: place_discounts; Type: TABLE; Schema: public; Owner: localyyz
--

CREATE TABLE public.place_discounts (
    id integer NOT NULL,
    place_id bigint NOT NULL,
    code character varying(255) NOT NULL,
    external_id bigint NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.place_discounts OWNER TO localyyz;

--
-- Name: place_discounts_id_seq; Type: SEQUENCE; Schema: public; Owner: localyyz
--

CREATE SEQUENCE public.place_discounts_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.place_discounts_id_seq OWNER TO localyyz;

--
-- Name: place_discounts_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: localyyz
--

ALTER SEQUENCE public.place_discounts_id_seq OWNED BY public.place_discounts.id;


--
-- Name: places; Type: TABLE; Schema: public; Owner: localyyz
--

CREATE TABLE public.places (
    id integer NOT NULL,
    name text DEFAULT ''::text NOT NULL,
    address text DEFAULT ''::text NOT NULL,
    phone text DEFAULT ''::text NOT NULL,
    website text DEFAULT ''::text NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    image_url text,
    shopify_id text DEFAULT ''::text NOT NULL,
    tos_agreed_at timestamp without time zone,
    approved_at timestamp without time zone,
    status smallint DEFAULT 0 NOT NULL,
    tos_ip character varying(64) DEFAULT ''::character varying NOT NULL,
    currency character varying(8) DEFAULT ''::character varying NOT NULL,
    gender smallint DEFAULT 3 NOT NULL,
    weight smallint DEFAULT 0 NOT NULL,
    payment_methods jsonb DEFAULT '[]'::jsonb NOT NULL,
    shopify_plan character varying(64) DEFAULT ''::character varying NOT NULL,
    fb_url text DEFAULT ''::text NOT NULL,
    instagram_url text DEFAULT ''::text NOT NULL,
    shipping_policy jsonb DEFAULT '{}'::jsonb NOT NULL,
    return_policy jsonb DEFAULT '{}'::jsonb NOT NULL,
    ratings jsonb DEFAULT '{}'::jsonb NOT NULL,
    is_used boolean DEFAULT false NOT NULL,
    is_dropshipper boolean DEFAULT false NOT NULL,
    plan_enabled boolean DEFAULT false NOT NULL
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
-- Name: priority_merchants; Type: TABLE; Schema: public; Owner: localyyz
--

CREATE TABLE public.priority_merchants (
    place_id integer NOT NULL
);


ALTER TABLE public.priority_merchants OWNER TO localyyz;

--
-- Name: product_categories; Type: TABLE; Schema: public; Owner: localyyz
--

CREATE TABLE public.product_categories (
    value character varying(64) DEFAULT ''::character varying NOT NULL,
    type smallint DEFAULT 0 NOT NULL,
    mapping character varying(64) DEFAULT ''::character varying NOT NULL,
    gender smallint DEFAULT 0 NOT NULL,
    weight smallint DEFAULT 0 NOT NULL
);


ALTER TABLE public.product_categories OWNER TO localyyz;

--
-- Name: product_images; Type: TABLE; Schema: public; Owner: localyyz
--

CREATE TABLE public.product_images (
    id integer NOT NULL,
    product_id bigint NOT NULL,
    external_id bigint NOT NULL,
    image_url text DEFAULT ''::text NOT NULL,
    ordering smallint DEFAULT 1 NOT NULL,
    width integer,
    height integer,
    score smallint DEFAULT '-1'::integer
);


ALTER TABLE public.product_images OWNER TO localyyz;

--
-- Name: product_images_id_seq; Type: SEQUENCE; Schema: public; Owner: localyyz
--

CREATE SEQUENCE public.product_images_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.product_images_id_seq OWNER TO localyyz;

--
-- Name: product_images_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: localyyz
--

ALTER SEQUENCE public.product_images_id_seq OWNED BY public.product_images.id;


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
    product_id bigint NOT NULL,
    price numeric(15,6) DEFAULT 0 NOT NULL,
    prev_price numeric(15,6) DEFAULT 0 NOT NULL
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
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    updated_at timestamp without time zone,
    deleted_at timestamp without time zone,
    external_id bigint,
    gender smallint DEFAULT 0 NOT NULL,
    category jsonb DEFAULT '{}'::jsonb NOT NULL,
    tsv tsvector,
    brand character varying(256) DEFAULT ''::character varying NOT NULL,
    status smallint DEFAULT 0 NOT NULL,
    score smallint DEFAULT '-1'::integer
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
-- Name: search_words; Type: MATERIALIZED VIEW; Schema: public; Owner: localyyz
--

CREATE MATERIALIZED VIEW public.search_words AS
 SELECT ts_stat.word,
    ts_stat.ndoc
   FROM ts_stat('SELECT to_tsvector(''simple'', title) FROM products'::text) ts_stat(word, ndoc, nentry)
  WHERE ((ts_stat.ndoc > 1) AND (to_tsvector(ts_stat.word) <> ''::tsvector))
  WITH NO DATA;


ALTER TABLE public.search_words OWNER TO localyyz;

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
-- Name: variant_images_pivot; Type: TABLE; Schema: public; Owner: localyyz
--

CREATE TABLE public.variant_images_pivot (
    variant_id bigint NOT NULL,
    image_id bigint NOT NULL
);


ALTER TABLE public.variant_images_pivot OWNER TO localyyz;

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
-- Name: billing_plans id; Type: DEFAULT; Schema: public; Owner: localyyz
--

ALTER TABLE ONLY public.billing_plans ALTER COLUMN id SET DEFAULT nextval('public.billing_plans_id_seq'::regclass);


--
-- Name: cart_items id; Type: DEFAULT; Schema: public; Owner: localyyz
--

ALTER TABLE ONLY public.cart_items ALTER COLUMN id SET DEFAULT nextval('public.cart_items_id_seq'::regclass);


--
-- Name: carts id; Type: DEFAULT; Schema: public; Owner: localyyz
--

ALTER TABLE ONLY public.carts ALTER COLUMN id SET DEFAULT nextval('public.carts_id_seq'::regclass);


--
-- Name: collections id; Type: DEFAULT; Schema: public; Owner: localyyz
--

ALTER TABLE ONLY public.collections ALTER COLUMN id SET DEFAULT nextval('public.collections_id_seq'::regclass);


--
-- Name: goose_db_version id; Type: DEFAULT; Schema: public; Owner: localyyz
--

ALTER TABLE ONLY public.goose_db_version ALTER COLUMN id SET DEFAULT nextval('public.goose_db_version_id_seq'::regclass);


--
-- Name: merchant_approvals id; Type: DEFAULT; Schema: public; Owner: localyyz
--

ALTER TABLE ONLY public.merchant_approvals ALTER COLUMN id SET DEFAULT nextval('public.merchant_approvals_id_seq'::regclass);


--
-- Name: place_billings id; Type: DEFAULT; Schema: public; Owner: localyyz
--

ALTER TABLE ONLY public.place_billings ALTER COLUMN id SET DEFAULT nextval('public.place_billings_id_seq'::regclass);


--
-- Name: place_charges id; Type: DEFAULT; Schema: public; Owner: localyyz
--

ALTER TABLE ONLY public.place_charges ALTER COLUMN id SET DEFAULT nextval('public.place_charges_id_seq'::regclass);


--
-- Name: place_discounts id; Type: DEFAULT; Schema: public; Owner: localyyz
--

ALTER TABLE ONLY public.place_discounts ALTER COLUMN id SET DEFAULT nextval('public.place_discounts_id_seq'::regclass);


--
-- Name: places id; Type: DEFAULT; Schema: public; Owner: localyyz
--

ALTER TABLE ONLY public.places ALTER COLUMN id SET DEFAULT nextval('public.places_id_seq'::regclass);


--
-- Name: product_images id; Type: DEFAULT; Schema: public; Owner: localyyz
--

ALTER TABLE ONLY public.product_images ALTER COLUMN id SET DEFAULT nextval('public.product_images_id_seq'::regclass);


--
-- Name: product_variants id; Type: DEFAULT; Schema: public; Owner: localyyz
--

ALTER TABLE ONLY public.product_variants ALTER COLUMN id SET DEFAULT nextval('public.promos_id_seq'::regclass);


--
-- Name: products id; Type: DEFAULT; Schema: public; Owner: localyyz
--

ALTER TABLE ONLY public.products ALTER COLUMN id SET DEFAULT nextval('public.products_id_seq'::regclass);


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
-- Name: billing_plans billing_plans_pkey; Type: CONSTRAINT; Schema: public; Owner: localyyz
--

ALTER TABLE ONLY public.billing_plans
    ADD CONSTRAINT billing_plans_pkey PRIMARY KEY (id);


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
-- Name: collections collections_pkey; Type: CONSTRAINT; Schema: public; Owner: localyyz
--

ALTER TABLE ONLY public.collections
    ADD CONSTRAINT collections_pkey PRIMARY KEY (id);


--
-- Name: goose_db_version goose_db_version_pkey; Type: CONSTRAINT; Schema: public; Owner: localyyz
--

ALTER TABLE ONLY public.goose_db_version
    ADD CONSTRAINT goose_db_version_pkey PRIMARY KEY (id);


--
-- Name: merchant_approvals merchant_approvals_pkey; Type: CONSTRAINT; Schema: public; Owner: localyyz
--

ALTER TABLE ONLY public.merchant_approvals
    ADD CONSTRAINT merchant_approvals_pkey PRIMARY KEY (id);


--
-- Name: place_billings place_billings_pkey; Type: CONSTRAINT; Schema: public; Owner: localyyz
--

ALTER TABLE ONLY public.place_billings
    ADD CONSTRAINT place_billings_pkey PRIMARY KEY (id);


--
-- Name: place_charges place_charges_pkey; Type: CONSTRAINT; Schema: public; Owner: localyyz
--

ALTER TABLE ONLY public.place_charges
    ADD CONSTRAINT place_charges_pkey PRIMARY KEY (id);


--
-- Name: place_discounts place_discounts_pkey; Type: CONSTRAINT; Schema: public; Owner: localyyz
--

ALTER TABLE ONLY public.place_discounts
    ADD CONSTRAINT place_discounts_pkey PRIMARY KEY (id);


--
-- Name: places places_pkey; Type: CONSTRAINT; Schema: public; Owner: localyyz
--

ALTER TABLE ONLY public.places
    ADD CONSTRAINT places_pkey PRIMARY KEY (id);


--
-- Name: product_images product_images_pkey; Type: CONSTRAINT; Schema: public; Owner: localyyz
--

ALTER TABLE ONLY public.product_images
    ADD CONSTRAINT product_images_pkey PRIMARY KEY (id);


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
-- Name: collection_products unique_collection_product; Type: CONSTRAINT; Schema: public; Owner: localyyz
--

ALTER TABLE ONLY public.collection_products
    ADD CONSTRAINT unique_collection_product UNIQUE (collection_id, product_id);


--
-- Name: place_discounts unique_place_discount; Type: CONSTRAINT; Schema: public; Owner: localyyz
--

ALTER TABLE ONLY public.place_discounts
    ADD CONSTRAINT unique_place_discount UNIQUE (place_id, external_id);


--
-- Name: product_categories unique_product_category_value; Type: CONSTRAINT; Schema: public; Owner: localyyz
--

ALTER TABLE ONLY public.product_categories
    ADD CONSTRAINT unique_product_category_value UNIQUE (value);


--
-- Name: products unique_product_place_external_id; Type: CONSTRAINT; Schema: public; Owner: localyyz
--

ALTER TABLE ONLY public.products
    ADD CONSTRAINT unique_product_place_external_id UNIQUE (place_id, external_id);


--
-- Name: users unique_username; Type: CONSTRAINT; Schema: public; Owner: localyyz
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT unique_username UNIQUE (username);


--
-- Name: variant_images_pivot unique_variant_image; Type: CONSTRAINT; Schema: public; Owner: localyyz
--

ALTER TABLE ONLY public.variant_images_pivot
    ADD CONSTRAINT unique_variant_image UNIQUE (variant_id, image_id);


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
-- Name: collection_products_collection_id_idx; Type: INDEX; Schema: public; Owner: localyyz
--

CREATE INDEX collection_products_collection_id_idx ON public.collection_products USING btree (collection_id);


--
-- Name: priority_merchant_index; Type: INDEX; Schema: public; Owner: localyyz
--

CREATE INDEX priority_merchant_index ON public.priority_merchants USING btree (place_id);


--
-- Name: product_categories_mapping_idx; Type: INDEX; Schema: public; Owner: localyyz
--

CREATE INDEX product_categories_mapping_idx ON public.product_categories USING btree (mapping);


--
-- Name: product_gender_idx; Type: INDEX; Schema: public; Owner: localyyz
--

CREATE INDEX product_gender_idx ON public.products USING btree (gender);


--
-- Name: product_images_product_id_idx; Type: INDEX; Schema: public; Owner: localyyz
--

CREATE INDEX product_images_product_id_idx ON public.product_images USING btree (product_id);


--
-- Name: product_status_index; Type: INDEX; Schema: public; Owner: localyyz
--

CREATE INDEX product_status_index ON public.products USING btree (status);


--
-- Name: product_variants_offer_id_idx; Type: INDEX; Schema: public; Owner: localyyz
--

CREATE INDEX product_variants_offer_id_idx ON public.product_variants USING btree (offer_id);


--
-- Name: product_variants_place_id_idx; Type: INDEX; Schema: public; Owner: localyyz
--

CREATE INDEX product_variants_place_id_idx ON public.product_variants USING btree (place_id);


--
-- Name: product_variants_product_id_idx; Type: INDEX; Schema: public; Owner: localyyz
--

CREATE INDEX product_variants_product_id_idx ON public.product_variants USING btree (product_id);


--
-- Name: products_category_ginidx; Type: INDEX; Schema: public; Owner: localyyz
--

CREATE INDEX products_category_ginidx ON public.products USING gin (category);


--
-- Name: products_category_type_idx; Type: INDEX; Schema: public; Owner: localyyz
--

CREATE INDEX products_category_type_idx ON public.products USING btree (((category ->> 'type'::text)));


--
-- Name: products_category_value_idx; Type: INDEX; Schema: public; Owner: localyyz
--

CREATE INDEX products_category_value_idx ON public.products USING btree (((category ->> 'value'::text)));


--
-- Name: products_created_at_idx; Type: INDEX; Schema: public; Owner: localyyz
--

CREATE INDEX products_created_at_idx ON public.products USING btree (created_at);


--
-- Name: products_place_id_idx; Type: INDEX; Schema: public; Owner: localyyz
--

CREATE INDEX products_place_id_idx ON public.products USING btree (place_id);


--
-- Name: tsv_idx; Type: INDEX; Schema: public; Owner: localyyz
--

CREATE INDEX tsv_idx ON public.products USING gin (tsv);


--
-- Name: unique_default_plan_type; Type: INDEX; Schema: public; Owner: localyyz
--

CREATE UNIQUE INDEX unique_default_plan_type ON public.billing_plans USING btree (plan_type) WHERE (is_default = true);


--
-- Name: unique_place_active_billing; Type: INDEX; Schema: public; Owner: localyyz
--

CREATE UNIQUE INDEX unique_place_active_billing ON public.place_billings USING btree (place_id, plan_id, status) WHERE (status = 3);


--
-- Name: variant_images_variant_id_idx; Type: INDEX; Schema: public; Owner: localyyz
--

CREATE INDEX variant_images_variant_id_idx ON public.variant_images_pivot USING btree (variant_id);


--
-- Name: words_idx; Type: INDEX; Schema: public; Owner: localyyz
--

CREATE INDEX words_idx ON public.search_words USING gin (word public.gin_trgm_ops);


--
-- Name: products tsvectorupdate; Type: TRIGGER; Schema: public; Owner: localyyz
--

CREATE TRIGGER tsvectorupdate BEFORE INSERT OR UPDATE ON public.products FOR EACH ROW EXECUTE PROCEDURE public.product_tsv_trigger();


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
-- Name: collection_products collection_products_collection_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: localyyz
--

ALTER TABLE ONLY public.collection_products
    ADD CONSTRAINT collection_products_collection_id_fkey FOREIGN KEY (collection_id) REFERENCES public.collections(id) ON DELETE CASCADE;


--
-- Name: collection_products collection_products_product_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: localyyz
--

ALTER TABLE ONLY public.collection_products
    ADD CONSTRAINT collection_products_product_id_fkey FOREIGN KEY (product_id) REFERENCES public.products(id) ON DELETE CASCADE;


--
-- Name: feature_products feature_products_product_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: localyyz
--

ALTER TABLE ONLY public.feature_products
    ADD CONSTRAINT feature_products_product_id_fkey FOREIGN KEY (product_id) REFERENCES public.products(id) ON DELETE CASCADE;


--
-- Name: merchant_approvals merchant_approvals_place_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: localyyz
--

ALTER TABLE ONLY public.merchant_approvals
    ADD CONSTRAINT merchant_approvals_place_id_fkey FOREIGN KEY (place_id) REFERENCES public.places(id) ON DELETE CASCADE;


--
-- Name: place_billings place_billings_place_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: localyyz
--

ALTER TABLE ONLY public.place_billings
    ADD CONSTRAINT place_billings_place_id_fkey FOREIGN KEY (place_id) REFERENCES public.places(id) ON DELETE CASCADE;


--
-- Name: place_billings place_billings_plan_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: localyyz
--

ALTER TABLE ONLY public.place_billings
    ADD CONSTRAINT place_billings_plan_id_fkey FOREIGN KEY (plan_id) REFERENCES public.billing_plans(id);


--
-- Name: place_charges place_charges_place_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: localyyz
--

ALTER TABLE ONLY public.place_charges
    ADD CONSTRAINT place_charges_place_id_fkey FOREIGN KEY (place_id) REFERENCES public.places(id) ON DELETE CASCADE;


--
-- Name: place_discounts place_discounts_place_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: localyyz
--

ALTER TABLE ONLY public.place_discounts
    ADD CONSTRAINT place_discounts_place_id_fkey FOREIGN KEY (place_id) REFERENCES public.places(id) ON DELETE CASCADE;


--
-- Name: priority_merchants priority_merchants_place_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: localyyz
--

ALTER TABLE ONLY public.priority_merchants
    ADD CONSTRAINT priority_merchants_place_id_fkey FOREIGN KEY (place_id) REFERENCES public.places(id) ON DELETE CASCADE;


--
-- Name: product_images product_images_product_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: localyyz
--

ALTER TABLE ONLY public.product_images
    ADD CONSTRAINT product_images_product_id_fkey FOREIGN KEY (product_id) REFERENCES public.products(id) ON DELETE CASCADE;


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
-- Name: variant_images_pivot variant_images_pivot_image_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: localyyz
--

ALTER TABLE ONLY public.variant_images_pivot
    ADD CONSTRAINT variant_images_pivot_image_id_fkey FOREIGN KEY (image_id) REFERENCES public.product_images(id) ON DELETE CASCADE;


--
-- Name: variant_images_pivot variant_images_pivot_variant_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: localyyz
--

ALTER TABLE ONLY public.variant_images_pivot
    ADD CONSTRAINT variant_images_pivot_variant_id_fkey FOREIGN KEY (variant_id) REFERENCES public.product_variants(id) ON DELETE CASCADE;


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

