# Claude Proxy Frontend

React + TypeScript + Vite frontend for the Claude Proxy application.

## Development

The frontend runs on port 5173 and proxies API requests to the backend on port 4000.

### Prerequisites

- Node.js 18+
- pnpm

### Install Dependencies

```bash
pnpm install
```

### Run Development Server

From the project root:

```bash
make run
```

This will start both the backend (port 4000) and frontend (port 5173) concurrently.

Or run frontend only:

```bash
make fe-dev
# or
cd frontend && pnpm dev
```

### Build for Production

From the project root:

```bash
make build
```

This will:
1. Build the frontend to `frontend/dist`
2. Embed the static files into the Go binary
3. Output the final binary to `bin/claude-proxy`

The production binary serves the frontend at the root path and API endpoints at their configured routes.

## Project Structure

```
frontend/
├── src/
│   ├── App.tsx          # Main application component
│   ├── main.tsx         # Application entry point
│   └── assets/          # Static assets
├── public/              # Public assets
├── index.html           # HTML template
├── vite.config.ts       # Vite configuration
└── tsconfig.json        # TypeScript configuration
```

## API Integration

The frontend communicates with the backend through the Vite dev server proxy (in development) or directly (in production).

Example API call:

```typescript
fetch('/health')
  .then(res => res.json())
  .then(data => console.log(data))
```

In development, this proxies to `http://localhost:4000/health`
In production, this calls the same binary serving both frontend and backend.

## Expanding the ESLint configuration

If you are developing a production application, we recommend updating the configuration to enable type-aware lint rules:

```js
export default defineConfig([
  globalIgnores(['dist']),
  {
    files: ['**/*.{ts,tsx}'],
    extends: [
      // Other configs...

      // Remove tseslint.configs.recommended and replace with this
      tseslint.configs.recommendedTypeChecked,
      // Alternatively, use this for stricter rules
      tseslint.configs.strictTypeChecked,
      // Optionally, add this for stylistic rules
      tseslint.configs.stylisticTypeChecked,

      // Other configs...
    ],
    languageOptions: {
      parserOptions: {
        project: ['./tsconfig.node.json', './tsconfig.app.json'],
        tsconfigRootDir: import.meta.dirname,
      },
      // other options...
    },
  },
])
```

You can also install [eslint-plugin-react-x](https://github.com/Rel1cx/eslint-react/tree/main/packages/plugins/eslint-plugin-react-x) and [eslint-plugin-react-dom](https://github.com/Rel1cx/eslint-react/tree/main/packages/plugins/eslint-plugin-react-dom) for React-specific lint rules:

```js
// eslint.config.js
import reactX from 'eslint-plugin-react-x'
import reactDom from 'eslint-plugin-react-dom'

export default defineConfig([
  globalIgnores(['dist']),
  {
    files: ['**/*.{ts,tsx}'],
    extends: [
      // Other configs...
      // Enable lint rules for React
      reactX.configs['recommended-typescript'],
      // Enable lint rules for React DOM
      reactDom.configs.recommended,
    ],
    languageOptions: {
      parserOptions: {
        project: ['./tsconfig.node.json', './tsconfig.app.json'],
        tsconfigRootDir: import.meta.dirname,
      },
      // other options...
    },
  },
])
```
