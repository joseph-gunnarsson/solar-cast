package solar

import (
	"math"
	"testing"
	"time"
)

func almostEqual(t *testing.T, got, want, tol float64) {
	t.Helper()
	if math.Abs(got-want) > tol {
		t.Fatalf("got %.6f, want %.6f (tol %.6f)", got, want, tol)
	}
}

func TestEstimatedCellTemperature_Valid(t *testing.T) {
	panel := SolarPanelData{
		NOCT_Temp:                  45,
		TemperatureCoefficientPmax: -0.003,
		MaximumPowerPmax:           400,
	}
	ambient := 25.0
	irr := 800.0

	tc, err := EstimatedCellTemperature(panel, ambient, irr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	almostEqual(t, tc, 50.0, 1e-9)
}

func TestEstimatedCellTemperature_RangeErrors(t *testing.T) {
	p := SolarPanelData{NOCT_Temp: 45}

	if _, err := EstimatedCellTemperature(p, -41, 500); err == nil {
		t.Fatal("expected error for ambient < -40")
	}
	if _, err := EstimatedCellTemperature(p, 86, 500); err == nil {
		t.Fatal("expected error for ambient > 85")
	}
	if _, err := EstimatedCellTemperature(p, 20, -1); err == nil {
		t.Fatal("expected error for irradiance < 0")
	}
	if _, err := EstimatedCellTemperature(p, 20, 1201); err == nil {
		t.Fatal("expected error for irradiance > 1200")
	}
}

func TestCalculateSolarPanelOutputByHour_Basic(t *testing.T) {
	panel := SolarPanelData{
		NOCT_Temp:                  45,
		TemperatureCoefficientPmax: -0.003,
		MaximumPowerPmax:           400.0,
	}
	ambient := 25.0
	irr := 1000.0

	got, err := CalculateSolarPanelOutputByHour(panel, ambient, irr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	almostEqual(t, got, 362.5, 1e-6)
}

func TestCalculateSolarPanelOutputByHour_ClampNegative(t *testing.T) {
	panel := SolarPanelData{
		NOCT_Temp:                  80,
		TemperatureCoefficientPmax: -0.05,
		MaximumPowerPmax:           400.0,
	}
	ambient := 85.0
	irr := 1200.0

	got, err := CalculateSolarPanelOutputByHour(panel, ambient, irr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 0 {
		t.Fatalf("expected energy 0 after clamp, got %.6f", got)
	}
}

func TestTiltBoostFactor_MidLat_NH_January(t *testing.T) {
	lat := 51.5
	ts := time.Date(2025, time.January, 15, 12, 0, 0, 0, time.UTC)
	got := TiltBoostFactor(lat, ts)
	almostEqual(t, got, boostMidLat[0], 1e-9)
}

func TestTiltBoostFactor_Equatorial_June(t *testing.T) {
	lat := 0.0
	ts := time.Date(2025, time.June, 15, 12, 0, 0, 0, time.UTC)
	got := TiltBoostFactor(lat, ts)
	almostEqual(t, got, boostEquatorial[5], 1e-9)
}

func TestTiltBoostFactor_SH_Flip_January(t *testing.T) {
	lat := -34.0
	ts := time.Date(2025, time.January, 15, 12, 0, 0, 0, time.UTC)
	got := TiltBoostFactor(lat, ts)

	eq := boostEquatorial[6]
	mid := boostMidLat[6]
	want := eq + (mid-eq)*0.95
	almostEqual(t, got, want, 1e-9)
}

func TestTiltBoostFactor_HighLat_Blend(t *testing.T) {
	lat := 60.0
	ts := time.Date(2025, time.December, 15, 12, 0, 0, 0, time.UTC)
	got := TiltBoostFactor(lat, ts)

	mid := boostMidLat[11]
	hi := boostHighLat[11]
	want := mid + (hi-mid)*(1.0/3.0)
	almostEqual(t, got, want, 1e-9)
}

func TestClamp01(t *testing.T) {
	if v := clamp01(-0.5); v != 0 {
		t.Fatalf("clamp01(-0.5) got %v, want 0", v)
	}
	almostEqual(t, clamp01(0.5), 0.5, 1e-12)
	if v := clamp01(2.0); v != 1 {
		t.Fatalf("clamp01(2) got %v, want 1", v)
	}
}

func TestValWrapsIndex(t *testing.T) {
	arr := [12]float64{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11}
	if v := val(arr, 1); v != 1 {
		t.Fatalf("val(arr,1) got %v, want 1", v)
	}
	if v := val(arr, 13); v != 1 {
		t.Fatalf("val(arr,13) got %v, want 1", v)
	}
}

func TestLerp(t *testing.T) {
	almostEqual(t, lerp(1, 3, 0), 1, 1e-12)
	almostEqual(t, lerp(1, 3, 0.5), 2, 1e-12)
	almostEqual(t, lerp(1, 3, 1), 3, 1e-12)
	almostEqual(t, lerp(1, 3, -1), 1, 1e-12)
	almostEqual(t, lerp(1, 3, 2), 3, 1e-12)
}
