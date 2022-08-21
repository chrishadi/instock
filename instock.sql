--
-- PostgreSQL database dump
--

--
-- Name: stocks; Type: TABLE; Schema: public
--

CREATE TABLE IF NOT EXISTS public.stocks (
    name character varying,
    code character varying,
    sub_sector_id integer,
    sub_sector_name character varying,
    sector_id integer,
    sector_name character varying,
    last numeric,
    prev_closing_price numeric,
    adjusted_open_price numeric,
    adjusted_high_price numeric,
    adjusted_low_price numeric,
    volume numeric,
    frequency numeric,
    value numeric,
    one_day numeric,
    last_update timestamp without time zone
);

CREATE INDEX ON public.stocks (code);

--
-- Name: stock_last_updates; Type: MATERIALIZED VIEW; Schema: public
--

CREATE MATERIALIZED VIEW IF NOT EXISTS public.stock_last_updates AS
 SELECT stocks.code,
    max(stocks.last_update) AS last_update
   FROM public.stocks
  GROUP BY stocks.code
  WITH NO DATA;

--
-- PostgreSQL database dump complete
--

REFRESH MATERIALIZED VIEW public.stock_last_updates;
