import React, { useMemo, useState } from "react";
import {
  AreaChart,
  Area,
  Line,
  XAxis,
  YAxis,
  Tooltip,
  CartesianGrid,
  ResponsiveContainer,
  ReferenceArea,
} from "recharts";

/* ---------- Dummy data ---------- */
function buildDummyDay() {
  const tz = "Europe/London";
  const date = "2025-08-11";
  const pts = Array.from({ length: 24 }).map((_, i) => {
    // daylight curve centered near 13:00
    const t = (i - 13) / 6;
    const daylight = Math.max(0, 1 - t * t);
    const ghi = Math.round(820 * daylight);     // W/m²
    const energy = Math.max(0, 360 * daylight); // Wh per hour (dummy)
    const low = energy * 0.9;
    const high = energy * 1.1;

    const iso = `${date}T${String(i).padStart(2, "0")}:00:00Z`;
    return {
      time: iso,
      ghi,
      energyWh: energy,
      energyWhLow: low,
      energyWhHigh: high,
    };
  });

  // cumulative totals
  let c = 0, cl = 0, ch = 0;
  pts.forEach((p) => {
    c += p.energyWh;
    cl += p.energyWhLow;
    ch += p.energyWhHigh;
    p.cumulativeWh = c;
    p.cumulativeLow = cl;
    p.cumulativeHigh = ch;
  });

  return { timezone: tz, date, points: pts, totalWh: c };
}

/* ---------- Helpers ---------- */
const nf = new Intl.NumberFormat(undefined, { maximumFractionDigits: 1 });
function kwh(wh) {
  const v = wh / 1000;
  return `${v < 10 ? v.toFixed(2) : v.toFixed(1)} kWh`;
}
function timeLabel(iso, tz) {
  return new Intl.DateTimeFormat(undefined, {
    hour: "2-digit",
    minute: "2-digit",
    timeZone: tz,
  }).format(new Date(iso));
}

/* ---------- Dark custom tooltips ---------- */
function EnergyTooltip({ active, payload, label }) {
  if (!active || !payload || !payload.length) return null;
  const byKey = Object.fromEntries(payload.map(p => [p.dataKey, p.value]));
  return (
    <div className="bg-gray-900/95 border border-amber-500/30 rounded-lg p-3 text-sm text-gray-100 shadow-lg">
      <div className="font-semibold text-orange-400 mb-1">{label}</div>
      {"low" in byKey && (
        <Row k="Lower Estimate" v={`${nf.format(byKey.low)} Wh`} tone="low" />
      )}
      {"base" in byKey && (
        <Row k="Base Estimate" v={`${nf.format(byKey.base)} Wh`} tone="base" />
      )}
      {"high" in byKey && (
        <Row k="Upper Estimate" v={`${nf.format(byKey.high)} Wh`} tone="high" />
      )}
    </div>
  );
}
function GHITooltip({ active, payload, label }) {
  if (!active || !payload || !payload.length) return null;
  const v = payload[0].value;
  return (
    <div className="bg-gray-900/95 border border-amber-500/30 rounded-lg p-3 text-sm text-gray-100 shadow-lg">
      <div className="font-semibold text-orange-400 mb-1">{label}</div>
      <Row k="GHI" v={`${nf.format(v)} W/m²`} tone="base" />
    </div>
  );
}
function Row({ k, v, tone }) {
  const toneClass =
    tone === "low" ? "text-amber-200" :
    tone === "high" ? "text-amber-300" :
    "text-orange-300";
  return (
    <div className="flex justify-between gap-6">
      <span className="text-gray-300">{k}:</span>
      <span className={toneClass}>{v}</span>
    </div>
  );
}

/* ---------- Band label element ---------- */
function BandLabel({ viewBox }) {
  const { x, y, width, height } = viewBox || {};
  if (x == null) return null;
  return (
    <text
      x={x + width / 2}
      y={y + height / 2}
      textAnchor="middle"
      fill="#FBBF24"
      opacity={0.85}
      fontSize={12}
    >
      Estimate Range
    </text>
  );
}

