# PR Agent Reviewer

A Go-based pull request review agent that automatically analyzes code changes using AI and posts reviews as comments on pull requests or merge requests. It supports GitHub and GitLab and sends review summaries to Slack.

## ✨ Features

- ✅ Listens for GitHub/GitLab webhook events for new PRs or MRs  
- 🤖 Reviews code changes using OpenAI or Ollama (local LLM)  
- 💬 Posts detailed AI-generated comments on the pull/merge request  
- 📢 Sends review summaries to a configured Slack channel  

---

## 🚀 Setup

1. **Clone the repository:**

   ```bash
   git clone https://github.com/your-org/pr-agent-reviewer.git
   cd pr-agent-reviewer
   ```

2. **Install dependencies:**

   ```bash
   go mod download
   ```

3. **Create your `.env` file:**

   Use the provided `.env.example` as a reference:

   ```bash
   cp .env.example .env
   ```

   Fill in your credentials (see [Configuration](#-configuration) below).

4. **Run the application:**

   ```bash
   go run main.go
   ```

---

## ⚙️ Configuration

Set the following environment variables in your `.env` file:

### 🔐 GitHub / GitLab

- `VCS_PROVIDER`: Either `github` or `gitlab`
- `GITHUB_WEBHOOK_SECRET`: Secret for GitHub webhook verification (GitHub only)
- `GITHUB_ACCESS_TOKEN`: GitHub personal access token
- `GITLAB_TOKEN`: GitLab personal access token
- `GITHUB_BOT_USERNAME`: Bot username to post comments on GitHub

### 🧠 AI Provider

- `AI_PROVIDER`: `openai` or `ollama`
- If using **OpenAI**:
  - `OPENAI_API_KEY`: Your OpenAI API key
- If using **Ollama (local)**:
  - `OLLAMA_BASE_URL`: e.g., `http://localhost:11434`
  - `OLLAMA_MODEL`: e.g., `deepseek-coder:6.7b`

### 📢 Slack Notifications

- `SLACK_BOT_TOKEN`: Your Slack bot token
- `SLACK_CHANNEL_ID`: Channel ID where summaries should be posted

### 🌐 Server

- `PORT`: Port for running the server (e.g., `8080`)

---

## 🛠 Usage

1. **Set up a webhook** in your GitHub or GitLab repository pointing to:

   ```
   http://<your-server-host>:<PORT>/webhook
   ```

2. When a new PR (GitHub) or MR (GitLab) is created or updated:
   - The agent fetches the diff
   - It sends the changes to the selected AI model
   - It posts comments inline and/or as a summary
   - A summary is sent to the Slack channel

---

## 📌 Example `.env`

```env
GITHUB_WEBHOOK_SECRET=your_webhook_secret
OPENAI_API_KEY=your_openai_api_key_here
SLACK_BOT_TOKEN=xoxb-...
SLACK_CHANNEL_ID=C08SMSQ4RJR
PORT=8080

AI_PROVIDER=ollama
OLLAMA_BASE_URL=http://localhost:11434
OLLAMA_MODEL=deepseek-coder:6.7b

VCS_PROVIDER=gitlab
GITLAB_TOKEN=glpat-...
GITHUB_TOKEN=ghp_...
GITHUB_BOT_USERNAME=Aboelsaud
```

---

## 🧪 Future Improvements

- Support for Bitbucket and other VCS providers
- Enhanced inline comment grouping
- PR summary scoring and recommendations
