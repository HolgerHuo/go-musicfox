package ui

import (
    "fmt"
    tea "github.com/anhoder/bubbletea"
    "github.com/fogleman/ease"
    "github.com/muesli/termenv"
	"go-musicfox/configs"
    "go-musicfox/constants"
    "go-musicfox/utils"
    "math"
    "strings"
    "time"
    "unicode/utf8"
)

var (
    progressRamp       []string
    progressLastWidth  float64
    progressStartColor string
    progressEndColor   string
)

type startupModel struct {
    TotalDuration  time.Duration
    loadedDuration time.Duration
    loadedPercent  float64
    loaded         bool
    quitting       bool
}

func NewStartup() (m *startupModel) {
    m = new(startupModel)
    m.TotalDuration = time.Second * 2 // 默认

    return
}

// startup func
func updateStartup(msg tea.Msg, m *NeteaseModel) (tea.Model, tea.Cmd) {
    switch msg.(type) {

    case tickStartupMsg:
        if m.loadedDuration >= m.TotalDuration {
            m.loaded = true
            m.isListeningKey = true
            return m, tea.ClearScreen
        }
        m.loadedDuration += constants.StartupTickDuration
        m.loadedPercent = float64(m.loadedDuration) / float64(m.TotalDuration)
        if configs.ConfigRegistry.StartupProgressOutBounce {
            m.loadedPercent = ease.OutBounce(m.loadedPercent)
        }
        return m, tickStartup(constants.StartupTickDuration)

    case tea.ClearScreenMsg:
        return m, tickStartup(time.Nanosecond)
    }

    return m, nil
}

// startup view
func startupView(m *NeteaseModel) string {

    if m.WindowWidth <= 0 || m.WindowHeight <= 0 {
        return ""
    }

    blankLine := 1
    tipsHeight := 1
    progressHeight := 1
    height := utils.AsciiHeight + blankLine + tipsHeight + blankLine + progressHeight
    var top, bottom int
    if m.WindowHeight - height > 0 {
        top = (m.WindowHeight - height) / 2
    }
    if m.WindowHeight - top - height > 0 {
        bottom = m.WindowHeight - top - height
    }

    var uiBuilder strings.Builder
    if top > 1 {
        uiBuilder.WriteString(strings.Repeat("\n", top-1))
    }
    uiBuilder.WriteString(logoView(m))
    uiBuilder.WriteString("\n")
    if top != 0 && bottom != 0 {
        uiBuilder.WriteString("\n")
    }
    uiBuilder.WriteString(tipsView(m))
    uiBuilder.WriteString("\n")
    if top != 0 && bottom != 0 {
        uiBuilder.WriteString("\n")
    }
    uiBuilder.WriteString(progressView(m))
    uiBuilder.WriteString(strings.Repeat("\n", bottom))

    return uiBuilder.String()
}

// get logo
func logoView(m *NeteaseModel) string {
    if m.WindowWidth <= 0 || m.WindowHeight <= 0 {
        return ""
    }

    originLogo := utils.GetAlphaAscii(configs.ConfigRegistry.StartupWelcome)
    var logoWidth int
    if logoArr := strings.Split(originLogo, "\n"); len(logoArr) > 1 {
        logoWidth = utf8.RuneCountInString(logoArr[1])
    }

    var left int
    if m.WindowWidth - logoWidth > 0 {
        left = (m.WindowWidth - logoWidth) / 2
    }

    var logoBuilder strings.Builder
    leftSpace := strings.Repeat(" ", left)
    lines := strings.Split(originLogo, "\n")
    for _, line := range lines {
        logoBuilder.WriteString(leftSpace)
        logoBuilder.WriteString(line)
        logoBuilder.WriteString("\n")
    }
    return SetFgStyle(logoBuilder.String(), GetPrimaryColor())
}

// get tips
func tipsView(m *NeteaseModel) string {
    example := "Enter after 11.1 seconds..."
    var left int
    if m.WindowWidth - len(example) > 0 {
        left = (m.WindowWidth - len(example)) / 2
    }
    tips := fmt.Sprintf("%sEnter after %.1f seconds...",
        strings.Repeat(" ", left),
        float64(m.TotalDuration - m.loadedDuration)/float64(time.Second))

    return SetFgStyle(tips, termenv.ANSIBrightBlack)
}

// get progress
func progressView(m *NeteaseModel) string {
    width := float64(m.WindowWidth)

    if progressStartColor == "" || progressEndColor == "" {
        progressStartColor, progressEndColor = GetRandomRgbColor(true)
    }
    if width != progressLastWidth {
        progressRamp = makeRamp(progressStartColor, progressEndColor, width)
        progressLastWidth = width
    }

    fullSize := int(math.Round(width*m.loadedPercent))
    var fullCells string
    for i := 0; i < fullSize && i < len(progressRamp); i++ {
        fullCells += termenv.String(string(configs.ConfigRegistry.ProgressFullChar)).Foreground(termProfile.Color(progressRamp[i])).String()
    }

    emptySize := 0
    if int(width) - fullSize > 0 {
        emptySize = int(width) - fullSize
    }
    emptyCells := strings.Repeat(string(configs.ConfigRegistry.ProgressEmptyChar), emptySize)

    return fmt.Sprintf("%s%s", fullCells, emptyCells)
}
