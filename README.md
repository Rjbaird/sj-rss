# Shonen Jump RSS

[sj-rss](https://sjapi.up.railway.app/) allows you to find to [RSS feeds](https://en.wikipedia.org/wiki/RSS) for Weekly [Shonen Jump](https://www.viz.com/shonenjump) manga such as Jujutsu Kaisen, Chainsaw Man and My Hero Academia!

Feeds are updated daily between 10am and 2pm, Central Time, UTC.

## Tech Stack

**Client:** [AlpineJS](https://alpinejs.dev/), [TailwindCSS](https://tailwindcss.com/)

**Server:** [Go](https://go.dev/), [Railway](https://railway.app/)

**Database:** [SQLite](https://www.sqlite.org/index.html)

## Prerequisites

Before installing sj-rss you'll need:

- Go v1.20 or later
- Node.js v20 or later
- SQLite v3 or later
- Make - optional, but recommended

## Environment Variables

When running this project, you can add the following environment variables to your .env file

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

Next, run the development server:

```bash
    go run ./cmd/app
```

Finally, run vite in a separate terminal window with the following command in the views directory:

```bash
    cd views
    npm run dev
```

## Issues

If you notice any problems with either this repository, the main site or individual feeds, please open an [issue](https://github.com/Rjbaird/sj-rss/issues) or a [pull request](https://github.com/Rjbaird/sj-rss/pulls).

## Upcoming Features

- Update UI design
- Add "copy to clipboard" for feed links