/* ---------- UI bits ---------- */
function Card({ children }) {
  return (
    <div className="rounded-2xl border border-gray-800 bg-gray-800/60 p-4 shadow-[0_1px_0_0_rgba(255,255,255,0.03)_inset]">
      {children}
    </div>
  );
}
function Segmented({ options, value, onChange }) {
  return (
    <div className="inline-flex items-center rounded-full border border-gray-700 bg-gray-900 p-1">
      {options.map((opt) => {
        const active = value === opt;
        return (
          <button
            key={opt}
            onClick={() => onChange(opt)}
            className={`px-3 py-1 text-xs rounded-full transition ${
              active
                ? "bg-orange-500 text-black shadow"
                : "text-gray-300 hover:bg-gray-800"
            }`}
          >
            {opt}
          </button>
        );
      })}
    </div>
  );
}
function PillToggle({ checked, onChange, labelOn = "Range: On", labelOff = "Range: Off" }) {
  return (
    <button
      onClick={() => onChange(!checked)}
      className={`relative inline-flex items-center text-xs rounded-full border border-gray-700 transition px-2 py-1 ${
        checked ? "bg-amber-500/20 text-amber-300" : "bg-gray-900 text-gray-300"
      }`}
      title="Toggle range band"
    >
      <span
        className={`mr-2 inline-block h-2 w-2 rounded-full ${
          checked ? "bg-amber-400" : "bg-gray-500"
        }`}
      />
      {checked ? labelOn : labelOff}
    </button>
  );
}

