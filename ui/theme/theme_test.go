package theme

import "testing"

func TestLightAndDarkThemes(t *testing.T) {
	light := LightTheme()
	dark := DarkTheme()

	if light.IsDark {
		t.Errorf("LightTheme().IsDark = true, want false")
	}
	if !dark.IsDark {
		t.Errorf("DarkTheme().IsDark = false, want true")
	}
	if light.Colors.Primary == "" || dark.Colors.Primary == "" {
		t.Errorf("themes must define a primary colour")
	}
	if light.Colors.Background == dark.Colors.Background {
		t.Errorf("light and dark backgrounds should differ")
	}
	if light.CornerRadius <= 0 || light.FontFamily == "" {
		t.Errorf("themes must define corner radius and font family")
	}
}
