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

Both the port and the database file path are resolved in this order:

1. command-line flag (`-port`, `-db`)
2. `.env` file (or the environment): `PORT`, `DB_PATH`
3. defaults: `9013`, `db/gr_addresses.db`

The database path is relative to the process's working directory unless you
give it an absolute path — this matters if you run the binary from somewhere
other than the repo root.

```sh
./main -port=8080 -db=/path/to/gr_addresses.db
# or
cp .env.example .env   # then edit PORT and DB_PATH
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

#### Response format

Every successful response is a JSON object with a single `results` key.
Its value is an array of one of the shapes below (or `null` if nothing
matched, or if none of the recognized query params were supplied).

**Prefecture** (`prefecture`):
```json
{"results": [{"id": 3, "name": "ΑΡΓΟΛΙΔΑΣ"}]}
```

**City** (`city`) — `prefecture` is preloaded:
```json
{"results": [{"id": 35, "prefecture_id": 6, "prefecture": {"id": 6, "name": "ΑΤΤΙΚΗΣ"}, "name": "ΑΘΗΝΑ"}]}
```

**Zipcode** (`zipcode`, `zipcode_id`, `zipcode=all`) — `city` and `prefecture` are preloaded:
```json
{"results": [{"id": 1, "city_id": 35, "city": {"id": 35, "prefecture_id": 6, "name": "ΑΘΗΝΑ"}, "prefecture_id": 6, "prefecture": {"id": 6, "name": "ΑΤΤΙΚΗΣ"}, "zipcode": "10431"}]}
```

**Street** (`street`) — `zipcode`, `city` and `prefecture` are preloaded. If the
query included a street number, `name` has it appended and the result set is
already filtered down to streets whose odd/even `ranges` cover that number:
```json
{"results": [{"id": 3, "zipcode_id": 1, "zipcode": {"id": 1, "city_id": 35, "prefecture_id": 6, "zipcode": "10431"}, "city_id": 35, "city": {"id": 35, "prefecture_id": 6, "name": "ΑΘΗΝΑ"}, "prefecture_id": 6, "prefecture": {"id": 6, "name": "ΑΤΤΙΚΗΣ"}, "name": "ΑΓΙΟΥ ΚΩΝΣΤΑΝΤΙΝΟΥ 12", "ranges": "{\"odd\": [{\"to\": \"23\", \"from\": \"1\"}], \"even\": [{\"to\": \"20\", \"from\": \"2\"}]}"}]}
```

`ranges` is itself a JSON string (not a nested object) with `odd`/`even`
arrays of `{"from": "...", "to": "..."}` number ranges.

#### Errors

A `prefecture_id`/`city_id`/`zipcode_id` for a record that doesn't exist is
not an error — it returns `200 OK` with `{"results": []}`, same as a text
filter with no matches.

An actual DB error returns a bare `500 Internal Server Error` with a
plain-text body (`Internal Server Error`) and no JSON — check
`gr_addresses.log` for the underlying error, it isn't returned in the
response.
