# AI Commit Message Generator

A robust, production-ready CLI tool written in Go that uses Ollama API to automatically generate meaningful git commit messages from your staged changes.

## Problem Statement
Writing good commit messages is essential for maintaining a clean history and facilitating code reviews. However, it can be tedious and inconsistent across teams. Developers often resort to generic messages like "fix bug" or "update code" when in a rush.

## Objectives
- **Automation**: Automate the process of reading git diffs and crafting commit messages.
- **Consistency**: Enforce team conventions via custom rules.
- **Intelligence**: Use advanced AI to analyze the *intent* of changes, suggesting splits for complex diffs.
- **Speed**: Built with the Go Standard Library and go-git library for zero external binary dependencies and fast execution.

## Features
- **Smart Diff Analysis**: Reads staged changes and context.
- **Split Suggestions**: Detects if a diff contains multiple logical changes and suggests breaking them down (displayed in Yellow).
- **Custom Rules**: Respects `.git-commit-rules-for-ai` in your repo root for team-specific guidelines.
- **Conventional Commits**: Generates messages in the `<type>(<scope>): <description>` format.
- **Easy Installation**: Platform-specific installation scripts for Windows, Mac, and Linux.
- **Pre-commit Hook Integration**: Automatically generate commit messages when you run `git commit`.
- **No External Dependencies**: Uses the go-git library - no git binary installation required.
- **Hexagonal Architecture**: Clean, testable, and maintainable codebase.

## Directory Structure
The project follows a Hexagonal (Ports and Adapters) Architecture:

```
/
├── cmd/
│   └── generate-commit/
│       └── main.go            # Entry point. Wires up dependencies.
├── internal/
│   ├── ai/
│   │   ├── generate_commit_message.go      # Ollama API Client (Conventional Commits defined here)
│   │   └── generate_commit_message_test.go # Unit tests
│   ├── config/
│   │   ├── config.go                        # Configuration loader (.commit-generator-config)
│   │   ├── config_test.go                  # Config tests
│   │   ├── git_commit_rules.go             # Rules Loader (.git-commit-rules-for-ai)
│   │   └── git_commit_rules_test.go        # Rules tests
│   ├── git/
│   │   ├── client.go           # Git Operations (using go-git library)
│   │   └── client_test.go      # Integration/Unit tests
│   └── app/
│       ├── app.go              # Core Application Logic / Orchestrator (init command)
│       └── app_test.go         # Table-Driven Unit Tests (Mocked)
├── install.sh                 # Mac/Linux installation script
├── install.ps1                # Windows PowerShell installation script
└── install.bat                # Windows batch installation script
├── go.mod                      # Go Module definition
└── README.md                   # Project Documentation
```

## Quick Start (Recommended)

### Download Pre-built Executable

