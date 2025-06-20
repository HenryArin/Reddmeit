🧠 ReddmeitAlpha – Personalized Subreddit Recommender
ReddmeitAlpha is a Go-based tool that securely connects to the Reddit API using OAuth2 to fetch a user's:

Subscribed subreddits

Upvoted posts

Commented posts

It then analyzes interaction patterns to build a behavioral profile across subreddits. This data can be used to:

Visualize your Reddit activity

Generate smart subreddit recommendations using AI (coming soon)

Understand where you're most active—even outside your subscriptions

✨ Features
OAuth2 authentication with Reddit (token stored in .env)

Secure .gitignore setup to protect sensitive info

Handles Reddit API pagination to fetch all relevant data

Combines subscriptions, upvotes, and comments into unified stats

Designed for AI integration (e.g., GPT-based recommendation engine)

🛠 Built With
Go (Golang) — fast, clean standard-library HTTP and JSON

Reddit API — authenticated access to user data

godotenv — for secure .env file loading

🚀 Upcoming (Next Steps)
Integrate with OpenAI to recommend new subreddits based on your activity

Save and visualize trends in user behavior

Export interaction data for analysis
