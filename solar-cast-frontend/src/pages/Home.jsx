import React, { useEffect, useId, useMemo, useRef, useState } from "react";

const API_BASE = import.meta.env.VITE_API_BASE_URL || "/api";

export default function Home({ onSubmit }) {
  const [mode, setMode] = useState("city"); 
  const [query, setQuery] = useState("");
  const [lat, setLat] = useState("");
  const [lon, setLon] = useState("");
  const [panel, setPanel] = useState("Mono-Default-400");
  const [error, setError] = useState("");

 
  const [citySuggestions, setCitySuggestions] = useState([]);
  const [panelSuggestions, setPanelSuggestions] = useState([]);
  const [cityLoading, setCityLoading] = useState(false);
  const [panelLoading, setPanelLoading] = useState(false);

  const [useDefaultPanel, setUseDefaultPanel] = useState(true);
  const defaultPanels = [
    { id: "Mono-Default-400", label: "Mono 400W" },
    { id: "Poly-Default-340", label: "Poly 340W" },
    { id: "Thin-Default-150", label: "Thin 150W" },
  ];

  useEffect(() => {
    if (mode !== "city") return;
    const q = query.trim();

    if (q.length < 2) {
      setCitySuggestions([]);
      setCityLoading(false);
      return;
    }


    setCityLoading(true);

    const ac = new AbortController();
    const t = setTimeout(() => {
      const url = `/api/api/location/autocomplete?q=${encodeURIComponent(q)}`;
      fetch(url, { signal: ac.signal })
        .then((res) => {
          if (!res.ok) throw new Error(`HTTP ${res.status}`);
          return res.json();
        })
        .then((json) => {
          const list = Array.isArray(json)
            ? json
                .map((it) =>
                  typeof it === "string"
                    ? it
                    : it.name
                    ? `${it.name}${it.country ? `, ${it.country}` : ""}`
                    : ""
                )
                .filter(Boolean)
            : [];
          setCitySuggestions(list.slice(0, 10));
          setCityLoading(false);
        })
        .catch((e) => {
          if (!(e instanceof DOMException && e.name === "AbortError")) {
            setCitySuggestions([]);
            setCityLoading(false);
          }
        });
    }, 2000);

    return () => {
      clearTimeout(t);
      ac.abort();
    };
  }, [mode, query]);

  useEffect(() => {
    if (useDefaultPanel) {
      setPanelSuggestions([]);
      setPanelLoading(false);
      return; 
    }
    const q = panel.trim();
    if (q.length < 1) {
      setPanelSuggestions([]);
      setPanelLoading(false);
      return;
    }

    setPanelLoading(true);

    const ac = new AbortController();
    const t = setTimeout(() => {
      const url = `/api/api/solar-panels/search/${encodeURIComponent(q)}`;
      fetch(url, { signal: ac.signal })
        .then((res) => {
          if (!res.ok) throw new Error(`HTTP ${res.status}`);
          return res.json();
        })
        .then((json) => {
          const list = Array.isArray(json) ? json : [];
          setPanelSuggestions(list.slice(0, 10));
          setPanelLoading(false);
        })
        .catch((e) => {
          if (!(e instanceof DOMException && e.name === "AbortError")) {
            setPanelSuggestions([]);
            setPanelLoading(false);
          }
        });
    }, 2000);

    return () => {
      clearTimeout(t);
      ac.abort();
    };
  }, [useDefaultPanel, panel]);

  function handleSubmit(e) {
    e.preventDefault();

    if (!panel.trim()) {
      setError("Please select a solar panel.");
      return;
    }

    if (mode === "city") {
      if (!query.trim()) {
        setError("Please enter a location.");
        return;
      }
      setError("");
      onSubmit?.({ mode, city: query, panel });
    } else {
      const latNum = parseFloat(lat);
      const lonNum = parseFloat(lon);
      if (Number.isNaN(latNum) || Number.isNaN(lonNum)) {
        setError("Please provide valid latitude and longitude.");
        return;
      }
      setError("");
      onSubmit?.({ mode, lat: latNum, lon: lonNum, panel });
    }
  }

  function clearAll() {
    setQuery("");
    setLat("");
    setLon("");
    setError("");
    setCitySuggestions([]);
    setPanelSuggestions([]);
    setCityLoading(false);
    setPanelLoading(false);
  }

  return (
    <div
      className="min-h-screen bg-gray-900 text-gray-100 flex items-center"
      style={{
        ["--solar-sun"]: "#ff9b00",
        ["--solar-panel"]: "#0F172A",
        ["--solar-beam"]: "#ffb84d",
        ["--solar-accent"]: "#60A5FA",
        ["--sky-blue"]: "#1E3A8A",
      }}
    >
      <main className="w-full">
        <section className="container mx-auto px-4 sm:px-6 lg:px-8 py-16 sm:py-24">
          <header className="max-w-3xl mx-auto text-center mb-10 sm:mb-12">
            <h1 className="text-3xl sm:text-4xl md:text-5xl font-extrabold tracking-tight">
              Estimate your solar potential
            </h1>
            <p className="mt-3 sm:mt-4 text-base sm:text-lg text-gray-300">
              Calculate expected solar panel output for your location.
            </p>
          </header>

          <div className="max-w-4xl mx-auto mb-10 sm:mb-12" aria-hidden="true">
            <SolarHeroIllustration />
          </div>

          <form
            role="search"
            aria-label="Solar output location and panel selection"
            onSubmit={handleSubmit}
            className="max-w-2xl mx-auto space-y-6"
          >
            {/* Panel select / autocomplete */}
            <div className="space-y-3">
              <div className="flex items-center justify-between">
                <label className="text-sm text-gray-300">Panel</label>
                <button
                  type="button"
                  onClick={() => setUseDefaultPanel((v) => !v)}
                  className="text-xs px-3 py-1 rounded-xl border border-gray-700 bg-gray-800 hover:bg-gray-700"
                  title={useDefaultPanel ? "Switch to custom panel input" : "Switch to defaults"}
                >
                  {useDefaultPanel ? "Use custom" : "Use defaults"}
                </button>
              </div>

              {useDefaultPanel ? (
                <div className="grid grid-cols-1 sm:grid-cols-3 gap-2">
                  {defaultPanels.map((dp) => {
                    const active = panel === dp.id;
                    return (
                      <button
                        key={dp.id}
                        type="button"
                        onClick={() => setPanel(dp.id)}
                        className={`rounded-2xl px-4 py-3 text-sm font-medium border transition ${
                          active
                            ? "border-amber-500 bg-amber-500/10 text-amber-300"
                            : "border-gray-700 bg-gray-800 hover:bg-gray-700 text-gray-200"
                        }`}
                      >
                        {dp.label}
                      </button>
                    );
                  })}
                </div>
              ) : (
                <AutocompleteBox
                  value={panel}
                  onChange={setPanel}
                  onSelect={setPanel}
                  placeholder="Search or enter your solar panel model…"
                  suggestions={panelSuggestions}
                  loading={panelLoading}
                  maxItems={8}
                />
              )}
            </div>

            {/* Mode toggle */}
            <div className="inline-flex rounded-2xl overflow-hidden border border-gray-700">
              <button
                type="button"
                className={`px-4 py-2 text-sm font-medium ${
                  mode === "city"
                    ? "bg-[var(--solar-accent)] text-white"
                    : "bg-gray-800 text-gray-200 hover:bg-gray-700"
                }`}
                onClick={() => setMode("city")}
              >
                City
              </button>
              <button
                type="button"
                className={`px-4 py-2 text-sm font-medium ${
                  mode === "coords"
                    ? "bg-[var(--solar-accent)] text-white"
                    : "bg-gray-800 text-gray-200 hover:bg-gray-700"
                }`}
                onClick={() => setMode("coords")}
              >
                Lat / Lon
              </button>
            </div>

            {/* Location inputs */}
            <div className="flex flex-col sm:flex-row gap-3 items-stretch">
              {mode === "city" ? (
                <div className="flex-1">
                  <AutocompleteBox
                    value={query}
                    onChange={setQuery}
                    onSelect={setQuery}
                    placeholder="Search your location…"
                    suggestions={citySuggestions}
                    loading={cityLoading}
                    maxItems={8}
                  />
                </div>
              ) : (
                <>
                  <div className="flex-1">
                    <label htmlFor="lat" className="sr-only">Latitude</label>
                    <input
                      id="lat"
                      name="lat"
                      type="text"
                      inputMode="decimal"
                      placeholder="Latitude (e.g., 51.5074)"
                      value={lat}
                      onChange={(e) => setLat(e.target.value)}
                      className="w-full rounded-2xl border border-gray-600 bg-gray-800 px-4 py-3 text-base shadow-sm placeholder-gray-400 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[var(--solar-accent)] focus-visible:ring-offset-2 focus-visible:ring-offset-gray-900"
                    />
                  </div>
                  <div className="flex-1">
                    <label htmlFor="lon" className="sr-only">Longitude</label>
                    <input
                      id="lon"
                      name="lon"
                      type="text"
                      inputMode="decimal"
                      placeholder="Longitude (e.g., -0.1278)"
                      value={lon}
                      onChange={(e) => setLon(e.target.value)}
                      className="w-full rounded-2xl border border-gray-600 bg-gray-800 px-4 py-3 text-base shadow-sm placeholder-gray-400 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[var(--solar-accent)] focus-visible:ring-offset-2 focus-visible:ring-offset-gray-900"
                    />
                  </div>
                </>
              )}

              <div className="flex gap-3">
                <button
                  type="submit"
                  className="inline-flex items-center justify-center rounded-2xl px-5 py-3 text-base font-semibold bg-[var(--solar-accent)] text-white shadow-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[var(--solar-accent)] focus-visible:ring-offset-2 focus-visible:ring-offset-gray-900 hover:brightness-105 active:brightness-95"
                >
                  Calculate
                </button>
                <button
                  type="button"
                  onClick={clearAll}
                  className="inline-flex items-center justify-center rounded-2xl px-5 py-3 text-base font-semibold border border-gray-600 text-gray-100 bg-gray-800 hover:bg-gray-700 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[var(--solar-accent)] focus-visible:ring-offset-2 focus-visible:ring-offset-gray-900"
                >
                  Clear
                </button>
              </div>
            </div>

            <div
              id="form-error"
              role="alert"
              aria-live="polite"
              className="min-h-[1.5rem] mt-1 text-sm font-medium text-red-400"
            >
              {error}
            </div>
          </form>
        </section>
      </main>
    </div>
  );
}

