# Shonen Jump RSS

[sj-rss](https://sjapi.up.railway.app/) allows you to find to [RSS feeds](https://en.wikipedia.org/wiki/RSS) for Weekly [Shonen Jump](https://www.viz.com/shonenjump) manga such as Jujutsu Kaisen, Chainsaw Man and My Hero Academia!

Feeds are updated daily between 10am and 2pm, Central Time, UTC.

## Tech Stack

**Client:** [AlpineJS](https://alpinejs.dev/), [TailwindCSS](https://tailwindcss.com/)

**Server:** [Go](https://go.dev/), [Railway](https://railway.app/)

**Database:** [Redis](https://redis.io/)


## Prerequisites

Before installing sj-rss you'll need:

- Go v1.2.0 or later
- Node.js v20.0.0 or later
- Make - optional, but recommended


## Environment Variables

To run this project, you will need to add the following environment variables to your .env file

`REDIS_URL`- required, the URL for your Redis instance

`PORT` - optional, default is 3000

## Installation

To install sj-rss, clone the repository and run the following commands:

If you have Make installed:

```bash
    cd sj-rss
    touch .env
    make install
```
Finally, run the development server:

```bash
    make dev
```

Otherwise install manually with the following commands:

```bash
    cd sj-rss
    go mod download
    go mod tidy
    cd views && npm install
    ./sj-rss
```

## Issues

If you notice any problems with either this repository, the main site or individual feeds, please open an [issue](https://github.com/Rjbaird/sj-rss/issues) or a [pull request](https://github.com/Rjbaird/sj-rss/pulls).


## Upcoming Features

- Replace Redis with SQLite
- Update UI design
- Add "copy to clipboard" for feed links
