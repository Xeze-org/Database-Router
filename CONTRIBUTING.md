# Contributing to Database Router

First off, thank you for considering contributing to Database Router! It's people like you that make this tool perfectly secure, fast, and reliable for the community.

## 🐛 Reporting Bugs
Before creating bug reports, please check the existing issues; you might find out that someone has already reported it or a fix is underway. When you create a bug report, please include as many details as possible:
* Your exact command or code snippet leading to the bug.
* The OS and runtime versions (Docker, Go, Python, Node) you are running.
* Logs or error stack traces.

## 💡 Suggesting Enhancements
Enhancement suggestions are tracked as [GitHub issues](https://github.com/Xeze-org/Database-Router/issues). When you create an enhancement issue, please describe:
* Exactly what you want to achieve.
* Why it is useful for the broader user base.
* Any potential technical approaches or architecture designs you have in mind.

## 🛠️ Local Development

The project consists of a core router (Go) and multiple client SDKs (Python, Node.js). 

### Prerequisites
* [Docker](https://docs.docker.com/get-docker/) & Docker Compose
* [Go 1.22+](https://go.dev/) (For core router development)
* [Python 3.12+](https://python.org/) & `pip` (For Python SDK development)
* [Node.js 22+](https://nodejs.org/) (For Node SDK development)

### Setting up the Core Router
1. Clone the repository: `git clone https://github.com/Xeze-org/Database-Router.git`
2. Change into the root directory: `cd Database-Router`
3. Spin up the local development stack (Postgres, MongoDB, Redis):
   ```bash
   cd local-deploy
   docker-compose up -d
   ```
4. Build the core router locally: `go build -o db-router ./cmd/router`

### Setting up the SDKs
* **Python**: `cd sdk/python && pip install -e .`
* **Node.js**: `cd sdk/node && npm install`

## 🚀 Pull Request Process

1. Fork the repo and create your feature branch from `main`.
2. Write clean, self-documenting code.
3. Ensure your code conforms to the existing style guides (e.g. `gofmt` for Go, `black` for Python, `prettier` for Node).
4. Update the `README.md` or `docs/` with details of changes to the interface or deployment steps, if applicable.
5. If your PR introduces a new feature, please include corresponding unit tests or update the examples under `/examples/`.
6. Submit your PR! A maintainer will review it and help merge it.

## 📜 License
By contributing to Database Router, you agree that your contributions will be licensed under its Apache 2.0 License.