**No Go installation required!** Download the pre-built binary for your platform from the [latest GitHub Release](https://github.com/YOUR_USERNAME/YOUR_REPO/releases/latest).

| Platform | Archive | Binary Name |
|----------|---------|-------------|
| 🪟 **Windows (64-bit)** | `generate-commit-windows.zip` | `generate-commit.exe` |
| 🪟 **Windows (ARM64)** | `generate-commit-windows-arm64.zip` | `generate-commit.exe` |
| 🍎 **Mac (Apple Silicon)** | `generate-commit-mac-arm64.tar.gz` | `generate-commit` |
| 🍎 **Mac (Intel)** | `generate-commit-mac-intel.tar.gz` | `generate-commit` |
| 🐧 **Linux (64-bit)** | `generate-commit-linux.tar.gz` | `generate-commit` |
| 🐧 **Linux (ARM64)** | `generate-commit-linux-arm64.tar.gz` | `generate-commit` |

### Setup Instructions

1. **Download the release**: Go to the [latest release page](https://github.com/YOUR_USERNAME/YOUR_REPO/releases/latest) and download the archive for your platform.

2. **Extract the archive**:
   - **Windows**: Extract the `.zip` file
   - **Mac/Linux**: Extract the `.tar.gz` file:
     ```bash
     tar -xzf generate-commit-*.tar.gz
     ```

3. **Make it executable** (Mac/Linux only):
   ```bash
   chmod +x generate-commit
   ```

4. **Set your API key**:
   ```bash
   # Mac/Linux
   export OLLAMA_API_KEY="your_api_key_here"
   
   # Windows (Command Prompt)
   set OLLAMA_API_KEY=your_api_key_here
   
   # Windows (PowerShell)
   $env:OLLAMA_API_KEY="your_api_key_here"
   ```

5. **Run the tool**:
   ```bash
   # Mac/Linux
   ./generate-commit
   
   # Windows
   generate-commit.exe
   ```

### Optional: Add to PATH

For easier access, move the binary to your PATH:

**Mac/Linux:**
```bash
sudo mv generate-commit /usr/local/bin/
```

**Windows:**
Move `generate-commit.exe` to a directory in your PATH (e.g., `C:\Windows\System32`)

---

## Build from Source

### Prerequisites
- **Go 1.21+** installed
- **Ollama API Key**: Get one from [Ollama](https://ollama.com)

**Note**: The tool uses the `go-git` library and does not require the `git` binary to be installed on your system.

### Installation Steps

1. Clone the repository:
   ```bash
   git clone <repo-url>
   cd ai-commit-message-generator
   ```

2. Build the binary:
   ```bash
   go build -o generate-commit ./cmd/generate-commit
   ```

3. (Optional) Move to your PATH:
   ```bash
   mv generate-commit /usr/local/bin/
   ```

## Installation

### Quick Install (One-Liner)

Install directly from GitHub with a single command:

#### Mac/Linux
```bash
curl -fsSL https://raw.githubusercontent.com/AllaySahoo222/AI-Commit-Message-Generator/main/install.sh | bash
```

#### Windows (PowerShell)
```powershell
iwr -useb https://raw.githubusercontent.com/AllaySahoo222/AI-Commit-Message-Generator/main/install.ps1 | iex
```

#### Windows (Command Prompt)
```cmd
powershell -Command "iwr -useb https://raw.githubusercontent.com/AllaySahoo222/AI-Commit-Message-Generator/main/install.ps1 | iex"
```

### Using Installation Scripts (Manual Download)

If you prefer to download the scripts first:

#### Mac/Linux
```bash
./install.sh
```

#### Windows (PowerShell)
```powershell
.\install.ps1
```

#### Windows (Command Prompt)
```cmd
install.bat
```

The installation scripts will:
- Detect your platform and architecture
- Download the latest release from GitHub
- Install the binary to a directory in your PATH
- Set up the tool for easy access

### Manual Installation

1. Download the appropriate binary for your platform (see Quick Start section)
2. Make it executable (Mac/Linux): `chmod +x generate-commit-*`
3. Move to a directory in your PATH or use it directly

## Usage

### Initial Setup

1. **Navigate to your git repository**:
   ```bash
   cd /path/to/your/project
   ```

2. **Initialize the tool** (first time only):
   ```bash
   generate-commit init
   ```

   This will create:
   - `.commit-generator-config` - Configuration file (update with your API key if needed)
   - `.git-commit-rules-for-ai` - Custom rules file (customize for your team)
   - `.git/hooks/pre-commit` - Pre-commit hook for automatic message generation

3. **Configure your API key** (if not set in environment):
   - Edit `.commit-generator-config` and add your `api_key`
   - Or set `OLLAMA_API_KEY` environment variable

### Generating Commit Messages

#### Option 1: Using Pre-commit Hook (Recommended)

After running `init`, the pre-commit hook is automatically set up:

1. **Stage your changes**:
   ```bash
   git add .
   ```

2. **Commit** (the hook will automatically generate the message):
   ```bash
   git commit
   ```

   The hook will:
   - Generate a commit message from your staged changes
   - Display it and prompt you to Accept, Reject, or Edit
   - Commit automatically if you accept

#### Option 2: Manual Generation

1. **Stage your changes**:
   ```bash
   git add .
   ```

2. **Run the tool**:
   ```bash
   generate-commit  # or 'generate-commit generate'
   ```

3. **Review the AI-generated commit message** and use it for your commit:
   ```bash
   git commit -m "feat(auth): add OAuth2 login support"
   ```

### Commands

- `generate-commit init` - Initialize repository with config, rules, and pre-commit hook
- `generate-commit generate` or `generate-commit` - Generate commit message from staged changes
- `generate-commit help` - Show help message

### Example Output

**Single commit message (Cyan):**
```
Generating commit message...

feat(auth): add OAuth2 login support
```

**Split suggestion (Yellow):**
```
Generating commit message...

AI Suggestion (Split Changes):
This diff contains multiple logical changes:
1. Authentication module (OAuth2 implementation)
2. Database schema updates (user table)
3. UI changes (login form)

Consider splitting into separate commits for better history.
```

### Conventional Commits

The tool generates commit messages following the [Conventional Commits](https://www.conventionalcommits.org/) specification.

**Format**: `<type>(<scope>): <description>`

**Supported Types** (defined in `internal/ai/generate_commit_message.go`):
- `feat` - A new feature
- `fix` - A bug fix
- `docs` - Documentation only changes
- `style` - Code style changes (formatting, missing semi-colons, etc.)
- `refactor` - Code refactoring without feature changes or bug fixes
- `test` - Adding or updating tests
- `chore` - Maintenance tasks, dependency updates, etc.

**Examples**:
- `feat(auth): add OAuth2 login support`
- `fix(api): resolve null pointer exception in user endpoint`
- `docs(readme): update installation instructions`
- `refactor(utils): simplify error handling logic`

The conventional commit types are hardcoded in the AI prompt at `internal/ai/generate_commit_message.go` (line 135). To customize the types or format, you can modify the prompt or add rules in `.git-commit-rules-for-ai`.

### Custom Rules

To enforce specific rules (e.g., "Mention Jira ID"), edit the `.git-commit-rules-for-ai` file created during `init`, or create it manually in the root of your repository.

**Example `.git-commit-rules-for-ai`:**
```text
- Always start with a verb (Add, Fix, Update).
- If the change affects the UI, mention it.
- Max 50 characters for the subject line.
- Include Jira ticket ID if applicable (e.g., PROJ-123).
```

### Configuration

The tool uses a configuration file `.commit-generator-config` (created during `init`) with the following options:

```json
{
  "api_key": "",              // Optional: Override OLLAMA_API_KEY env var
  "model": "gpt-oss:120b",    // AI model to use
  "base_url": "http://localhost:11434/api/generate",
  "timeout_seconds": 60
}
```

**Configuration Priority**:
1. Config file (`.commit-generator-config`)
2. Environment variable (`OLLAMA_API_KEY`)
3. Default values

## Running Tests
Run the comprehensive test suite (Unit + Integration):
```bash
go test -v ./...
```

## Where Are Conventional Commits Defined?

The conventional commit types and format are defined in the AI prompt at:
- **File**: `internal/ai/generate_commit_message.go`
- **Function**: `buildPrompt()` (lines 127-145)
- **Line 133**: References Conventional Commits specification
- **Line 134**: Defines the format: `<type>(<scope>): <description>`
- **Line 135**: Lists allowed types: `feat, fix, docs, style, refactor, test, chore`

To modify the supported types or format, edit the `buildPrompt()` function in that file.
