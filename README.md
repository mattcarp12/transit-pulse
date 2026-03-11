# 🚇 Transit Pulse

![Go](https://img.shields.io/badge/go-%2300ADD8.svg?style=for-the-badge&logo=go&logoColor=white)
![React](https://img.shields.io/badge/react-%2320232a.svg?style=for-the-badge&logo=react&logoColor=%2361DAFB)
![AWS](https://img.shields.io/badge/AWS-%23FF9900.svg?style=for-the-badge&logo=amazon-aws&logoColor=white)
![MapLibre](https://img.shields.io/badge/MapLibre-1E88E5?style=for-the-badge&logo=maplibre&logoColor=white)

> **A highly scalable, serverless transit dashboard ingesting real-time GTFS data to render live train movements.**

*(✏️ TODO: Replace this text with a 5-second GIF of your map zooming in and the trains moving. You can use a free tool like Kap or Gifox to record your screen.)*

---

## 🏗️ The Architecture

Municipal transit data (GTFS) is notoriously fragmented. This engine mathematically bridges static geospatial tracks (`shapes.txt`) with real-time schedule updates (`TripUpdates`) using a lean, automated AWS serverless pipeline. 

*(✏️ TODO: Insert your Excalidraw/draw.io architecture diagram here. Save it as `architecture.png` in a `docs/` folder)*
`![Architecture Diagram](./docs/architecture.png)`

**Engineering Constraints & Trade-offs:**
Designed for civic-tech and enterprise environments, this architecture minimizes idle compute costs. By utilizing an event-driven AWS Lambda "Warm Cache" pattern rather than an always-on EC2 instance, it processes complex geospatial interpolations securely while keeping monthly infrastructure costs near zero.

---

## 🚀 Quick Start (Local Development)

To run the Go ingestion engine and the React map locally:

### 1. Environment Setup
Rename `.env.example` file in the `/frontend` directory to `.env.development`:
```env
VITE_LIVE_DATA_URL=http://localhost:8080/trains
VITE_SHAPES_URL=http://localhost:8080/route-shapes
```

### 2. Start the Go Backend
```bash
cd backend
LOCAL_MODE=true go run ./cmd/lambda
```

### 3. Start the React Frontend
```bash
cd ../frontend
npm install
npm run dev
```

Now navigate to `http://localhost:5173` to view the live dashboard.

## 📖 Deep Dives

Read more about the specific engineering decisions behind this architecture on my blog: