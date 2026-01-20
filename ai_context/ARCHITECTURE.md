# 🗺️ Project Directory Map

Agent, strictly follow this directory structure when creating or reading files.

```text
my-copycat-backend/
├── cmd/
│   └── server/
│       └── main.go           # [Entry] App initialization: Config -> DB -> Router -> Run
├── config/                   # [Config] Configuration structures and yaml files
├── internal/
│   ├── api/                  # [Interface Layer] HTTP handling ONLY
│   │   ├── handler/          # Request parsing, validation, calling Service, response formatting
│   │   ├── middleware/       # Global middleware (CORS, Auth, Logger)
│   │   └── router.go         # Gin route registration
│   ├── core/                 # [Service Layer] Business Logic (The Brain)
│   │   ├── agent/            # Orchestration logic (Deciding flow between Crawler/LLM)
│   │   ├── crawler/          # Web scraping logic (Colly)
│   │   └── llm/              # AI integration (OpenAI/DeepSeek SDK & Prompts)
│   ├── model/                # [Data Model] GORM Structs & DB Schemas
│   └── repository/           # [DAO Layer] Direct Database Operations (CRUD)
├── pkg/                      # [Shared] specific generic tools
│   ├── logger/               # Structured logging setup
│   ├── response/             # Standard API response wrapper ({code, msg, data})
│   └── utils/                # Helper functions (Hash, Time, etc.)
├── Dockerfile                # Deployment
├── go.mod                    # Dependencies
└── Makefile                  # Task runner (make run, make build)