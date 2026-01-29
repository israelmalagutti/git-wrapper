package colors

import (
	"strings"
	"testing"
)

func TestColorConstants(t *testing.T) {
	t.Run("reset code is correct", func(t *testing.T) {
		if Reset != "\033[0m" {
			t.Errorf("expected Reset to be \\033[0m, got %q", Reset)
		}
	})

	t.Run("standard colors use 30-37 range", func(t *testing.T) {
		colors := map[string]string{
			"Black":   Black,
			"Red":     Red,
			"Green":   Green,
			"Yellow":  Yellow,
			"Blue":    Blue,
			"Magenta": Magenta,
			"Cyan":    Cyan,
			"White":   White,
		}

		for name, code := range colors {
			if !strings.HasPrefix(code, "\033[3") {
				t.Errorf("%s should start with \\033[3, got %q", name, code)
			}
		}
	})

	t.Run("bright colors use 90-97 range", func(t *testing.T) {
		colors := map[string]string{
			"BrightBlack":   BrightBlack,
			"BrightRed":     BrightRed,
			"BrightGreen":   BrightGreen,
			"BrightYellow":  BrightYellow,
			"BrightBlue":    BrightBlue,
			"BrightMagenta": BrightMagenta,
			"BrightCyan":    BrightCyan,
			"BrightWhite":   BrightWhite,
		}

		for name, code := range colors {
			if !strings.HasPrefix(code, "\033[9") {
				t.Errorf("%s should start with \\033[9, got %q", name, code)
			}
		}
	})
}

func TestSemanticColors(t *testing.T) {
	// Force colors enabled for testing
	originalEnabled := enabled
	SetEnabled(true)
	defer SetEnabled(originalEnabled)

	t.Run("Success wraps with green", func(t *testing.T) {
		result := Success("test")
		if !strings.Contains(result, Green) {
			t.Error("Success should contain Green color code")
		}
		if !strings.HasSuffix(result, Reset) {
			t.Error("Success should end with Reset code")
		}
	})

	t.Run("Error wraps with bright red", func(t *testing.T) {
		result := Error("test")
		if !strings.Contains(result, BrightRed) {
			t.Error("Error should contain BrightRed color code")
		}
	})

	t.Run("Warning wraps with yellow", func(t *testing.T) {
		result := Warning("test")
		if !strings.Contains(result, Yellow) {
			t.Error("Warning should contain Yellow color code")
		}
	})

	t.Run("Info wraps with cyan", func(t *testing.T) {
		result := Info("test")
		if !strings.Contains(result, Cyan) {
			t.Error("Info should contain Cyan color code")
		}
	})

	t.Run("Muted wraps with bright black", func(t *testing.T) {
		result := Muted("test")
		if !strings.Contains(result, BrightBlack) {
			t.Error("Muted should contain BrightBlack color code")
		}
	})
}

func TestBranchColors(t *testing.T) {
	originalEnabled := enabled
	SetEnabled(true)
	defer SetEnabled(originalEnabled)

	t.Run("BranchCurrent is green and bold", func(t *testing.T) {
		result := BranchCurrent("main")
		if !strings.Contains(result, Bold) {
			t.Error("BranchCurrent should be bold")
		}
		if !strings.Contains(result, Green) {
			t.Error("BranchCurrent should be green")
		}
	})

	t.Run("BranchTrunk is blue and bold", func(t *testing.T) {
		result := BranchTrunk("main")
		if !strings.Contains(result, Bold) {
			t.Error("BranchTrunk should be bold")
		}
		if !strings.Contains(result, Blue) {
			t.Error("BranchTrunk should be blue")
		}
	})

	t.Run("BranchParent is cyan", func(t *testing.T) {
		result := BranchParent("develop")
		if !strings.Contains(result, Cyan) {
			t.Error("BranchParent should be cyan")
		}
	})

	t.Run("BranchChild is magenta", func(t *testing.T) {
		result := BranchChild("feature")
		if !strings.Contains(result, Magenta) {
			t.Error("BranchChild should be magenta")
		}
	})
}

