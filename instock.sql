--
-- PostgreSQL database dump
--

-- Dumped from database version 12.6
-- Dumped by pg_dump version 12.6

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: stocks; Type: TABLE; Schema: public
--

CREATE TABLE public.stocks (
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
    last_update timestamp without time zone
);

CREATE INDEX ON public.stocks (code);

--
-- Name: stock_last_updates; Type: MATERIALIZED VIEW; Schema: public
--

CREATE MATERIALIZED VIEW public.stock_last_updates AS
 SELECT stocks.code,
    max(stocks.last_update) AS last_update
   FROM public.stocks
  GROUP BY stocks.code
  WITH NO DATA;

--
-- PostgreSQL database dump complete
--