/* ---------- Page ---------- */
export default function Estimate() {
  const data = useMemo(() => buildDummyDay(), []);
  const [cumulative, setCumulative] = useState(true);
  const [showRange, setShowRange] = useState(true);

  const rowsEnergy = useMemo(() => {
    const tz = data.timezone;
    return data.points.map((p) => ({
      label: timeLabel(p.time, tz),
      base: cumulative ? p.cumulativeWh : p.energyWh,
      low: cumulative ? p.cumulativeLow : p.energyWhLow,
      high: cumulative ? p.cumulativeHigh : p.energyWhHigh,
    }));
  }, [data, cumulative]);

  const rowsGHI = useMemo(() => {
    const tz = data.timezone;
    return data.points.map((p) => ({
      label: timeLabel(p.time, tz),
      ghi: p.ghi,
    }));
  }, [data]);

  const peakPoint = useMemo(() => {
    let max = -1, best = null;
    for (const p of data.points) {
      if (p.energyWh > max) { max = p.energyWh; best = p; }
    }
    return best;
  }, [data]);

  // y-bounds for band label box
  const bandY1 = useMemo(
    () => Math.min(...rowsEnergy.map((d) => d.low ?? d.base)),
    [rowsEnergy]
  );
  const bandY2 = useMemo(
    () => Math.max(...rowsEnergy.map((d) => d.high ?? d.base)),
    [rowsEnergy]
  );

  return (
    <div className="min-h-screen bg-gray-900 text-gray-100">
      <div className="max-w-6xl mx-auto px-4 py-8 space-y-6">
        {/* Header */}
        <div className="flex items-end justify-between">
          <div>
            <h1 className="text-2xl font-semibold tracking-tight">Today’s estimate</h1>
            <div className="text-sm text-gray-300 mt-1">
              <span>Date:</span> <span className="font-medium">{data.date}</span>
              <span className="mx-2 opacity-40">|</span>
              <span>Timezone:</span> <span className="font-medium">{data.timezone}</span>
            </div>
          </div>
        </div>

        {/* Stat cards */}
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
          <Card>
            <div className="text-xs text-gray-400">Total generated</div>
            <div className="mt-1 text-3xl font-bold text-orange-400">{kwh(data.totalWh)}</div>
          </Card>
          <Card>
            <div className="text-xs text-gray-400">Peak hour (by output)</div>
            <div className="mt-1 text-3xl font-bold text-orange-400">
              {peakPoint ? timeLabel(peakPoint.time, data.timezone) : "—"}
            </div>
            <div className="text-sm text-gray-400">
              {peakPoint ? `${nf.format(peakPoint.energyWh)} Wh` : ""}
            </div>
          </Card>
        </div>

        {/* ENERGY CHART FIRST */}
        <section className="rounded-2xl border border-gray-800 bg-gray-800/60 p-4">
          <div className="flex items-center justify-between mb-2">
            <h2 className="text-sm font-medium text-gray-200">
              {cumulative ? "Cumulative Energy" : "Hourly Energy"}{" "}
              <span className="text-gray-400">(Wh)</span>
            </h2>
            <div className="flex items-center gap-2">
              <Segmented
                options={["Cumulative", "Hourly"]}
                value={cumulative ? "Cumulative" : "Hourly"}
                onChange={(v) => setCumulative(v === "Cumulative")}
              />
              <PillToggle
                checked={showRange}
                onChange={setShowRange}
                labelOn="Range: On"
                labelOff="Range: Off"
              />
            </div>
          </div>

          <div className="h-[340px] w-full">
            <ResponsiveContainer width="100%" height="100%">
              <AreaChart data={rowsEnergy} margin={{ left: 8, right: 8, top: 10, bottom: 10 }}>
                <defs>
                  <linearGradient id="rangeFill" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="0%" stopColor="#FBBF24" stopOpacity={0.25} />
                    <stop offset="100%" stopColor="#F59E0B" stopOpacity={0.05} />
                  </linearGradient>
                  <linearGradient id="lineFill" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="0%" stopColor="#F97316" stopOpacity={0.25} />
                    <stop offset="100%" stopColor="#EA580C" stopOpacity={0.05} />
                  </linearGradient>
                </defs>
                <CartesianGrid strokeDasharray="3 3" stroke="#4B5563" />
                <XAxis dataKey="label" stroke="#E5E7EB" />
                <YAxis unit=" Wh" stroke="#E5E7EB" />
                <Tooltip content={<EnergyTooltip />} />
                {showRange && (
                  <>
                    <Area dataKey="high" type="monotone" stroke="#FBBF24" fill="url(#rangeFill)" />
                    <Area dataKey="low"  type="monotone" stroke="#FBBF24" fill="url(#rangeFill)" />
                    <ReferenceArea
                      x1={rowsEnergy[0]?.label}
                      x2={rowsEnergy[rowsEnergy.length - 1]?.label}
                      y1={bandY1}
                      y2={bandY2}
                      label={<BandLabel />}
                      strokeOpacity={0}
                      fillOpacity={0}
                    />
                  </>
                )}
                <Area dataKey="base" type="monotone" stroke="#F97316" fill="url(#lineFill)" />
                <Line dataKey="base" type="monotone" stroke="#F97316" strokeWidth={2} dot={false} />
              </AreaChart>
            </ResponsiveContainer>
          </div>
        </section>

        {/* GHI CHART SECOND */}
        <section className="rounded-2xl border border-gray-800 bg-gray-800/60 p-4">
          <div className="flex items-center justify-between mb-2">
            <h2 className="text-sm font-medium text-gray-200">Global Horizontal Irradiance (GHI)</h2>
          </div>
          <div className="h-[300px] w-full">
            <ResponsiveContainer width="100%" height="100%">
              <AreaChart data={rowsGHI}>
                <defs>
                  <linearGradient id="ghiFill" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="0%" stopColor="#F97316" stopOpacity={0.25} />
                    <stop offset="100%" stopColor="#EA580C" stopOpacity={0.05} />
                  </linearGradient>
                </defs>
                <CartesianGrid strokeDasharray="3 3" stroke="#4B5563" />
                <XAxis dataKey="label" stroke="#E5E7EB" />
                <YAxis unit=" W/m²" stroke="#E5E7EB" />
                <Tooltip content={<GHITooltip />} />
                <Area dataKey="ghi" type="monotone" stroke="#F97316" fill="url(#ghiFill)" />
                <Line dataKey="ghi" type="monotone" stroke="#F97316" strokeWidth={2} dot={false} />
              </AreaChart>
            </ResponsiveContainer>
          </div>
        </section>
      </div>
    </div>
  );
}
