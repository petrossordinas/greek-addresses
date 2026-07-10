# grAddresses

A read-only Go REST API for searching Greek addresses (prefectures, cities,
zip codes and streets) backed by a SQLite database. No frontend, no writes —
just HTTP GET.

## Requirements

- Go 1.25
- SQLite (the database ships in `db/gr_addresses.db`)

## Running locally

```sh
go run .
```

The server listens on `9013` by default. Build a binary with:

```sh
go build -o main .
./main
```

## Configuration

The port is resolved in this order:

1. `-port` command-line flag
2. `PORT` in a `.env` file (or the environment)
3. default: `9013`

```sh
./main -port=8080
# or
cp .env.example .env   # then edit PORT=8080
./main
```

## Docker

```sh
docker build -t graddresses .
docker run -p 9013:9013 graddresses
```

## API

### `GET /search`

All filters are combined with the query parameters below. Text filters
(`prefecture`, `city`, `zipcode`, `street`) match as a **prefix** by default
(`name LIKE "qry%"`). Include a literal `%` anywhere in the value to control
the match yourself — e.g. `%word%` for a substring search, `%word` for a
suffix search.

| Param | Description |
|---|---|
| `prefecture` | Search prefectures by name. Ignored if `prefecture_id` is set. |
| `prefecture_id` | Look up a single prefecture by id. |
| `city` | Search cities by name. Ignored if `city_id` is set. |
| `city_id` | Look up a single city by id. |
| `zipcode` | Search zip codes. Ignored if `zipcode_id` is set. Use `zipcode=all` to return every zip code, ordered ascending. |
| `zipcode_id` | Look up a single zip code by id. |
| `street` | Search streets by name. May end with a street number (e.g. `ΑΓΙΟΥ ΚΩΝΣΤΑΝΤΙΝΟΥ 12`) to also filter by the street's odd/even number ranges. |

`prefecture_id`, `city_id` and `zipcode_id` narrow a street search to that
exact prefecture/city/zipcode. Results are capped at 40.

**When `street` is present**, `prefecture`, `city` and `zipcode` stop being
standalone lookups and instead become wildcard filters on the street search
(joined against the related tables), combinable with the `*_id` filters
above as AND conditions. Without `street`, they behave as standalone
searches as described in the table.

#### Examples

```
# Cascading lookup: find a prefecture, then cities in it, then zip codes, then streets
GET /search?prefecture=ΑΤΤ
GET /search?prefecture_id=6&city=ΑΘ
GET /search?city_id=35&zipcode=104
GET /search?zipcode_id=1&street=ΑΓΙΟΥ ΚΩΝΣΤΑΝΤΙΝΟΥ 12

# All zip codes
GET /search?zipcode=all

# Street search narrowed by wildcard prefecture/city instead of prior lookups
GET /search?street=28ΗΣ ΟΚΤΩΒΡΙΟΥ&prefecture=ΘΕΣΣΑΛ%
GET /search?street=28ΗΣ ΟΚΤΩΒΡΙΟΥ&prefecture_id=6&city=%ΘΗΝΑ

# Substring / suffix search
GET /search?prefecture=%ΚΑΔ%
GET /search?city=%ΒΑΡΒΑΡΑ
```
