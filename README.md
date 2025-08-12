# Code Editing Agent

This project is a command-line AI agent powered by Anthropic's Claude LLM, capable of reading, listing, and editing files in your workspace using natural language.

## Prerequisites

- [Go](https://golang.org/dl/) (version 1.20 or newer recommended)
- An Anthropic API key ([get one here](https://console.anthropic.com/))

## Setup

1. **Clone the repository:**

   ```sh
   git clone https://github.com/your-username/code-editing-agent.git
   cd code-editing-agent
   ```

2. **Set your Anthropic API key:**

   On macOS/Linux:
   ```sh
   export ANTHROPIC_API_KEY=your_api_key_here
   ```

   On Windows (PowerShell):
   ```powershell
   $env:ANTHROPIC_API_KEY="your_api_key_here"
   ```

3. **Install dependencies:**

   ```sh
   go mod tidy
   ```

4. **Run the agent:**

   ```sh
   go run main.go
   ```

## Usage

- Type your questions or commands in the terminal (e.g., "Show me the contents of main.go" or "List all files").
- Use `ctrl-c` to exit.

## Features

- Read file contents
- List files in directories
- Edit files using natural language

## Notes

- The agent only has access to files in the current working directory.
- Make sure your API key is kept secret.

---