function AutocompleteBox({
  value,
  onChange,
  onSelect,
  placeholder = "Type to search…",
  suggestions = [],
  loading = false,
  maxItems = 8,
}) {
  const [open, setOpen] = useState(false);
  const [activeIndex, setActiveIndex] = useState(-1);
  const wrapRef = useRef(null);
  const inputRef = useRef(null);
  const listId = useId();

  const filtered = useMemo(() => {
    const q = value.trim().toLowerCase();
    if (!q) return suggestions.slice(0, maxItems);
    return suggestions.filter((s) => s.toLowerCase().includes(q)).slice(0, maxItems);
  }, [value, suggestions, maxItems]);

  useEffect(() => {
    function onDocClick(e) {
      if (!wrapRef.current) return;
      if (!wrapRef.current.contains(e.target)) setOpen(false);
    }
    document.addEventListener("mousedown", onDocClick);
    return () => document.removeEventListener("mousedown", onDocClick);
  }, []);

  function commitSelect(item) {
    onSelect?.(item);
    setOpen(false);
    setActiveIndex(-1);
    inputRef.current?.focus();
  }

  function onInput(e) {
    onChange?.(e.target.value);
    setOpen(true);
    setActiveIndex(-1);
  }

  function onKeyDown(e) {
    if (!open && (e.key === "ArrowDown" || e.key === "ArrowUp")) {
      setOpen(true);
      return;
    }
    if (e.key === "Escape") {
      setOpen(false);
      return;
    }
    if (e.key === "ArrowDown") {
      e.preventDefault();
      setActiveIndex((i) => (filtered.length ? (i + 1) % filtered.length : -1));
    } else if (e.key === "ArrowUp") {
      e.preventDefault();
      setActiveIndex((i) => (filtered.length ? (i - 1 + filtered.length) % filtered.length : -1));
    } else if (e.key === "Enter") {
      if (open && activeIndex >= 0 && filtered[activeIndex]) {
        e.preventDefault();
        commitSelect(filtered[activeIndex]);
      }
    }
  }

  return (
    <div ref={wrapRef} className="relative" role="combobox" aria-expanded={open} aria-owns={listId}>
      <div className="relative">
        <svg className="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 h-5 w-5 text-gray-400" viewBox="0 0 24 24" fill="none" stroke="currentColor">
          <circle cx="11" cy="11" r="7" strokeWidth="2" />
          <line x1="21" y1="21" x2="16.65" y2="16.65" strokeWidth="2" />
        </svg>

        <input
          ref={inputRef}
          value={value}
          onChange={onInput}
          onKeyDown={onKeyDown}
          onFocus={() => setOpen(true)}
          placeholder={placeholder}
          aria-autocomplete="list"
          aria-controls={listId}
          aria-activedescendant={activeIndex >= 0 ? `${listId}-opt-${activeIndex}` : undefined}
          className="w-full rounded-2xl border border-gray-600 bg-gray-800 pl-10 pr-10 py-3 text-base shadow-sm placeholder-gray-400 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[var(--solar-accent)] focus-visible:ring-offset-2 focus-visible:ring-offset-gray-900"
        />

        {/* Clear button */}
        {value && (
          <button
            type="button"
            onClick={() => {
              onChange?.("");
              setActiveIndex(-1);
              setOpen(true);
              inputRef.current?.focus();
            }}
            aria-label="Clear"
            className="absolute right-2 top-1/2 -translate-y-1/2 rounded-md p-1 text-gray-400 hover:text-gray-200 hover:bg-gray-700"
          >
            <svg viewBox="0 0 24 24" className="h-4 w-4" fill="none" stroke="currentColor" strokeWidth="2">
              <line x1="18" y1="6" x2="6" y2="18" />
              <line x1="6" y1="6" x2="18" y2="18" />
            </svg>
          </button>
        )}
      </div>

      {/* Dropdown */}
      {open && (
        <div id={listId} role="listbox" className="absolute z-20 mt-2 w-full rounded-2xl border border-gray-700 bg-gray-800/95 shadow-xl backdrop-blur-sm overflow-hidden">
          {loading ? (
            <div className="flex items-center gap-2 px-3 py-3 text-sm text-gray-300">
              <svg className="animate-spin h-4 w-4 text-amber-300" viewBox="0 0 24 24" fill="none">
                <circle className="opacity-20" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
                <path className="opacity-90" fill="currentColor" d="M4 12a8 8 0 0 1 8-8v4A4 4 0 0 0 8 12H4z" />
              </svg>
              Searching…
            </div>
          ) : filtered.length === 0 ? (
            <div className="px-3 py-3 text-sm text-gray-400">No results</div>
          ) : (
            filtered.map((item, idx) => {
              const active = idx === activeIndex;
              return (
                <button
                  key={item + idx}
                  id={`${listId}-opt-${idx}`}
                  role="option"
                  aria-selected={active}
                  onMouseEnter={() => setActiveIndex(idx)}
                  onMouseDown={(e) => e.preventDefault()}
                  onClick={() => commitSelect(item)}
                  className={`w-full text-left px-3 py-2 text-sm transition ${
                    active ? "bg-amber-500/15 text-amber-200" : "hover:bg-gray-700 text-gray-200"
                  }`}
                >
                  <Highlighted text={item} query={value} />
                </button>
              );
            })
          )}
        </div>
      )}
    </div>
  );
}

