# Web client

This document describes the Mino web application, built with Next.js 14
and React 18. The web client provides the full feature set for managing
recordings, memories, tasks, and chatting with the AI assistant.

## Technology stack

The web application uses the following technologies.

| Technology | Purpose |
|------------|---------|
| Next.js 14 | App Router, server-side rendering, API proxying |
| React 18 | UI components |
| Zustand 4.5 | Global state management |
| Tailwind CSS | Styling with custom dark theme |
| Lucide React | Icon library |
| TypeScript | Type safety |

## Project structure

The web application follows the Next.js App Router convention.

```
web/
├── app/                    # Pages and layouts
│   ├── layout.tsx          # Root layout (zh-CN locale)
│   ├── page.tsx            # Redirects to /auth/login
│   ├── globals.css         # Global styles and CSS variables
│   ├── auth/
│   │   └── login/
│   │       └── page.tsx    # Login page
│   └── dashboard/
│       ├── layout.tsx      # Dashboard shell (sidebar + content)
│       ├── page.tsx        # Home: stats, recent items
│       ├── memories/       # Memory management
│       ├── tasks/          # Task management
│       ├── audio/          # Audio recordings
│       ├── chat/           # AI chat interface
│       ├── extensions/     # Extension management
│       └── settings/       # Settings (cloud, MCP)
├── components/
│   ├── ui/                 # Reusable primitives
│   ├── layout/             # Sidebar, header
│   └── features/           # Domain-specific components
├── lib/
│   ├── api/                # Backend API client modules
│   ├── store.ts            # Zustand global store
│   ├── utils.ts            # Utility functions
│   └── demo-data.ts        # Demo data for development
└── types/
    └── index.ts            # TypeScript interfaces
```

## Pages

The web application provides the following pages.

### Login (`/auth/login`)

Username and password authentication form. On successful login, the
access token and refresh token are stored in `localStorage`. The user
is redirected to the dashboard.

### Dashboard (`/dashboard`)

The home page displays an overview of your data: total conversations,
memories, and tasks. It shows recent memories, recent conversations,
and pending tasks in a grid layout.

### Memories (`/dashboard/memories`)

Lists all extracted memories with filtering by category (insight, fact,
preference, event) and text search. Each memory card shows the content,
category icon, importance level (displayed as a 5-dot indicator), and
a link to the source conversation audio.

### Tasks (`/dashboard/tasks`)

Lists all tasks with filtering by status (pending, in progress,
completed, cancelled). A stats bar shows counts for each status. Tasks
display their title, description, priority badge, due date, and an
inline status picker dropdown for quick updates.

### Audio (`/dashboard/audio`)

Lists all recorded conversations with search functionality. Each
conversation card shows the title, summary, duration, tags, and a
play/pause button for audio playback. A stats bar shows total
recordings, total duration, and completed count.

### Chat (`/dashboard/chat`)

A full chat interface with a session sidebar on the left and a message
area on the right. You can create, rename, and delete chat sessions.
Messages display with user/assistant styling, timestamps, and source
citations that link to original conversations.

### Extensions (`/dashboard/extensions`)

Manage custom extensions with enable/disable toggles. You can add new
extensions and delete existing ones. Each extension has a name,
description, icon, and JSON configuration.

### Settings (`/dashboard/settings`)

The settings hub provides configuration for theme, language,
notification preferences, and LLM provider selection. Sub-pages
include cloud sync settings (MinIO and PostgreSQL connection
configuration) and MCP protocol settings (enable/disable with server
configuration).

## API layer

The web application communicates with the backend through a client-side
API layer. All API calls go through the Next.js rewrite proxy, which
forwards `/api/v1/*` requests to the backend.

### API client

The base API client (`lib/api/client.ts`) provides the following
features.

- Automatic JWT token management using `localStorage`
- Automatic token refresh on 401 responses
- Redirect to login page on authentication failure
- Typed request/response wrappers

### API modules

Each backend resource has a dedicated API module.

| Module | File | Endpoints |
|--------|------|-----------|
| Auth | `lib/api/auth.ts` | Sign in, sign out, change password |
| Conversations | `lib/api/conversations.ts` | List, get, delete |
| Memories | `lib/api/memories.ts` | List, get, update, delete |
| Tasks | `lib/api/tasks.ts` | List, create, update, delete |
| Chat | `lib/api/chat.ts` | Sessions CRUD, messages |
| Extensions | `lib/api/extensions.ts` | Full CRUD |
| Search | `lib/api/search.ts` | Search, reindex |

All modules are re-exported through `lib/api/index.ts` as namespaced
objects (`authApi`, `conversationsApi`, `memoriesApi`, `tasksApi`,
`chatApi`, `extensionsApi`, `searchApi`).

## State management

The application uses a single Zustand store (`lib/store.ts`) that
manages all client-side state.

| State | Description |
|-------|-------------|
| `user` | Current authenticated user |
| `conversations` | Conversation list |
| `memories` | Memory list |
| `tasks` | Task list |
| `extensions` | Extension list |
| `chatSessions` | Chat session list |
| `activeSessionId` | Currently selected chat session |
| `chatMessages` | Messages for the active session |
| `settings` | App settings (theme, language, LLM config) |
| `isRecording` | Whether a recording is in progress |
| `currentTranscript` | Live transcript text during recording |

## Design system

The web application uses a dark theme optimized for OLED displays. The
color palette is defined in `tailwind.config.js` and `globals.css`.

| Token | Value | Usage |
|-------|-------|-------|
| Background | `#09090b` | Page background |
| Surface | `#18181b` | Card backgrounds |
| Border | `#27272a` | Borders and dividers |
| Text primary | `#fafafa` | Main text |
| Text secondary | `#a1a1aa` | Secondary text |
| CTA / Accent | `#a3e635` | Buttons, active states |

Typography uses Inter for body text and Newsreader for serif accents.
The application supports `prefers-reduced-motion` for accessibility.

## Global search

The header includes a global search bar activated by clicking the
search icon or pressing Cmd+K (Ctrl+K on Windows/Linux). Search
queries are sent to the Typesense-backed search API with debouncing.
Results appear in a dropdown showing matches from both conversations
and memories.

## Recording

The header includes a recording toggle button. When recording is
active, a `RecordingBanner` component appears below the header showing
the live transcript. The recording state and transcript are managed
through the Zustand store.

## Development

### Install dependencies

```bash
cd web
npm install
```

### Run the development server

```bash
npm run dev
```

The app starts at `http://localhost:3000`.

### Build for production

```bash
npm run build
```

### Type checking

```bash
npm run typecheck
```

### Linting

```bash
npm run lint
```

## Configuration

The web application uses the following environment variables.

| Variable | Default | Description |
|----------|---------|-------------|
| `API_URL` | `http://localhost:8000` | Backend API URL for the proxy rewrite |
| `NEXT_PUBLIC_API_URL` | `/api/v1` | Client-side API base URL |

The `next.config.js` file configures a rewrite rule that proxies
`/api/v1/*` requests to the backend, avoiding CORS issues during
development.

## Next steps

For backend API details, see the [API reference](api-reference.md).
For production deployment of the web app, see the
[Deployment](deployment.md) guide.
