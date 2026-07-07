package tui

import (
	"context"
	"fmt"

	"redisshow/internal/client"

	"github.com/gdamore/tcell/v2"
	"github.com/redis/go-redis/v9"
	"github.com/rivo/tview"
)

func Run(ctx context.Context, rdb *redis.Client, cfg client.Config) error {
	app := tview.NewApplication()
	currentRdb := rdb
	currentCfg := cfg

	var loadedKeys []string

	keyList := tview.NewList().
		ShowSecondaryText(false).
		SetHighlightFullLine(true).
		SetSelectedStyle(tcell.StyleDefault.Background(tcell.ColorDarkCyan).Foreground(tcell.ColorWhite))

	detailView := tview.NewTextView().
		SetDynamicColors(true).
		SetWrap(true).
		SetScrollable(true)

	patternInput := tview.NewInputField().
		SetLabel("匹配: ").
		SetFieldWidth(30).
		SetText("*")

	statusBar := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft)

	header := tview.NewTextView().SetDynamicColors(true)

	updateHeader := func() {
		header.SetText(fmt.Sprintf(
			"[::b]redisshow[::-]  %s  db=%d  |  [yellow]r[-]刷新  [yellow]/[-]搜索  [yellow]d[-]删除  [yellow]e[-]编辑  [yellow]b[-]切库  [yellow]q[-]退出",
			currentCfg.Addr, currentCfg.DB,
		))
	}
	updateHeader()

	setStatus := func(msg string) {
		statusBar.SetText(msg)
	}

	showKeyDetail := func(key string) {
		if key == "" {
			detailView.SetText("")
			return
		}
		text, err := formatKeyDetail(ctx, currentRdb, key)
		if err != nil {
			detailView.SetText(fmt.Sprintf("[red]读取失败: %v", err))
			return
		}
		detailView.SetText(text)
	}

	refreshKeys := func() {
		pattern := patternInput.GetText()
		if pattern == "" {
			pattern = "*"
			patternInput.SetText(pattern)
		}

		setStatus(fmt.Sprintf("正在加载键列表: %s ...", pattern))
		app.Draw()

		keys, err := scanKeys(ctx, currentRdb, pattern)
		if err != nil {
			setStatus(fmt.Sprintf("[red]加载失败: %v", err))
			return
		}

		current := keyList.GetCurrentItem()
		loadedKeys = keys
		keyList.Clear()
		for _, key := range keys {
			keyList.AddItem(highlightKeyName(key, pattern), "", 0, nil)
		}

		if len(keys) == 0 {
			detailView.SetText("[gray]未找到匹配的键")
			setStatus(fmt.Sprintf("共 0 个键 (pattern: %s)", pattern))
			return
		}

		if current >= 0 && current < len(keys) {
			keyList.SetCurrentItem(current)
		} else if current >= len(keys) {
			keyList.SetCurrentItem(len(keys) - 1)
		}
		showKeyDetail(loadedKeys[keyList.GetCurrentItem()])
		setStatus(fmt.Sprintf("共 %d 个键 (pattern: %s)", len(keys), pattern))
	}

	keyList.SetChangedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
		if index >= 0 && index < len(loadedKeys) {
			showKeyDetail(loadedKeys[index])
		}
	})

	patternInput.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			refreshKeys()
			app.SetFocus(keyList)
		}
	})

	leftTitle := tview.NewTextView().SetText(" 键列表 ").SetTextAlign(tview.AlignCenter)
	rightTitle := tview.NewTextView().SetText(" 键详情 ").SetTextAlign(tview.AlignCenter)

	leftPanel := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(leftTitle, 1, 0, false).
		AddItem(keyList, 0, 1, true)

	rightPanel := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(rightTitle, 1, 0, false).
		AddItem(detailView, 0, 1, false)

	mainFlex := tview.NewFlex().
		AddItem(leftPanel, 0, 1, true).
		AddItem(rightPanel, 0, 2, false)

	layout := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(header, 1, 0, false).
		AddItem(mainFlex, 0, 1, true).
		AddItem(patternInput, 1, 0, false).
		AddItem(statusBar, 1, 0, false)

	actions := &appActions{
		ctx:          ctx,
		app:          app,
		layout:       layout,
		cfg:          &currentCfg,
		rdb:          &currentRdb,
		keyList:      keyList,
		loadedKeys:   &loadedKeys,
		refresh:      refreshKeys,
		setStatus:    setStatus,
		updateHeader: updateHeader,
		showDetail:   showKeyDetail,
	}

	layout.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 'q':
			app.Stop()
			return nil
		case 'r':
			refreshKeys()
			return nil
		case '/':
			app.SetFocus(patternInput)
			return nil
		case 'd':
			actions.confirmDelete()
			return nil
		case 'e':
			actions.editValue()
			return nil
		case 'b':
			actions.switchDB()
			return nil
		}
		if event.Key() == tcell.KeyCtrlC {
			app.Stop()
			return nil
		}
		return event
	})

	app.SetRoot(layout, true).SetFocus(keyList)
	refreshKeys()

	return app.Run()
}