function Highlighted({ text, query }) {
  if (!query) return <span>{text}</span>;
  const q = query.trim();
  const i = text.toLowerCase().indexOf(q.toLowerCase());
  if (i === -1) return <span>{text}</span>;
  const before = text.slice(0, i);
  const match = text.slice(i, i + q.length);
  const after = text.slice(i + q.length);
  return (
    <span>
      {before}
      <span className="text-amber-300">{match}</span>
      {after}
    </span>
  );
}

function SolarHeroIllustration() {
  return (
    <svg className="w-full h-auto" viewBox="0 0 960 360" role="img" aria-label="Sun charging a pulsating side-profile solar panel">
      {/* ... unchanged illustration ... */}
      <defs>
        <linearGradient id="panelGradient" x1="0" x2="0" y1="0" y2="1">
          <stop offset="0%" stopColor="rgba(255,180,100,0.2)" />
          <stop offset="100%" stopColor="rgba(255,140,0,0.05)" />
        </linearGradient>
        <clipPath id="panelClip">
          <polygon points="480,250 740,190 755,215 495,275" />
        </clipPath>
        <linearGradient id="skyGradient" x1="0" x2="0" y1="0" y2="1">
          <stop offset="0%" stopColor="var(--sky-blue)" />
          <stop offset="100%" stopColor="transparent" />
        </linearGradient>
        <filter id="panelGlow" x="-20%" y="-20%" width="140%" height="140%">
          <feGaussianBlur in="SourceGraphic" stdDeviation="6" result="blur" />
          <feMerge>
            <feMergeNode in="blur" />
            <feMergeNode in="SourceGraphic" />
          </feMerge>
        </filter>
      </defs>
      <style>{`
        .sun { fill: #ff9b00; }
        .ray { stroke: #ffb84d; stroke-width: 6; stroke-linecap: round; opacity: 0.9; }
        .raySpin { transform-origin: 180px 100px; animation: raySpin 12s linear infinite; }
        .rayPulse { animation: rayPulse 1.4s ease-in-out infinite alternate; }
        .panel-pulse { animation: panelPulse 2.5s ease-in-out infinite; filter: url(#panelGlow); }
        .stand { stroke: var(--solar-panel); stroke-width: 8; stroke-linecap: round; }
        @keyframes raySpin { 0% { transform: rotate(0deg); } 100% { transform: rotate(360deg); } }
        @keyframes rayPulse { 0% { stroke-width: 6; opacity: 0.7; } 100% { stroke-width: 8; opacity: 1; } }
        @keyframes panelPulse { 0% { opacity: 0.6; } 50% { opacity: 1; } 100% { opacity: 0.6; } }
        @media (prefers-reduced-motion: reduce) { .raySpin, .rayPulse, .panel-pulse { animation: none !important; } }
      `}</style>
      <rect x="0" y="0" width="960" height="290" fill="url(#skyGradient)" />
      <g>
        <circle className="sun" cx="180" cy="100" r="48" />
        <g className="raySpin">
          {Array.from({ length: 16 }).map((_, i) => {
            const angle = (i * 360) / 16;
            const rad = (angle * Math.PI) / 180;
            const x1 = 180 + Math.cos(rad) * 58;
            const y1 = 100 + Math.sin(rad) * 58;
            const x2 = 180 + Math.cos(rad) * 92;
            const y2 = 100 + Math.sin(rad) * 92;
            return <line key={i} className="ray rayPulse" x1={x1} y1={y1} x2={x2} y2={y2} />;
          })}
        </g>
      </g>
      <rect x="0" y="290" width="960" height="70" fill="#374151" />
      <g>
        <polygon className="panel-body" points="480,250 740,190 755,215 495,275" opacity="0.98" fill="#0F172A" />
        <polygon className="panel-frame" points="480,250 740,190 755,215 495,275" stroke="#4B5563" strokeWidth="2" fill="none" />
        <g clipPath="url(#panelClip)">
          {Array.from({ length: 6 }).map((_, i) => (
            <line key={`v-${i}`} x1={500 + i * 40} y1={215} x2={540 + i * 40} y2={280} stroke="rgba(255,255,255,0.15)" strokeWidth="1" />
          ))}
          {Array.from({ length: 5 }).map((_, i) => (
            <line key={`h-${i}`} x1={475} y1={250 + i * 10} x2={755} y2={200 + i * 10} stroke="rgba(255,255,255,0.15)" strokeWidth="1" />
          ))}
          <polygon className="panel-glass panel-pulse" points="480,250 740,190 755,215 495,275" fill="url(#panelGradient)" opacity="0.7" />
        </g>
        <line className="stand" x1="500" y1="275" x2="500" y2="300" />
        <line className="stand" x1="740" y1="215" x2="740" y2="300" />
      </g>
    </svg>
  );
}
