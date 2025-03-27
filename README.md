# AgentD

AgentD is a next-generation **distributed agent execution framework** designed for **modular automation, LLM orchestration, and vector database integration**. It follows the **DAP (Distributed Agent Protocol) v0.1** for structured, traceable execution.

## 🚀 Features

### 🔹 Modular Execution
- Supports **instruction URIs** (`llm://`, `tool://`, `vector-read://`, `vector-write://`, etc.)
- Enables **declarative task execution** using DAGs (Directed Acyclic Graphs)
- Built-in **parallel execution** and **conditional branching**

### 🔹 LLM Orchestration
- Supports multiple **LLM providers** (OpenAI, Hugging Face, Local models)
- **Memory persistence** via `.memory` flag
- **Streaming responses** for real-time interaction

### 🔹 Vector Database Integration
- Supports **Qdrant** and **ChromaDB** for vector search (Available Soon)
- Read/write vector embeddings with **`vector-read://`** and **`vector-write://`**
- Seamless **multi-tenant support**

### 🔹 Tracing & Observability
- Every execution step is **recorded in a structured trace**
- Supports **debugging, logging, and monitoring**

### 🔹 Extensibility & Plugins
- Custom **tooling and function calling** (`tool://` execution)
- Easily **extendable instruction types**
- Supports **third-party API integrations**

## 🏗️ Tech Stack

### 🖥️ Backend
- **Go (Golang)** → High-performance execution core
- **gRPC** → Internal microservices communication
- **ChromaDB / Qdrant** → Vector search and retrieval

### 🔌 Integrations
- **OpenAI API** → GPT models
- **Anthropic Claude, Mistral, Llama** → Additional LLM support

## 🔧 Installation
```sh
git clone https://github.com/aboyai/agentd.git
cd agentd
go build ./cmd/server
./server 
```

## 📜 License
Distributed under the **Apache 2.0 License**. Contributions welcome!

## 🌍 Community & Support
- **GitHub Issues**: Report bugs, request features
- **Discord**: Join the community discussion
- **Contributing**: PRs are welcome! Check `CONTRIBUTING.md`

---

🚀 **AgentD: Powering the future of AI-driven automation!**
