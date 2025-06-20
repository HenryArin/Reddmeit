# 🧠 ReddmeitAlpha

**ReddmeitAlpha** is a Go application that securely connects to the Reddit API to fetch and analyze your subreddit activity. It combines your **subscriptions**, **upvoted posts**, and **comments** into a unified profile that can later be used for **smart subreddit recommendations** powered by AI.

---

## 📦 Features

- ✅ Fetches all subscribed subreddits  
- ✅ Retrieves subreddits from upvoted posts  
- ✅ Retrieves subreddits from user comments  
- ✅ Combines data into a structured view  
- 🔐 Uses `.env` to store sensitive credentials  
- 🧠 Prepares data for future AI-based subreddit suggestions

---

## 🔧 Setup

### 1. Clone the Repository

```bash
git clone https://github.com/your-username/ReddmeitAlpha.git
cd ReddmeitAlpha

2. Create .env File
Make a .env file with your Reddit API credentials:
REDDIT_ACCESS_TOKEN=your_access_token_here
REDDIT_USERNAME=your_reddit_username_here
⚠️ Never share your access token publicly.

3. Install Dependencies
go mod tidy
🚀 Run the App
go run get_subs.go

🛠 Built With
Go

Reddit API

godotenv

📌 TODO / Coming Soon
 Generate subreddit recommendations using AI

 Save interaction data to JSON or DB

 Web dashboard or CLI report

 Token auto-refresh

📄 License
MIT – feel free to use and modify.

```