func TestCyclePalette(t *testing.T) {
	t.Run("Cycle returns colors in sequence", func(t *testing.T) {
		first := Cycle(0)
		second := Cycle(1)

		if first == second {
			t.Error("Adjacent cycle colors should be different")
		}
	})

	t.Run("Cycle wraps around palette", func(t *testing.T) {
		paletteLen := len(cyclePalette)
		first := Cycle(0)
		wrapped := Cycle(paletteLen)

		if first != wrapped {
			t.Error("Cycle should wrap around to beginning")
		}
	})

	t.Run("CycleText applies color to text", func(t *testing.T) {
		originalEnabled := enabled
		SetEnabled(true)
		defer SetEnabled(originalEnabled)

		result := CycleText("test", 0)
		if !strings.Contains(result, "test") {
			t.Error("CycleText should contain the original text")
		}
		if !strings.HasSuffix(result, Reset) {
			t.Error("CycleText should end with Reset")
		}
	})
}

func TestColorDisabled(t *testing.T) {
	originalEnabled := enabled
	SetEnabled(false)
	defer SetEnabled(originalEnabled)

	t.Run("colors disabled returns plain text", func(t *testing.T) {
		result := Success("test")
		if result != "test" {
			t.Errorf("expected plain 'test', got %q", result)
		}
	})

	t.Run("Error returns plain text when disabled", func(t *testing.T) {
		result := Error("error")
		if result != "error" {
			t.Errorf("expected plain 'error', got %q", result)
		}
	})

	t.Run("CycleText returns plain text when disabled", func(t *testing.T) {
		result := CycleText("branch", 5)
		if result != "branch" {
			t.Errorf("expected plain 'branch', got %q", result)
		}
	})
}

func TestEmptyText(t *testing.T) {
	originalEnabled := enabled
	SetEnabled(true)
	defer SetEnabled(originalEnabled)

	t.Run("empty text returns empty", func(t *testing.T) {
		result := Success("")
		if result != "" {
			t.Errorf("expected empty string, got %q", result)
		}
	})
}

func TestTreeChars(t *testing.T) {
	t.Run("DefaultTreeChars has unicode characters", func(t *testing.T) {
		chars := DefaultTreeChars()
		if chars.Vertical != "│" {
			t.Errorf("expected │, got %s", chars.Vertical)
		}
		if chars.Corner != "└" {
			t.Errorf("expected └, got %s", chars.Corner)
		}
	})

	t.Run("ASCIITreeChars has ASCII characters", func(t *testing.T) {
		chars := ASCIITreeChars()
		if chars.Vertical != "|" {
			t.Errorf("expected |, got %s", chars.Vertical)
		}
		if chars.Corner != "`" {
			t.Errorf("expected `, got %s", chars.Corner)
		}
	})
}

func TestIsEnabled(t *testing.T) {
	originalEnabled := enabled

	SetEnabled(true)
	if !IsEnabled() {
		t.Error("IsEnabled should return true after SetEnabled(true)")
	}

	SetEnabled(false)
	if IsEnabled() {
		t.Error("IsEnabled should return false after SetEnabled(false)")
	}

	SetEnabled(originalEnabled)
}

func TestAdditionalFormatting(t *testing.T) {
	originalEnabled := enabled
	SetEnabled(true)
	defer SetEnabled(originalEnabled)

	if Highlight("hi") == "" {
		t.Error("Highlight should return formatted text")
	}
	if StatusApproved("ok") == "" || StatusPending("wait") == "" || StatusChangesRequested("chg") == "" || StatusDraft("draft") == "" {
		t.Error("status helpers should return formatted text")
	}
	if BoldText("b") == "" || SubduedText("s", 2) == "" || ItalicText("i") == "" {
		t.Error("text helpers should return formatted text")
	}
	if Sprintf("hi %s", "there") != "hi there" {
		t.Error("Sprintf should behave like fmt.Sprintf")
	}
	if CommitSHA("abc") == "" {
		t.Error("CommitSHA should return formatted text")
	}
}

func TestApplyHelpers(t *testing.T) {
	originalEnabled := enabled
	SetEnabled(true)
	defer SetEnabled(originalEnabled)

	if got := apply(Green, ""); got != "" {
		t.Errorf("expected empty text to stay empty, got %q", got)
	}
	if got := applyMultiple("", Bold, Green); got != "" {
		t.Errorf("expected empty text to stay empty, got %q", got)
	}

	SetEnabled(false)
	if got := apply(Green, "text"); got != "text" {
		t.Errorf("expected disabled colors to return plain text, got %q", got)
	}
	if got := applyMultiple("text", Bold, Green); got != "text" {
		t.Errorf("expected disabled colors to return plain text, got %q", got)
	}
}
