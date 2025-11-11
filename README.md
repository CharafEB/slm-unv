# SLM-UNV: Smart Language Model for University Information

This project is a sophisticated system that combines a Go backend with a Python-based machine learning service to create a smart assistant capable of answering questions, with a special focus on university-related topics. It leverages the Model Context Protocol (MCP) to allow a large language model (LLM) to interact with a PostgreSQL database.

## Architecture Overview

The system is composed of three main parts that communicate via Redis:

1.  **Python Embeddings Service**: A FastAPI application that serves as the entry point for user queries. It classifies incoming messages to determine if they are university-related and then routes them accordingly.
2.  **Go Router**: A message consumer that listens for classified messages from the embeddings service. It interacts with an LLM (via `mcphost` and Ollama) to understand the user's intent and utilize the tools provided by the Go backend.
3.  **Go Backend (`mcp-qwen-server`)**: The main MCP server that provides tools for the LLM to interact with the database. These tools allow the LLM to search for articles by author or title.

### Data Flow

1.  A user sends a query to the **Python Embeddings Service**.
2.  The service classifies the query and publishes it to a Redis channel.
3.  The **Go Router**, subscribed to this channel, picks up the message.
4.  The router uses `mcphost` to pass the query to an Ollama-hosted LLM.
5.  The LLM, if it needs to find information, uses the tools exposed by the **Go Backend** (e.g., `author_article`, `article_search`).
6.  The Go Backend executes the tool function (e.g., a database query) and returns the result to the LLM.
7.  The LLM formulates a final answer and sends it back to the Go Router.
8.  The Go Router publishes the final answer to a user-specific Redis channel.
9.  The user receives the answer via a websocket connection to the Python Embeddings Service.

## Components

### Go Backend (`mcp-qwen-server`)

-   **`main.go`**: Entry point for the MCP server. It establishes a connection to the PostgreSQL database and registers the available tools.
-   **`model/`**: Handles all database logic. It uses the `pg_trgm` extension for fuzzy string matching on author and article titles, making the search more robust against typos.
-   **`tools/`**: Defines the tools available to the LLM, such as `AuthorTool` and `ArticleTool`.
-   **`types/`**: Contains the Go structs for data structures used throughout the backend.

### Go Router

-   **`router/main.go`**: This application connects to Redis and `mcphost`. It orchestrates the communication between the user's query, the LLM, and the Go backend.

### Python Embeddings Service

-   **`embeddings/main.py`**: A FastAPI application with two main endpoints:
    -   `/message/{user_id}` (POST): To receive user messages.
    -   `/ws/{user_id}` (WebSocket): To send responses back to the user in real-time.
-   **`embeddings/classifier.py`**: Uses a pre-trained FastText model to classify user input. It includes a fallback mechanism for keyword-based similarity and logs uncertain predictions for later review.
-   **`embeddings/train.py`**: A script to train the FastText classification model on your own data (`data.txt`).

## Setup and Installation

### Prerequisites

-   Go (version 1.20 or later)
-   Python (version 3.9 or later)
-   PostgreSQL
-   Redis
-   Ollama (with a model like `qwen` pulled)

### Configuration

The project uses a `.env` file for configuration. Create a `.env` file in the root of the project and add the following variables:

```
# PostgreSQL
DB="postgres://user:password@localhost:5432/database_name?sslmode=disable"

# Redis
REDISADD_GO="localhost:6379"
REDISADD_PY="localhost"
REDISPORT="6379"
REDISUSERNAME=""
REDISPASSWORD=""
REDIS_CHANNEL="new_channel"

# MCPHost and Ollama
OLLAMA_HOST="http://localhost:11434"
MCPHOST_MODEL="qwen"
```

### Database Setup

1.  Connect to your PostgreSQL instance.
2.  Create the `pg_trgm` extension, which is required for the fuzzy search functionality:
    ```sql
    CREATE EXTENSION IF NOT EXISTS pg_trgm;
    ```
3.  Create an `articles` table:
    ```sql
    CREATE TABLE articles (
        id SERIAL PRIMARY KEY,
        title TEXT NOT NULL,
        author TEXT NOT NULL,
        content TEXT
    );
    ```
4.  Populate the table with some data.

### Python Service Setup

1.  Install the required Python packages:
    ```bash
    pip install -r embeddings/requirements.txt
    ```
    *(Note: You may need to create a `requirements.txt` file with `fastapi`, `redis`, `uvicorn[standard]`, `fasttext`, `python-dotenv`)*

2.  Prepare your training data in `embeddings/data.txt`. The format should be one line per training example, with the label prefixed by `__label__`. For example:
    ```
    __label__university_query What are the registration dates?
    __label__general_query Tell me a fun fact.
    ```

3.  Train the classification model:
    ```bash
    python3 embeddings/train.py
    ```
    This will create a `university_model_v2.bin` file.

## Running the System

You will need to run the three main components in separate terminals.

### 1. Run the Go Backend

```bash
go run main.go
```

### 2. Run the Go Router

```bash
go run router/main.go
```

### 3. Run the Python Embeddings Service

```bash
cd embeddings
uvicorn main:app --reload
```

Your system is now running! You can send requests to `http://127.0.0.1:8000/message/{any_user_id}` and listen for responses on the websocket `ws://127.0.0.1:8000/ws/{any_user_id}`.