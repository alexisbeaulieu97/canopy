package tui

import (
	"github.com/alexisbeaulieu97/canopy/internal/tui/components"
)

// Action constants for confirmation dialogs - aliased from components.
const (
	actionClose = string(components.ActionClose)
	actionPush  = string(components.ActionPush)
)

// Status indicator styles - aliased from components
var (
	statusCleanStyle = components.StatusCleanStyle
	statusDirtyStyle = components.StatusDirtyStyle
	statusWarnStyle  = components.StatusWarnStyle
)

// Text styles - aliased from components
var (
	subtleTextStyle = components.SubtleTextStyle
	mutedTextStyle  = components.MutedTextStyle
	boldTextStyle   = components.BoldTextStyle
	accentTextStyle = components.AccentTextStyle
)

// Badge styles - aliased from components
var (
	badgeWarnStyle = components.BadgeWarnStyle
	badgeInfoStyle = components.BadgeInfoStyle
)

// Layout styles - aliased from components
var (
	titleStyle        = components.TitleStyle
	detailHeaderStyle = components.DetailHeaderStyle
	detailLabelStyle  = components.DetailLabelStyle
	detailValueStyle  = components.DetailValueStyle
)

// Interactive element styles - aliased from components
var (
	helpTextStyle = components.HelpTextStyle
)
