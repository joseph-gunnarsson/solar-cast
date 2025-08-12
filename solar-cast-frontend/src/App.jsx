// src/App.jsx
import React, { useState } from "react";
import Home from "./pages/Home";
import Estimate from "./pages/Estimate";

export default function App() {
  const [view, setView] = useState("home");
  const [request_data, setRequestData] = useState({});
  return (
    <div className="min-h-screen bg-gray-900 text-gray-100">
      {view === "home" && (
        <Home
          setRequestData={setRequestData}
          onEstimate={() => setView("estimate")}
        />
      )}
      {view === "estimate" && (
        <>
          <button
            onClick={() => setView("home")}
            className="fixed top-4 right-4 z-50 rounded-xl border border-gray-700 bg-gray-800 hover:bg-gray-700 px-3 py-2 text-sm"
            aria-label="Go back to home"
          >
            ← Back
          </button>
          <Estimate
            requestData={request_data}
          />
        </>
      )}
    </div>
  );
}
