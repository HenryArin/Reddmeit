# ReddmeitAlpha

ReddmeitAlpha is a small Go application that connects to the Reddit API to collect your subreddit activity. The data is combined into a single profile that can later be used for AI powered subreddit recommendations.

---

## Features

- Fetch subscribed subreddits
- Retrieve subreddits from upvoted posts
- Retrieve subreddits from user comments
- Combine all data into a structured plan
- `.env` file keeps your credentials out of the source

---

## Setup

1. **Clone the repository**

```bash
git clone https://github.com/your-username/ReddmeitAlpha.git
cd ReddmeitAlpha
```

2. **Create a `.env` file** with your Reddit credentials:

```bash
REDDIT_ACCESS_TOKEN=your_access_token_here
REDDIT_USERNAME=your_reddit_username_here
```

3. **Install dependencies**

```bash
go mod tidy
```

4. **Run the application**

```bash
go run main.go
```

---

## Built With

- Go
- Reddit API
- [godotenv](https://github.com/joho/godotenv)

---

## TODO / Coming Soon

- Generate subreddit recommendations using AI
- Save interaction data to JSON or a database
- Web dashboard or CLI report
- Token auto-refresh

---

## License

MIT – feel free to use and modify.
