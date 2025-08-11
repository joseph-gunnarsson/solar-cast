// src/App.jsx
import React, { useState } from "react";
import Home from "./pages/Home";
import Estimate from "./pages/Estimate";

export default function App() {
  const [view, setView] = useState("estimate");

  return (
    <div className="min-h-screen bg-gray-900 text-gray-100">
      {view === "home" && (
        <Home
          onEstimate={() => setView("estimate")}
        />
      )}
      {view === "estimate" && (
        <>
          {/* Floating back button top-right */}
          <button
            onClick={() => setView("home")}
            className="fixed top-4 right-4 z-50 rounded-xl border border-gray-700 bg-gray-800 hover:bg-gray-700 px-3 py-2 text-sm"
            aria-label="Go back to home"
          >
            ← Back
          </button>
          <Estimate />
        </>
      )}
    </div>
  );
}
