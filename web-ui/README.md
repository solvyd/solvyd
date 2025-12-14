# Ritmo Web UI

Modern React-based dashboard for Ritmo CI/CD platform.

## Features

- **Dashboard**: Real-time overview of builds, workers, and system metrics
- **Jobs Management**: Create, configure, and trigger CI/CD jobs
- **Build Monitoring**: Live build status, logs, and artifacts
- **Worker Management**: Monitor and manage worker fleet
- **Deployment Tracking**: Track CD deployments across environments
- **Plugin Marketplace**: Browse and install plugins

## Tech Stack

- React 18
- TypeScript
- Vite
- TailwindCSS
- React Query (TanStack Query)
- React Router
- Axios
- Recharts (for visualizations)
- Lucide React (icons)

## Development

```bash
# Install dependencies
npm install

# Start development server
npm run dev

# Build for production
npm run build

# Preview production build
npm run preview
```

The development server will start on http://localhost:3000 with API proxying to the backend at http://localhost:8080.

## Project Structure

```
src/
  components/       # Reusable UI components
    Layout.tsx      # Main layout with sidebar
  pages/            # Page components (routes)
    Dashboard.tsx   # Main dashboard
    Jobs.tsx        # Jobs list
    Builds.tsx      # Builds list
    Workers.tsx     # Workers management
    Deployments.tsx # Deployments tracking
    Plugins.tsx     # Plugins management
  lib/              # Utilities and helpers
    api.ts          # API client and functions
  types/            # TypeScript type definitions
    index.ts        # Shared types
  App.tsx           # Root component with routing
  main.tsx          # Application entry point
  index.css         # Global styles
```

## Features To Implement

- [ ] Job creation and configuration forms
- [ ] Real-time build log streaming via WebSocket
- [ ] Build metrics and analytics charts
- [ ] Worker capacity and utilization graphs
- [ ] Deployment pipeline visualization
- [ ] Plugin installation UI
- [ ] User authentication and authorization
- [ ] Dark mode support
- [ ] Mobile responsive design improvements
- [ ] Advanced filtering and search
- [ ] Notifications and alerts
- [ ] Artifact browser and download

## API Integration

All API calls are made through the centralized API client in `src/lib/api.ts` using Axios with proper error handling and authentication.

The UI automatically connects to the API server via Vite proxy configuration.
